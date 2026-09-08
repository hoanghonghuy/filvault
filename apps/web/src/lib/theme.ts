import { ref } from 'vue'

export type ThemeId =
  | 'default'
  | 'cyan'
  | 'teal'
  | 'sage'
  | 'sunshine'
  | 'peach'
  | 'lavender'
  | 'pearl'
  | 'pebble'
  | 'dark'
  | 'material'
  | 'spring'
  | 'summer'
  | 'autumn'
  | 'winter'

export interface ThemeDef {
  id: ThemeId
  nameKey: string
  nameDefault: string
  category: 'colors' | 'seasonal'
  isPro?: boolean
  swatchGradient: string
  accentColor: string
  previewBg: string
  todayAccent: string
}

export const THEMES: ThemeDef[] = [
  {
    id: 'default',
    nameKey: 'themeDefault',
    nameDefault: 'Mặc định',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%)',
    accentColor: '#2563eb',
    previewBg: '#f0f7ff',
    todayAccent: '#2563eb',
  },
  {
    id: 'cyan',
    nameKey: 'themeCyan',
    nameDefault: 'Xanh ngọc',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #7dd3fc 0%, #0284c7 100%)',
    accentColor: '#0284c7',
    previewBg: '#f0f9ff',
    todayAccent: '#0284c7',
  },
  {
    id: 'teal',
    nameKey: 'themeTeal',
    nameDefault: 'Ngọc lam',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #5eead4 0%, #0d9488 100%)',
    accentColor: '#0d9488',
    previewBg: '#f0fdfa',
    todayAccent: '#0d9488',
  },
  {
    id: 'sage',
    nameKey: 'themeSage',
    nameDefault: 'Cây sậy',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #bef264 0%, #84cc16 100%)',
    accentColor: '#65a30d',
    previewBg: '#f7fee7',
    todayAccent: '#65a30d',
  },
  {
    id: 'sunshine',
    nameKey: 'themeSunshine',
    nameDefault: 'Ánh nắng',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #fde68a 0%, #f59e0b 100%)',
    accentColor: '#d97706',
    previewBg: '#fffbeb',
    todayAccent: '#d97706',
  },
  {
    id: 'peach',
    nameKey: 'themePeach',
    nameDefault: 'Đào',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #fbcfe8 0%, #f472b6 100%)',
    accentColor: '#f472b6',
    previewBg: '#fff0f3',
    todayAccent: '#f472b6',
  },
  {
    id: 'lavender',
    nameKey: 'themeLavender',
    nameDefault: 'Tím chiều',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #ddd6fe 0%, #8b5cf6 100%)',
    accentColor: '#7c3aed',
    previewBg: '#fbf8ff',
    todayAccent: '#8b5cf6',
  },
  {
    id: 'pearl',
    nameKey: 'themePearl',
    nameDefault: 'Ngọc trai',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #f8fafc 0%, #cbd5e1 100%)',
    accentColor: '#475569',
    previewBg: '#f8fafc',
    todayAccent: '#475569',
  },
  {
    id: 'pebble',
    nameKey: 'themePebble',
    nameDefault: 'Sỏi',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #94a3b8 0%, #475569 100%)',
    accentColor: '#52525b',
    previewBg: '#fafafa',
    todayAccent: '#52525b',
  },
  {
    id: 'dark',
    nameKey: 'themeDark',
    nameDefault: 'Tối',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #1e293b 0%, #020617 100%)',
    accentColor: '#3b82f6',
    previewBg: '#0f172a',
    todayAccent: '#3b82f6',
  },
  {
    id: 'material',
    nameKey: 'themeMaterial',
    nameDefault: 'Vật liệu của bạn',
    category: 'colors',
    isPro: true,
    swatchGradient: 'linear-gradient(135deg, #a78bfa 0%, #38bdf8 50%, #34d399 100%)',
    accentColor: '#0284c7',
    previewBg: '#f1f5f9',
    todayAccent: '#0284c7',
  },
  // Seasonal
  {
    id: 'spring',
    nameKey: 'themeSpring',
    nameDefault: 'Mùa xuân',
    category: 'seasonal',
    isPro: true,
    swatchGradient: 'linear-gradient(135deg, #f472b6 0%, #34d399 100%)',
    accentColor: '#db2777',
    previewBg: '#fdf2f8',
    todayAccent: '#db2777',
  },
  {
    id: 'summer',
    nameKey: 'themeSummer',
    nameDefault: 'Mùa hè',
    category: 'seasonal',
    isPro: true,
    swatchGradient: 'linear-gradient(135deg, #0ea5e9 0%, #eab308 100%)',
    accentColor: '#0284c7',
    previewBg: '#f0f9ff',
    todayAccent: '#0284c7',
  },
  {
    id: 'autumn',
    nameKey: 'themeAutumn',
    nameDefault: 'Mùa thu',
    category: 'seasonal',
    isPro: true,
    swatchGradient: 'linear-gradient(135deg, #ea580c 0%, #f59e0b 100%)',
    accentColor: '#ea580c',
    previewBg: '#fff7ed',
    todayAccent: '#ea580c',
  },
  {
    id: 'winter',
    nameKey: 'themeWinter',
    nameDefault: 'Mùa đông',
    category: 'seasonal',
    isPro: true,
    swatchGradient: 'linear-gradient(135deg, #38bdf8 0%, #a5b4fc 100%)',
    accentColor: '#0284c7',
    previewBg: '#f0f9ff',
    todayAccent: '#0284c7',
  },
]

const currentColorTheme = ref<ThemeId>(
  (localStorage.getItem('filvault.colorTheme') as ThemeId) || 'default'
)

const followSystemDark = ref<boolean>(
  localStorage.getItem('filvault.followSystemDark') === 'true'
)

const isDarkMode = ref<boolean>(
  typeof document !== 'undefined' &&
    (document.documentElement.dataset.theme === 'dark' ||
      localStorage.getItem('filvault.theme') === 'dark')
)

function setDarkMode(dark: boolean) {
  isDarkMode.value = dark
  if (typeof document !== 'undefined') {
    if (dark) {
      document.documentElement.dataset.theme = 'dark'
      localStorage.setItem('filvault.theme', 'dark')
    } else {
      delete document.documentElement.dataset.theme
      localStorage.setItem('filvault.theme', 'light')
    }
  }
}

function applyColorTheme(themeId: ThemeId) {
  currentColorTheme.value = themeId
  localStorage.setItem('filvault.colorTheme', themeId)
  if (typeof document !== 'undefined') {
    document.documentElement.dataset.colorTheme = themeId
    if (themeId === 'dark') {
      setDarkMode(true)
    }
  }
}

function setFollowSystemDark(follow: boolean) {
  followSystemDark.value = follow
  localStorage.setItem('filvault.followSystemDark', follow ? 'true' : 'false')
  if (follow && typeof window !== 'undefined' && window.matchMedia) {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    setDarkMode(prefersDark)
  }
}

// Attach listener for prefers-color-scheme if followSystemDark is active
if (typeof window !== 'undefined' && window.matchMedia) {
  const mql = window.matchMedia('(prefers-color-scheme: dark)')
  mql.addEventListener('change', (e) => {
    if (followSystemDark.value) {
      setDarkMode(e.matches)
    }
  })
}

export function useTheme() {
  return {
    currentColorTheme,
    followSystemDark,
    isDarkMode,
    applyColorTheme,
    setDarkMode,
    setFollowSystemDark,
    THEMES,
  }
}
