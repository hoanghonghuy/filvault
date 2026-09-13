import type { Locale } from '@/lib/i18n'

export type ProfileAvatarCopy = {
  avatarAlt: string
  fallbackAccount: string
  changeAvatar: string
  removeAvatar: string
  invalidType: string
  tooLarge: string
  heicDecodeFailed: string
  readFailed: string
  invalidData: string
  dimensionsFailed: string
  graphicsFailed: string
  decodeFailed: string
  processFailed: string
  uploadFailed: string
  uploadSuccess: string
  removeFailed: string
  removeSuccess: string
}

const COPY: Record<Locale, ProfileAvatarCopy> = {
  vi: {
    avatarAlt: 'Ảnh đại diện',
    fallbackAccount: 'Tài khoản của bạn',
    changeAvatar: 'Đổi ảnh đại diện',
    removeAvatar: 'Xóa ảnh đại diện',
    invalidType: 'Vui lòng chọn file hình ảnh (JPG, PNG, WebP, HEIC)',
    tooLarge: 'Kích thước ảnh quá lớn (tối đa 25MB)',
    heicDecodeFailed: 'Không thể giải mã file ảnh HEIC. Vui lòng chọn ảnh JPG, PNG hoặc thử lại.',
    readFailed: 'Không thể đọc file ảnh từ thiết bị',
    invalidData: 'Dữ liệu ảnh không hợp lệ',
    dimensionsFailed: 'Không thể xác định kích thước ảnh',
    graphicsFailed: 'Không thể xử lý đồ họa ảnh',
    decodeFailed: 'Không thể giải mã định dạng ảnh này. Vui lòng chọn ảnh JPG, PNG hoặc WebP.',
    processFailed: 'Không thể xử lý ảnh',
    uploadFailed: 'Không thể tải ảnh lên',
    uploadSuccess: 'Đã cập nhật ảnh đại diện',
    removeFailed: 'Không thể xóa ảnh',
    removeSuccess: 'Đã xóa ảnh đại diện',
  },
  en: {
    avatarAlt: 'Profile picture',
    fallbackAccount: 'Your account',
    changeAvatar: 'Change profile picture',
    removeAvatar: 'Remove profile picture',
    invalidType: 'Please choose an image file (JPG, PNG, WebP, HEIC)',
    tooLarge: 'Image is too large (maximum 25MB)',
    heicDecodeFailed: 'Could not decode the HEIC image. Choose a JPG or PNG image, or try again.',
    readFailed: 'Could not read the image from this device',
    invalidData: 'Image data is invalid',
    dimensionsFailed: 'Could not determine the image dimensions',
    graphicsFailed: 'Could not process the image graphics',
    decodeFailed: 'Could not decode this image format. Choose a JPG, PNG, or WebP image.',
    processFailed: 'Could not process the image',
    uploadFailed: 'Could not upload the image',
    uploadSuccess: 'Profile picture updated',
    removeFailed: 'Could not remove the profile picture',
    removeSuccess: 'Profile picture removed',
  },
}

export function profileAvatarCopy(locale: Locale): ProfileAvatarCopy {
  return COPY[locale]
}
