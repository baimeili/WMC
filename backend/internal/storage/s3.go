package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage 兼容 S3 协议的对象存储（阿里云 OSS / 腾讯云 COS / MinIO 均可用）。
type S3Storage struct {
	client   *s3.Client
	presign  *s3.PresignClient
	bucket   string
	publicURL string
}

// S3Options S3 配置。
type S3Options struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	ForcePathStyle bool
	PublicURL      string
}

// NewS3 创建 S3 存储实现。
func NewS3(ctx context.Context, opt S3Options) (*S3Storage, error) {
	if opt.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET 不能为空")
	}
	if opt.Region == "" {
		opt.Region = "us-east-1"
	}
	cfgOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(opt.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, "")),
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("加载 S3 配置失败: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if opt.Endpoint != "" {
			o.BaseEndpoint = aws.String(opt.Endpoint)
		}
		o.UsePathStyle = opt.ForcePathStyle || opt.Endpoint != ""
	})
	return &S3Storage{
		client:    client,
		presign:   s3.NewPresignClient(client),
		bucket:    opt.Bucket,
		publicURL: opt.PublicURL,
	}, nil
}

// Kind 实现 ObjectStorage。
func (s *S3Storage) Kind() string { return "s3" }

// Put 实现 ObjectStorage。
func (s *S3Storage) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

// Delete 实现 ObjectStorage。
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// SignedURL 实现 ObjectStorage（预签名 GET）。
func (s *S3Storage) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
