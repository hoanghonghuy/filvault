import { describe, expect, it } from 'vitest'
import { chatWallpaperLabelCopy } from './chatWallpaperLabelCopy'

describe('chatWallpaperLabelCopy', () => {
  it('provides Vietnamese wallpaper fallback labels', () => {
    expect(chatWallpaperLabelCopy('vi')).toEqual({
      default: 'Mặc định',
      customImage: 'Ảnh tùy chỉnh',
    })
  })

  it('provides English wallpaper fallback labels', () => {
    expect(chatWallpaperLabelCopy('en')).toEqual({
      default: 'Default',
      customImage: 'Custom image',
    })
  })
})
