export interface WallpaperTheme {
  primary: string
  secondary: string
  gradient: string
  isDark: boolean
  surfaceTint: string
  headerBg: string
  composerBg: string
  borderTint: string
  scrimOverlay: string
}

export const PRESET_THEMES: Record<string, WallpaperTheme> = {
  doodle: {
    primary: '#6366f1',
    secondary: '#8b5cf6',
    gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    isDark: false,
    surfaceTint: 'rgba(99, 102, 241, 0.08)',
    headerBg: 'rgba(255, 255, 255, 0.82)',
    composerBg: 'rgba(255, 255, 255, 0.85)',
    borderTint: 'rgba(99, 102, 241, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.02)',
  },
  aurora: {
    primary: '#10b981',
    secondary: '#06b6d4',
    gradient: 'linear-gradient(135deg, #059669 0%, #06b6d4 100%)',
    isDark: true,
    surfaceTint: 'rgba(16, 185, 129, 0.12)',
    headerBg: 'rgba(9, 26, 36, 0.82)',
    composerBg: 'rgba(9, 26, 36, 0.88)',
    borderTint: 'rgba(16, 185, 129, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.2)',
  },
  sunset: {
    primary: '#f43f5e',
    secondary: '#f97316',
    gradient: 'linear-gradient(135deg, #f43f5e 0%, #f97316 100%)',
    isDark: true,
    surfaceTint: 'rgba(244, 63, 94, 0.12)',
    headerBg: 'rgba(30, 17, 42, 0.82)',
    composerBg: 'rgba(30, 17, 42, 0.88)',
    borderTint: 'rgba(244, 63, 94, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.2)',
  },
  cosmos: {
    primary: '#8b5cf6',
    secondary: '#ec4899',
    gradient: 'linear-gradient(135deg, #8b5cf6 0%, #ec4899 100%)',
    isDark: true,
    surfaceTint: 'rgba(139, 92, 246, 0.12)',
    headerBg: 'rgba(10, 8, 24, 0.82)',
    composerBg: 'rgba(10, 8, 24, 0.88)',
    borderTint: 'rgba(139, 92, 246, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.25)',
  },
  emerald: {
    primary: '#059669',
    secondary: '#10b981',
    gradient: 'linear-gradient(135deg, #047857 0%, #10b981 100%)',
    isDark: true,
    surfaceTint: 'rgba(5, 150, 105, 0.12)',
    headerBg: 'rgba(6, 26, 20, 0.82)',
    composerBg: 'rgba(6, 26, 20, 0.88)',
    borderTint: 'rgba(5, 150, 105, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.2)',
  },
  sakura: {
    primary: '#fb7185',
    secondary: '#c084fc',
    gradient: 'linear-gradient(135deg, #f43f5e 0%, #c084fc 100%)',
    isDark: false,
    surfaceTint: 'rgba(244, 63, 94, 0.08)',
    headerBg: 'rgba(253, 242, 248, 0.82)',
    composerBg: 'rgba(253, 242, 248, 0.88)',
    borderTint: 'rgba(244, 63, 94, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.05)',
  },
  cyber: {
    primary: '#06b6d4',
    secondary: '#ec4899',
    gradient: 'linear-gradient(135deg, #06b6d4 0%, #ec4899 100%)',
    isDark: true,
    surfaceTint: 'rgba(6, 182, 212, 0.12)',
    headerBg: 'rgba(9, 9, 11, 0.82)',
    composerBg: 'rgba(9, 9, 11, 0.88)',
    borderTint: 'rgba(6, 182, 212, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.25)',
  },
  sky_breeze: {
    primary: '#0284c7',
    secondary: '#38bdf8',
    gradient: 'linear-gradient(135deg, #0284c7 0%, #38bdf8 100%)',
    isDark: false,
    surfaceTint: 'rgba(2, 132, 199, 0.08)',
    headerBg: 'rgba(240, 249, 255, 0.82)',
    composerBg: 'rgba(240, 249, 255, 0.88)',
    borderTint: 'rgba(2, 132, 199, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.03)',
  },
  matcha: {
    primary: '#16a34a',
    secondary: '#4ade80',
    gradient: 'linear-gradient(135deg, #15803d 0%, #22c55e 100%)',
    isDark: false,
    surfaceTint: 'rgba(22, 163, 74, 0.08)',
    headerBg: 'rgba(240, 253, 244, 0.82)',
    composerBg: 'rgba(240, 253, 244, 0.88)',
    borderTint: 'rgba(22, 163, 74, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.03)',
  },
  terracotta: {
    primary: '#ea580c',
    secondary: '#fb923c',
    gradient: 'linear-gradient(135deg, #ea580c 0%, #f97316 100%)',
    isDark: false,
    surfaceTint: 'rgba(234, 88, 12, 0.08)',
    headerBg: 'rgba(255, 247, 237, 0.82)',
    composerBg: 'rgba(255, 247, 237, 0.88)',
    borderTint: 'rgba(234, 88, 12, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.03)',
  },
  lavender_mist: {
    primary: '#9333ea',
    secondary: '#c084fc',
    gradient: 'linear-gradient(135deg, #7e22ce 0%, #a855f7 100%)',
    isDark: false,
    surfaceTint: 'rgba(147, 51, 234, 0.08)',
    headerBg: 'rgba(250, 245, 255, 0.82)',
    composerBg: 'rgba(250, 245, 255, 0.88)',
    borderTint: 'rgba(147, 51, 234, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.03)',
  },
  notebook_grid: {
    primary: '#475569',
    secondary: '#64748b',
    gradient: 'linear-gradient(135deg, #334155 0%, #64748b 100%)',
    isDark: false,
    surfaceTint: 'rgba(71, 85, 105, 0.06)',
    headerBg: 'rgba(248, 250, 252, 0.82)',
    composerBg: 'rgba(248, 250, 252, 0.88)',
    borderTint: 'rgba(71, 85, 105, 0.16)',
    scrimOverlay: 'rgba(255, 255, 255, 0.02)',
  },
  geometric_pastel: {
    primary: '#db2777',
    secondary: '#3b82f6',
    gradient: 'linear-gradient(135deg, #db2777 0%, #6366f1 100%)',
    isDark: false,
    surfaceTint: 'rgba(219, 39, 119, 0.08)',
    headerBg: 'rgba(253, 244, 255, 0.82)',
    composerBg: 'rgba(253, 244, 255, 0.88)',
    borderTint: 'rgba(219, 39, 119, 0.18)',
    scrimOverlay: 'rgba(255, 255, 255, 0.03)',
  },
  amoled_carbon: {
    primary: '#71717a',
    secondary: '#a1a1aa',
    gradient: 'linear-gradient(135deg, #27272a 0%, #3f3f46 100%)',
    isDark: true,
    surfaceTint: 'rgba(113, 113, 122, 0.1)',
    headerBg: 'rgba(9, 9, 11, 0.85)',
    composerBg: 'rgba(9, 9, 11, 0.9)',
    borderTint: 'rgba(113, 113, 122, 0.2)',
    scrimOverlay: 'rgba(0, 0, 0, 0.15)',
  },
  obsidian_gold: {
    primary: '#eab308',
    secondary: '#ca8a04',
    gradient: 'linear-gradient(135deg, #854d0e 0%, #ca8a04 50%, #eab308 100%)',
    isDark: true,
    surfaceTint: 'rgba(234, 179, 8, 0.1)',
    headerBg: 'rgba(15, 14, 10, 0.85)',
    composerBg: 'rgba(15, 14, 10, 0.9)',
    borderTint: 'rgba(234, 179, 8, 0.2)',
    scrimOverlay: 'rgba(0, 0, 0, 0.2)',
  },
  ocean_abyss: {
    primary: '#0284c7',
    secondary: '#0ea5e9',
    gradient: 'linear-gradient(135deg, #0369a1 0%, #0284c7 50%, #38bdf8 100%)',
    isDark: true,
    surfaceTint: 'rgba(2, 132, 199, 0.12)',
    headerBg: 'rgba(3, 7, 18, 0.85)',
    composerBg: 'rgba(3, 7, 18, 0.9)',
    borderTint: 'rgba(2, 132, 199, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.2)',
  },
  twilight_violet: {
    primary: '#a855f7',
    secondary: '#ec4899',
    gradient: 'linear-gradient(135deg, #6b21a8 0%, #a855f7 50%, #ec4899 100%)',
    isDark: true,
    surfaceTint: 'rgba(168, 85, 247, 0.12)',
    headerBg: 'rgba(15, 7, 40, 0.85)',
    composerBg: 'rgba(15, 7, 40, 0.9)',
    borderTint: 'rgba(168, 85, 247, 0.22)',
    scrimOverlay: 'rgba(0, 0, 0, 0.22)',
  },
}

export function rgbToHsl(r: number, g: number, b: number): [number, number, number] {
  const rn = r / 255
  const gn = g / 255
  const bn = b / 255
  const max = Math.max(rn, gn, bn)
  const min = Math.min(rn, gn, bn)
  let h = 0
  let s = 0
  const l = (max + min) / 2

  if (max !== min) {
    const d = max - min
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    switch (max) {
      case rn:
        h = (gn - bn) / d + (gn < bn ? 6 : 0)
        break
      case gn:
        h = (bn - rn) / d + 2
        break
      case bn:
        h = (rn - gn) / d + 4
        break
    }
    h *= 60
  }

  return [Math.round(h), s, l]
}

export function hslToRgb(h: number, s: number, l: number): [number, number, number] {
  const c = (1 - Math.abs(2 * l - 1)) * s
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1))
  const m = l - c / 2
  let r = 0
  let g = 0
  let b = 0

  if (h >= 0 && h < 60) {
    r = c; g = x; b = 0
  } else if (h >= 60 && h < 120) {
    r = x; g = c; b = 0
  } else if (h >= 120 && h < 180) {
    r = 0; g = c; b = x
  } else if (h >= 180 && h < 240) {
    r = 0; g = x; b = c
  } else if (h >= 240 && h < 300) {
    r = x; g = 0; b = c
  } else {
    r = c; g = 0; b = x
  }

  return [
    Math.round((r + m) * 255),
    Math.round((g + m) * 255),
    Math.round((b + m) * 255),
  ]
}

export function rgbToHex(r: number, g: number, b: number): string {
  const toHex = (n: number) => Math.max(0, Math.min(255, Math.round(n))).toString(16).padStart(2, '0')
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

/**
 * Fast client-side dynamic palette extraction from an image using 32x32 HTML5 Canvas.
 */
export async function extractPaletteFromImage(src: string): Promise<WallpaperTheme> {
  return new Promise((resolve) => {
    // Default fallback in case image fails to load or canvas is restricted
    const fallbackTheme: WallpaperTheme = {
      primary: '#0084ff',
      secondary: '#00c6ff',
      gradient: 'linear-gradient(135deg, #0084ff 0%, #00c6ff 100%)',
      isDark: false,
      surfaceTint: 'rgba(0, 132, 255, 0.08)',
      headerBg: 'rgba(var(--canvas-rgb, 255, 255, 255), 0.82)',
      composerBg: 'rgba(var(--canvas-rgb, 255, 255, 255), 0.85)',
      borderTint: 'rgba(0, 132, 255, 0.18)',
      scrimOverlay: 'rgba(0, 0, 0, 0.12)',
    }

    if (typeof document === 'undefined') {
      resolve(fallbackTheme)
      return
    }

    let settled = false
    const finish = (theme: WallpaperTheme) => {
      if (settled) return
      settled = true
      clearTimeout(safetyTimer)
      resolve(theme)
    }

    const safetyTimer = setTimeout(() => {
      finish(fallbackTheme)
    }, 500)

    const img = new Image()
    img.crossOrigin = 'anonymous'

    img.onload = () => {
      try {
        const size = 32
        const canvas = document.createElement('canvas')
        canvas.width = size
        canvas.height = size
        const ctx = canvas.getContext('2d', { willReadFrequently: true })

        if (!ctx) {
          finish(fallbackTheme)
          return
        }

        ctx.drawImage(img, 0, 0, size, size)
        const imgData = ctx.getImageData(0, 0, size, size).data
        const pixelCount = size * size

        let totalLuminance = 0
        // 12 hue buckets (every 30 degrees)
        const buckets: { count: number; weight: number; sumR: number; sumG: number; sumB: number }[] = Array.from(
          { length: 12 },
          () => ({ count: 0, weight: 0, sumR: 0, sumG: 0, sumB: 0 }),
        )

        let nonNeutralWeight = 0
        let dominantBucketIdx = -1
        let maxBucketWeight = 0

        for (let i = 0; i < imgData.length; i += 4) {
          const r = imgData[i]!
          const g = imgData[i + 1]!
          const b = imgData[i + 2]!

          // Standard relative luminance
          const lum = 0.299 * r + 0.587 * g + 0.114 * b
          totalLuminance += lum

          const [h, s, l] = rgbToHsl(r, g, b)

          // Filter out near-blacks, near-whites, and low saturation grays
          if (l >= 0.12 && l <= 0.88 && s >= 0.16) {
            const bucketIdx = Math.min(11, Math.floor(h / 30))
            const weight = s * (1 - Math.abs(l - 0.5) * 1.4)
            const bkt = buckets[bucketIdx]!
            bkt.count++
            bkt.weight += weight
            bkt.sumR += r
            bkt.sumG += g
            bkt.sumB += b
            nonNeutralWeight += weight

            if (bkt.weight > maxBucketWeight) {
              maxBucketWeight = bkt.weight
              dominantBucketIdx = bucketIdx
            }
          }
        }

        const avgLuminance = totalLuminance / pixelCount
        const isDark = avgLuminance < 135

        let primaryHex: string
        let secondaryHex: string

        if (dominantBucketIdx >= 0 && nonNeutralWeight > 0) {
          const dom = buckets[dominantBucketIdx]!
          const avgR = dom.sumR / dom.count
          const avgG = dom.sumG / dom.count
          const avgB = dom.sumB / dom.count

          const [domH, domS] = rgbToHsl(avgR, avgG, avgB)
          // Moderate saturation for tasteful, balanced UI colors (avoid garish loud neon)
          const primaryS = Math.min(0.60, Math.max(0.40, domS * 0.8))
          // Lightness calibrated for rich jewel tones and high WCAG contrast with white text
          const primaryL = isDark ? 0.46 : 0.42
          const [pr, pg, pb] = hslToRgb(domH, primaryS, primaryL)
          primaryHex = rgbToHex(pr, pg, pb)

          // Secondary color: shift hue by 22 degrees for a harmonic gradient
          const secH = (domH + 22) % 360
          const [sr, sg, sb] = hslToRgb(secH, primaryS, isDark ? 0.42 : 0.46)
          secondaryHex = rgbToHex(sr, sg, sb)
        } else {
          // If image is mostly grayscale or monochrome
          if (isDark) {
            primaryHex = '#38bdf8'
            secondaryHex = '#6366f1'
          } else {
            primaryHex = '#0284c7'
            secondaryHex = '#4f46e5'
          }
        }

        const [pr, pg, pb] = [
          parseInt(primaryHex.slice(1, 3), 16),
          parseInt(primaryHex.slice(3, 5), 16),
          parseInt(primaryHex.slice(5, 7), 16),
        ]

        const gradient = `linear-gradient(135deg, ${primaryHex} 0%, ${secondaryHex} 100%)`
        const surfaceTint = isDark
          ? `rgba(${pr}, ${pg}, ${pb}, 0.16)`
          : `rgba(${pr}, ${pg}, ${pb}, 0.08)`
        const borderTint = isDark
          ? `rgba(${pr}, ${pg}, ${pb}, 0.28)`
          : `rgba(${pr}, ${pg}, ${pb}, 0.20)`

        const headerBg = isDark
          ? `rgba(${Math.round(pr * 0.15 + 10)}, ${Math.round(pg * 0.15 + 12)}, ${Math.round(pb * 0.15 + 18)}, 0.82)`
          : `rgba(255, 255, 255, 0.84)`

        const composerBg = isDark
          ? `rgba(${Math.round(pr * 0.15 + 10)}, ${Math.round(pg * 0.15 + 12)}, ${Math.round(pb * 0.15 + 18)}, 0.88)`
          : `rgba(255, 255, 255, 0.88)`

        const scrimOverlay = isDark ? 'rgba(0, 0, 0, 0.42)' : 'rgba(0, 0, 0, 0.12)'

        finish({
          primary: primaryHex,
          secondary: secondaryHex,
          gradient,
          isDark,
          surfaceTint,
          headerBg,
          composerBg,
          borderTint,
          scrimOverlay,
        })
      } catch {
        finish(fallbackTheme)
      }
    }

    img.onerror = () => {
      finish(fallbackTheme)
    }

    img.src = src
  })
}

const memoryPaletteCache = new Map<string, WallpaperTheme>()

export async function resolveWallpaperTheme(
  wallpaperIdOrUrl: string | null | undefined,
): Promise<WallpaperTheme | null> {
  if (!wallpaperIdOrUrl || wallpaperIdOrUrl === 'none') {
    return null
  }

  // Check preset themes first
  if (PRESET_THEMES[wallpaperIdOrUrl]) {
    return PRESET_THEMES[wallpaperIdOrUrl]
  }

  // Check in-memory cache
  if (memoryPaletteCache.has(wallpaperIdOrUrl)) {
    return memoryPaletteCache.get(wallpaperIdOrUrl)!
  }

  // Check localStorage cache
  try {
    const raw = localStorage.getItem('filvault.chat.wallpaper_palettes')
    if (raw) {
      const parsed = JSON.parse(raw) as Record<string, WallpaperTheme>
      if (parsed[wallpaperIdOrUrl]) {
        memoryPaletteCache.set(wallpaperIdOrUrl, parsed[wallpaperIdOrUrl]!)
        return parsed[wallpaperIdOrUrl]!
      }
    }
  } catch {}

  // Extract from image dynamically
  const extracted = await extractPaletteFromImage(wallpaperIdOrUrl)
  memoryPaletteCache.set(wallpaperIdOrUrl, extracted)

  // Persist into localStorage
  try {
    const raw = localStorage.getItem('filvault.chat.wallpaper_palettes')
    const parsed = raw ? (JSON.parse(raw) as Record<string, WallpaperTheme>) : {}
    // Keep cache size bounded to last 20 wallpapers
    const keys = Object.keys(parsed)
    if (keys.length > 20) {
      delete parsed[keys[0]!]
    }
    parsed[wallpaperIdOrUrl] = extracted
    localStorage.setItem('filvault.chat.wallpaper_palettes', JSON.stringify(parsed))
  } catch {}

  return extracted
}
