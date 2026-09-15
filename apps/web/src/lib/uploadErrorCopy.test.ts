import { describe, expect, it } from 'vitest'
import { uploadErrorCopy } from '@/lib/uploadErrorCopy'

describe('uploadErrorCopy', () => {
  it('provides Vietnamese upload and API failure copy', () => {
    const copy = uploadErrorCopy('vi')
    expect(copy.fallback).toBe('Không thể tải tệp lên')
    expect(copy.api.network).toContain('Không thể kết nối')
    expect(copy.api.codes?.QUOTA_EXCEEDED).toContain('Dung lượng')
    expect(copy.api.codes?.UPLOAD_EXPIRED).toContain('hết hạn')
  })

  it('preserves existing English formatter defaults', () => {
    const copy = uploadErrorCopy('en')
    expect(copy.fallback).toBe('Upload failed')
    expect(copy.api).toEqual({})
  })
})
