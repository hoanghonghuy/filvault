import type { Locale } from '@/lib/i18n'

const themeNames = {
  vi: {
    blue: 'Xanh Messenger', purple: 'Hoàng hôn Tím', emerald: 'Ngọc lục bảo',
    rose: 'Hồng ngọt ngào', amber: 'Hổ phách ấm', cyan: 'Đại dương xanh',
    indigo: 'Chàm huyền bí', coral: 'Cam san hô', slate: 'Than chì thanh lịch',
  },
  en: {
    blue: 'Messenger Blue', purple: 'Purple Sunset', emerald: 'Emerald',
    rose: 'Sweet Rose', amber: 'Warm Amber', cyan: 'Ocean Cyan',
    indigo: 'Mystic Indigo', coral: 'Coral Orange', slate: 'Elegant Slate',
  },
} as const

const wallpaperNames = {
  vi: {
    none: 'Mặc định', doodle: 'Doodle họa tiết', sakura: 'Hoa anh đào',
    sky_breeze: 'Mây trời sáng', matcha: 'Trà xanh Matcha', terracotta: 'Gốm ấm áp',
    lavender_mist: 'Oải hương sương mai', notebook_grid: 'Giấy kẻ ô', geometric_pastel: 'Hình học tối giản',
    aurora: 'Cực quang xanh', sunset: 'Hoàng hôn', cosmos: 'Vũ trụ huyền ảo',
    emerald: 'Rừng nhiệt đới', cyber: 'Đêm Neon', amoled_carbon: 'Lưới Carbon AMOLED',
    obsidian_gold: 'Hắc diện kim', ocean_abyss: 'Vực thẳm đại dương', twilight_violet: 'Hoàng hôn Tím đêm',
  },
  en: {
    none: 'Default', doodle: 'Doodle Pattern', sakura: 'Cherry Blossom',
    sky_breeze: 'Sky Breeze', matcha: 'Matcha Green', terracotta: 'Warm Terracotta',
    lavender_mist: 'Lavender Mist', notebook_grid: 'Notebook Grid', geometric_pastel: 'Minimal Geometry',
    aurora: 'Green Aurora', sunset: 'Sunset', cosmos: 'Dreamy Cosmos',
    emerald: 'Tropical Forest', cyber: 'Neon Night', amoled_carbon: 'AMOLED Carbon Grid',
    obsidian_gold: 'Obsidian Gold', ocean_abyss: 'Ocean Abyss', twilight_violet: 'Twilight Violet',
  },
} as const

export function chatThemeName(locale: Locale, id: string): string {
  return themeNames[locale][id as keyof typeof themeNames.vi] ?? id
}

export function chatWallpaperName(locale: Locale, id: string): string {
  return wallpaperNames[locale][id as keyof typeof wallpaperNames.vi] ?? id
}

export function chatCustomWallpaperLabel(locale: Locale): string {
  return locale === 'vi' ? 'Ảnh tùy chỉnh' : 'Custom image'
}

export function chatStorageWallpaperFallback(locale: Locale): string {
  return locale === 'vi' ? 'Ảnh kho cá nhân' : 'Personal storage image'
}
