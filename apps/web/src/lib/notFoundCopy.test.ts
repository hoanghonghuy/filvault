import { describe, expect, it } from 'vitest'
import { notFoundCopy } from './notFoundCopy'

describe('notFoundCopy', () => {
  it('provides Vietnamese not-found labels', () => {
    expect(notFoundCopy('vi')).toEqual({
      eyebrow: 'Không tìm thấy trang',
      title: 'Đường dẫn này không tồn tại',
      body: 'Liên kết có thể đã cũ hoặc địa chỉ đã được nhập sai. Bạn có thể quay lại nơi an toàn mà không mất thông tin tài khoản.',
      home: 'Về trang chủ',
      files: 'Mở Tệp của tôi',
      signIn: 'Đăng nhập',
      back: 'Quay lại',
    })
  })

  it('provides English not-found labels', () => {
    expect(notFoundCopy('en')).toEqual({
      eyebrow: 'Page not found',
      title: 'This path does not exist',
      body: 'The link may be stale or the address may have been mistyped. You can recover safely without exposing account details.',
      home: 'Go home',
      files: 'Open My Files',
      signIn: 'Sign in',
      back: 'Go back',
    })
  })
})
