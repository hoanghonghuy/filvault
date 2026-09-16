import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./CallModal.vue', import.meta.url)), 'utf-8')
const copySource = readFileSync(fileURLToPath(new URL('../lib/callModalCopy.ts', import.meta.url)), 'utf-8')

describe('CallModal localization contract', () => {
  it('derives call copy reactively from the shared locale helper', () => {
    expect(source).toContain("import { useI18n } from '@/lib/i18n'")
    expect(source).toContain("import { callModalCopy } from '@/lib/callModalCopy'")
    expect(source).toContain('const { locale } = useI18n()')
    expect(source).toMatch(/const copy = computed\(\(\) => callModalCopy\(locale\.value\)\)/)
    expect(source).not.toMatch(/locale\.value === 'en'/)
  })

  it('binds visible and accessible call controls to localized copy', () => {
    expect(source).toContain(':aria-label="copy.declineCall"')
    expect(source).toContain(':aria-label="copy.answerCall"')
    expect(source).toContain(':aria-label="copy.cancelCall"')
    expect(source).toContain(':aria-label="isFullscreen ? copy.minimize : copy.fullscreen"')
    expect(source).toContain(':aria-label="callStore.isMicEnabled ? copy.muteMic : copy.unmuteMic"')
    expect(source).toContain(':aria-label="callStore.isCamEnabled ? copy.turnCameraOff : copy.turnCameraOn"')
    expect(source).toContain(':aria-label="copy.endCall"')
    expect(source).toContain('{{ copy.callEnded }}')
  })

  it('keeps both supported locale variants for representative call states in callModalCopy', () => {
    expect(copySource).toContain("incomingVideo: 'Incoming video call'")
    expect(copySource).toContain("incomingVideo: 'Cuộc gọi video đến'")
    expect(copySource).toContain("callEnded: 'Call ended'")
    expect(copySource).toContain("callEnded: 'Cuộc gọi đã kết thúc'")
  })
})
