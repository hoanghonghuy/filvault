import { computed, ref } from 'vue'

export type AppearanceMode = 'system' | 'light' | 'dark'

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
  | 'material'
  | 'spring'
  | 'summer'
  | 'autumn'
  | 'winter'

export const APPEARANCE_MODE_STORAGE_KEY = 'filvault.appearanceMode'
export const COLOR_THEME_STORAGE_KEY = 'filvault.colorTheme'

const LEGACY_THEME_KEY = 'filvault.theme'
const LEGACY_FOLLOW_SYSTEM_KEY = 'filvault.followSystemDark'
const LEGACY_DARK_COLOR_THEME = 'dark'

export interface ThemeDef {
  id: ThemeId
  nameKey: string
  nameDefault: string
  category: 'colors' | 'seasonal'
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
    id: 'material',
    nameKey: 'themeMaterial',
    nameDefault: 'Vật liệu của bạn',
    category: 'colors',
    swatchGradient: 'linear-gradient(135deg, #a78bfa 0%, #38bdf8 50%, #34d399 100%)',
    accentColor: '#0284c7',
    previewBg: '#f1f5f9',
    todayAccent: '#0284c7',
  },
  {
    id: 'spring',
    nameKey: 'themeSpring',
    nameDefault: 'Mùa xuân',
    category: 'seasonal',
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
    swatchGradient: 'linear-gradient(135deg, #38bdf8 0%, #a5b4fc 100%)',
    accentColor: '#0284c7',
    previewBg: '#f0f9ff',
    todayAccent: '#0284c7',
  },
]

const VALID_THEME_IDS = new Set(THEMES.map((theme) => theme.id))

function isAppearanceMode(value: string | null): value is AppearanceMode {
  return value === 'system' || value === 'light' || value === 'dark'
}

function readSystemPrefersDark(): boolean {
  return typeof window !== 'undefined' &&
    window.matchMedia !== undefined &&
    window.matchMedia('(prefers-color-scheme: dark)').matches
}

export function resolveAppearanceIsDark(mode: AppearanceMode): boolean {
  if (mode === 'dark') return true
  if (mode === 'light') return false
  return readSystemPrefersDark()
}

function normalizeColorThemeId(raw: string | null): ThemeId {
  if (raw === LEGACY_DARK_COLOR_THEME) return 'default'
  if (raw && VALID_THEME_IDS.has(raw as ThemeId)) return raw as ThemeId
  return 'default'
}

export function migrateAppearanceMode(): AppearanceMode {
  if (typeof localStorage === 'undefined') return 'light'

  const stored = localStorage.getItem(APPEARANCE_MODE_STORAGE_KEY)
  if (isAppearanceMode(stored)) return stored

  const legacyColorTheme = localStorage.getItem(COLOR_THEME_STORAGE_KEY)
  const legacyFollowSystem = localStorage.getItem(LEGACY_FOLLOW_SYSTEM_KEY) === 'true'
  const legacyTheme = localStorage.getItem(LEGACY_THEME_KEY)

  if (legacyFollowSystem) return 'system'
  if (legacyColorTheme === LEGACY_DARK_COLOR_THEME || legacyTheme === 'dark') return 'dark'
  return 'light'
}

function persistAppearanceMode(mode: AppearanceMode) {
  localStorage.setItem(APPEARANCE_MODE_STORAGE_KEY, mode)
  localStorage.removeItem(LEGACY_THEME_KEY)
  localStorage.setItem(LEGACY_FOLLOW_SYSTEM_KEY, mode === 'system' ? 'true' : 'false')
}

function applyResolvedDarkToDom(isDark: boolean) {
  if (typeof document === 'undefined') return
  if (isDark) {
    document.documentElement.dataset.theme = 'dark'
  } else {
    delete document.documentElement.dataset.theme
  }
}

function applyColorThemeToDom(themeId: ThemeId) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.colorTheme = themeId
}

function migratePersistedState(): { appearanceMode: AppearanceMode; colorTheme: ThemeId } {
  const legacyColorTheme = localStorage.getItem(COLOR_THEME_STORAGE_KEY)
  const appearanceMode = migrateAppearanceMode()
  const colorTheme = normalizeColorThemeId(legacyColorTheme)

  persistAppearanceMode(appearanceMode)
  localStorage.setItem(COLOR_THEME_STORAGE_KEY, colorTheme)
  applyResolvedDarkToDom(resolveAppearanceIsDark(appearanceMode))
  applyColorThemeToDom(colorTheme)

  return { appearanceMode, colorTheme }
}

const initialState =
  typeof localStorage !== 'undefined' ? migratePersistedState() : { appearanceMode: 'light' as AppearanceMode, colorTheme: 'default' as ThemeId }

const appearanceMode = ref<AppearanceMode>(initialState.appearanceMode)
const currentColorTheme = ref<ThemeId>(initialState.colorTheme)
const systemPrefersDark = ref(readSystemPrefersDark())

function syncReactiveAppearanceState(state: { appearanceMode: AppearanceMode; colorTheme: ThemeId }) {
  appearanceMode.value = state.appearanceMode
  currentColorTheme.value = state.colorTheme
  systemPrefersDark.value = readSystemPrefersDark()
}

export function hydrateAppearance() {
  if (typeof localStorage === 'undefined' || typeof document === 'undefined') return
  syncReactiveAppearanceState(migratePersistedState())
}

/** Re-read persisted appearance from storage into reactive state (used in tests and after external resets). */
export function syncAppearanceFromStorage() {
  if (typeof localStorage === 'undefined') return
  syncReactiveAppearanceState(migratePersistedState())
}

const resolvedIsDark = computed(() => {
  if (appearanceMode.value === 'dark') return true
  if (appearanceMode.value === 'light') return false
  return systemPrefersDark.value
})

function applyAppearanceMode(mode: AppearanceMode) {
  appearanceMode.value = mode
  if (mode === 'system') {
    systemPrefersDark.value = readSystemPrefersDark()
  }
  persistAppearanceMode(mode)
  applyResolvedDarkToDom(mode === 'system' ? systemPrefersDark.value : mode === 'dark')
}

function setAppearanceMode(mode: AppearanceMode) {
  applyAppearanceMode(mode)
}

function applyColorTheme(themeId: ThemeId) {
  currentColorTheme.value = themeId
  localStorage.setItem(COLOR_THEME_STORAGE_KEY, themeId)
  applyColorThemeToDom(themeId)
}

function toggleResolvedAppearance() {
  setAppearanceMode(resolvedIsDark.value ? 'light' : 'dark')
}

export function applySystemAppearanceIfNeeded() {
  systemPrefersDark.value = readSystemPrefersDark()
  if (appearanceMode.value === 'system') {
    applyResolvedDarkToDom(systemPrefersDark.value)
  }
}

if (typeof window !== 'undefined' && window.matchMedia) {
  const mql = window.matchMedia('(prefers-color-scheme: dark)')
  mql.addEventListener('change', () => {
    applySystemAppearanceIfNeeded()
  })
}

export function useTheme() {
  return {
    appearanceMode,
    currentColorTheme,
    resolvedIsDark,
    isDarkMode: resolvedIsDark,
    followSystemDark: computed(() => appearanceMode.value === 'system'),
    applyColorTheme,
    setAppearanceMode,
    setDarkMode: (dark: boolean) => setAppearanceMode(dark ? 'dark' : 'light'),
    setFollowSystemDark: (follow: boolean) => setAppearanceMode(follow ? 'system' : (resolvedIsDark.value ? 'dark' : 'light')),
    toggleResolvedAppearance,
    THEMES,
  }
}
