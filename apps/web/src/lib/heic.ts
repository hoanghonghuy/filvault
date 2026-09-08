/**
 * HEIC / HEIF image processing utility for Filvault
 * Decodes and converts HEIC images to JPEG in the browser via WebAssembly (heic-to)
 */

const HEIC_EXT_REGEX = /\.(heic|heif)$/i
const HEIC_MIME_SET = new Set([
  'image/heic',
  'image/heif',
  'image/heic-sequence',
  'image/heif-sequence',
])

/**
 * Check if a file name or MIME type represents a HEIC/HEIF image.
 */
export function isHeic(nameOrMime?: string | null, mime?: string | null): boolean {
  if (!nameOrMime && !mime) return false

  if (mime && HEIC_MIME_SET.has(mime.toLowerCase().trim())) {
    return true
  }

  if (nameOrMime) {
    const trimmed = nameOrMime.toLowerCase().trim()
    if (HEIC_MIME_SET.has(trimmed)) return true
    if (HEIC_EXT_REGEX.test(trimmed)) return true
  }

  return false
}

/**
 * Check if a File or Blob object is HEIC.
 */
export function isHeicFile(file: File | Blob): boolean {
  if ('name' in file && typeof file.name === 'string' && HEIC_EXT_REGEX.test(file.name)) {
    return true
  }
  if (file.type && HEIC_MIME_SET.has(file.type.toLowerCase().trim())) {
    return true
  }
  return false
}

/**
 * Convert a HEIC Blob to a JPEG Blob in the browser.
 */
export async function convertHeicBlobToJpeg(blob: Blob, quality = 0.85): Promise<Blob> {
  const { heicTo } = await import('heic-to')
  return await heicTo({
    blob,
    type: 'image/jpeg',
    quality,
  })
}

/**
 * Convert a HEIC Blob to a JPEG Data URL.
 */
export async function convertHeicBlobToDataUrl(blob: Blob, quality = 0.85): Promise<string> {
  const jpegBlob = await convertHeicBlobToJpeg(blob, quality)
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        resolve(reader.result)
      } else {
        reject(new Error('Failed to read converted image as Data URL'))
      }
    }
    reader.onerror = () => reject(new Error('FileReader error'))
    reader.readAsDataURL(jpegBlob)
  })
}

// In-memory cache for converted display URLs: sourceUrl -> objectUrl
const displayUrlCache = new Map<string, string>()
const pendingConversions = new Map<string, Promise<string>>()

// Concurrency limiter for background conversions (max 2 at once)
let activeConversions = 0
const conversionQueue: Array<() => void> = []

function acquireSlot(): Promise<void> {
  if (activeConversions < 2) {
    activeConversions++
    return Promise.resolve()
  }
  return new Promise((resolve) => {
    conversionQueue.push(() => {
      activeConversions++
      resolve()
    })
  })
}

function releaseSlot(): void {
  activeConversions--
  const next = conversionQueue.shift()
  if (next) {
    next()
  }
}

/**
 * Load a HEIC image from a URL or Blob and return a browser-renderable Object URL.
 * Automatically caches results so the same URL is never decoded twice.
 */
export async function getHeicDisplayUrl(source: string | Blob, quality = 0.85): Promise<string> {
  const isStringSource = typeof source === 'string'

  if (isStringSource) {
    const cached = displayUrlCache.get(source)
    if (cached) return cached

    const pending = pendingConversions.get(source)
    if (pending) return pending
  }

  const conversionPromise = (async () => {
    await acquireSlot()
    try {
      let blob: Blob
      if (isStringSource) {
        const response = await fetch(source)
        if (!response.ok) {
          throw new Error(`Failed to fetch image: ${response.status}`)
        }
        blob = await response.blob()
      } else {
        blob = source
      }

      const jpegBlob = await convertHeicBlobToJpeg(blob, quality)
      const objectUrl = URL.createObjectURL(jpegBlob)

      if (isStringSource) {
        displayUrlCache.set(source, objectUrl)
      }
      return objectUrl
    } finally {
      releaseSlot()
      if (isStringSource) {
        pendingConversions.delete(source)
      }
    }
  })()

  if (isStringSource) {
    pendingConversions.set(source, conversionPromise)
  }

  return conversionPromise
}

/**
 * Revoke and clean up an object URL previously generated for a source URL.
 */
export function cleanupHeicUrl(sourceUrl: string): void {
  const objectUrl = displayUrlCache.get(sourceUrl)
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    displayUrlCache.delete(sourceUrl)
  }
}

/**
 * Trigger download of a HEIC image converted to JPEG.
 */
export async function downloadHeicAsJpeg(url: string, originalName: string): Promise<void> {
  const displayUrl = await getHeicDisplayUrl(url, 0.92)
  const baseName = originalName.replace(HEIC_EXT_REGEX, '') || 'image'
  const downloadName = `${baseName}.jpg`

  const link = document.createElement('a')
  link.href = displayUrl
  link.download = downloadName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
