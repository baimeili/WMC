// 我的照片接口：列表、上传（自动去 EXIF）、删除。
import { API_BASE_URL } from '@/utils/constants'
import { ApiError, authHeader, del, get, parseEnvelope } from './http'
import type { Photo } from './types'

export function listPhotos(): Promise<Photo[]> {
  return get<Photo[]>('/me/photos')
}

export function deletePhoto(photoId: string): Promise<void> {
  return del<void>('/me/photos/' + encodeURIComponent(photoId))
}

// 上传照片：multipart/form-data，字段名 file
export function uploadPhoto(filePath: string, makePrimary = false): Promise<Photo> {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: API_BASE_URL + '/me/photos',
      filePath,
      name: 'file',
      formData: { make_primary: makePrimary ? 'true' : 'false' },
      header: authHeader(),
      success: (response) => {
        let body: unknown = null
        try {
          body = typeof response.data === 'string' ? JSON.parse(response.data) : response.data
        } catch (error) {
          reject(new ApiError('INTERNAL', '照片上传失败，请稍后重试', response.statusCode || 0, ''))
          return
        }
        try {
          resolve(parseEnvelope<Photo>(response.statusCode || 0, body))
        } catch (error) {
          reject(error as ApiError)
        }
      },
      fail: () => {
        reject(new ApiError('NETWORK_ERROR', '照片上传失败，请检查网络后重试', 0, ''))
      }
    })
  })
}
