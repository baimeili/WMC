// Package photo 处理用户照片：校验、去除 EXIF/GPS、压缩、签名 URL。
package photo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"time"

	// 注册解码器（仅解码，重新编码后自动丢弃 EXIF 与 GPS 信息）
	_ "image/gif"
	_ "image/png"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/profile"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/store"
)

// 上传限制。
const (
	MaxFileSize   = 5 << 20 // 5MB
	MaxDimension  = 8000
	TargetLongest = 1280 // 长边压缩到 1280 像素
	JPEGQuality   = 85
	SignedURLTTL  = time.Hour
)

// Service 照片服务。
type Service struct {
	store *store.Store
	obj   storage.ObjectStorage
}

// NewService 构造照片服务。
func NewService(st *store.Store, obj storage.ObjectStorage) *Service {
	return &Service{store: st, obj: obj}
}

// Photo DTO。
type Photo struct {
	ID               string `json:"id"`
	URL              string `json:"url"`
	SortOrder        int16  `json:"sort_order"`
	IsPrimary        bool   `json:"is_primary"`
	ModerationStatus int16  `json:"moderation_status"`
}

// Process 解码图片、限制尺寸并重新编码为 JPEG（自动去除 EXIF/GPS）。
func Process(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, apperr.Invalid("图片内容为空")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, apperr.Invalid("不支持的图片格式")
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, apperr.Invalid("图片尺寸非法")
	}
	if w > MaxDimension || h > MaxDimension {
		return nil, apperr.Invalid("图片尺寸过大")
	}
	if w > TargetLongest || h > TargetLongest {
		scale := float64(TargetLongest) / float64(maxInt(w, h))
		newW := int(float64(w) * scale)
		newH := int(float64(h) * scale)
		if newW < 1 {
			newW = 1
		}
		if newH < 1 {
			newH = 1
		}
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
		img = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return buf.Bytes(), nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Upload 上传并保存照片。
func (s *Service) Upload(ctx context.Context, userID uuid.UUID, r io.Reader, size int64, makePrimary bool) (Photo, error) {
	if size > MaxFileSize {
		return Photo{}, apperr.Invalid("图片不能超过 5MB")
	}
	raw, err := io.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return Photo{}, apperr.Invalid("读取上传内容失败")
	}
	if int64(len(raw)) > MaxFileSize {
		return Photo{}, apperr.Invalid("图片不能超过 5MB")
	}
	processed, err := Process(raw)
	if err != nil {
		return Photo{}, err
	}

	key := fmt.Sprintf("photos/%s/%s.jpg", userID.String(), uuid.NewString())
	if err := s.obj.Put(ctx, key, processed, "image/jpeg"); err != nil {
		return Photo{}, apperr.ErrInternal.Wrap(err)
	}

	existing, err := s.store.Q.CountPhotosByUser(ctx, userID)
	if err != nil {
		return Photo{}, apperr.ErrInternal.Wrap(err)
	}
	primary := makePrimary || existing == 0

	var created sqlcdb.UserPhoto
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		if primary {
			if err := tx.Q.ClearPrimaryPhotos(ctx, userID); err != nil {
				return err
			}
		}
		row, err := tx.Q.CreatePhoto(ctx, sqlcdb.CreatePhotoParams{
			UserID: userID, ObjectKey: key, SortOrder: int16(existing), IsPrimary: primary,
			ModerationStatus: enums.ModerationPending,
		})
		if err != nil {
			return err
		}
		created = row
		return s.recomputeCompletion(ctx, tx, userID)
	})
	if err != nil {
		return Photo{}, apperr.ErrInternal.Wrap(err)
	}
	return s.toDTO(ctx, created)
}

// List 返回用户照片（含签名 URL）。
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Photo, error) {
	rows, err := s.store.Q.ListPhotosByUser(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]Photo, 0, len(rows))
	for _, row := range rows {
		dto, err := s.toDTO(ctx, row)
		if err != nil {
			continue
		}
		out = append(out, dto)
	}
	return out, nil
}

// Delete 删除照片（软删除）；如果是主图则顺延下一张。
func (s *Service) Delete(ctx context.Context, userID, photoID uuid.UUID) error {
	row, err := s.store.Q.GetPhotoByID(ctx, photoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound.WithMessage("照片不存在")
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if row.UserID != userID {
		return apperr.ErrForbidden.WithMessage("无权删除该照片")
	}
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		if err := tx.Q.SoftDeletePhoto(ctx, photoID); err != nil {
			return err
		}
		if row.IsPrimary {
			remaining, err := tx.Q.ListPhotosByUser(ctx, userID)
			if err != nil {
				return err
			}
			if len(remaining) > 0 {
				if err := tx.Q.ClearPrimaryPhotos(ctx, userID); err != nil {
					return err
				}
				return tx.Q.SetPrimaryPhoto(ctx, remaining[0].ID)
			}
		}
		return nil
	})
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	_ = s.obj.Delete(ctx, row.ObjectKey)
	return nil
}

func (s *Service) toDTO(ctx context.Context, row sqlcdb.UserPhoto) (Photo, error) {
	url, err := s.obj.SignedURL(ctx, row.ObjectKey, SignedURLTTL)
	if err != nil {
		url = ""
	}
	return Photo{
		ID: row.ID.String(), URL: url, SortOrder: row.SortOrder, IsPrimary: row.IsPrimary,
		ModerationStatus: row.ModerationStatus,
	}, nil
}

func (s *Service) recomputeCompletion(ctx context.Context, tx *store.Store, userID uuid.UUID) error {
	prof, err := tx.Q.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	data := profile.Data{
		Nickname: prof.Nickname, Gender: prof.Gender, BirthDate: prof.BirthDate,
		ProvinceCode: prof.ProvinceCode, CityCode: prof.CityCode, DistrictCode: prof.DistrictCode,
		HeightCM: prof.HeightCm, EducationLevel: prof.EducationLevel,
		OccupationCategory: prof.OccupationCategory, EmploymentStatus: prof.EmploymentStatus,
		AnnualIncomeBand: prof.AnnualIncomeBand, FamilyAssetBand: prof.FamilyAssetBand,
		DebtBand: prof.DebtBand, HouseStatus: prof.HouseStatus, VehicleStatus: prof.VehicleStatus,
		MaritalStatus: prof.MaritalStatus, ChildrenCount: prof.ChildrenCount,
		WantsChildren: prof.WantsChildren, MarriageIntent: prof.MarriageIntent,
		MarriageTimeline: prof.MarriageTimeline, SmokingStatus: prof.SmokingStatus,
		DrinkingStatus: prof.DrinkingStatus, Bio: prof.Bio,
	}
	photos, err := tx.Q.CountPhotosByUser(ctx, userID)
	if err != nil {
		return err
	}
	return tx.Q.UpdateProfileCompletion(ctx, sqlcdb.UpdateProfileCompletionParams{
		UserID: userID, ProfileCompletion: int16(profile.Compute(data, photos > 0)),
	})
}
