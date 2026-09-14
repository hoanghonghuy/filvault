import { describe, expect, it } from 'vitest'
import {
  chatCustomWallpaperLabel,
  chatStorageWallpaperFallback,
  chatThemeName,
  chatWallpaperName,
} from './chatPersonalizationCopy'

describe('chat personalization copy', () => {
  it('localizes representative chat theme names in both locales', () => {
    expect(chatThemeName('vi', 'blue')).toBe('Xanh Messenger')
    expect(chatThemeName('en', 'blue')).toBe('Messenger Blue')
    expect(chatThemeName('vi', 'purple')).toBe('Hoàng hôn Tím')
    expect(chatThemeName('en', 'purple')).toBe('Purple Sunset')
  })

  it('localizes light and dark wallpaper preset names', () => {
    expect(chatWallpaperName('vi', 'sakura')).toBe('Hoa anh đào')
    expect(chatWallpaperName('en', 'sakura')).toBe('Cherry Blossom')
    expect(chatWallpaperName('vi', 'ocean_abyss')).toBe('Vực thẳm đại dương')
    expect(chatWallpaperName('en', 'ocean_abyss')).toBe('Ocean Abyss')
  })

  it('localizes system-provided wallpaper labels without changing unknown identifiers', () => {
    expect(chatWallpaperName('vi', 'none')).toBe('Mặc định')
    expect(chatWallpaperName('en', 'none')).toBe('Default')
    expect(chatCustomWallpaperLabel('vi')).toBe('Ảnh tùy chỉnh')
    expect(chatCustomWallpaperLabel('en')).toBe('Custom image')
    expect(chatStorageWallpaperFallback('vi')).toBe('Ảnh kho cá nhân')
    expect(chatStorageWallpaperFallback('en')).toBe('Personal storage image')
    expect(chatWallpaperName('en', 'future-preset')).toBe('future-preset')
  })
})
