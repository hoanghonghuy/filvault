import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('VaultView UX contract', () => {
  const vault = readSrc('./VaultView.vue')
  const lightbox = readSrc('../components/MediaLightbox.vue')

  it('uses shared MediaLightbox instead of a custom preview dialog', () => {
    expect(vault).toContain('MediaLightbox')
    expect(vault).not.toMatch(/class="preview-backdrop"/)
    expect(vault).not.toMatch(/class="preview-dialog"/)
    expect(vault).not.toMatch(/\bautoplay\b/)
  })

  it('loads preview URLs through the existing file download endpoint', () => {
    expect(vault).toMatch(/api<\{ downloadUrl: string \}>\(`\/files\/\$\{file\.id\}\/download`\)/)
  })

  it('uses locale strings for core Vault copy instead of hardcoded Vietnamese', () => {
    expect(vault).toMatch(/t\.value\.vaultAutoLockedToast/)
    expect(vault).toMatch(/t\.value\.vaultLockedToast/)
    expect(vault).toMatch(/t\.vaultPinPlaceholder/)
    expect(vault).toMatch(/t\.vaultGoToMyFiles/)
    expect(vault).not.toContain("'Kho cá nhân đã tự động khóa'")
    expect(vault).not.toContain("'Đi tới Tệp của tôi'")
  })

  it('shows distinct credential visibility icons and accessible labels', () => {
    expect(vault).toMatch(/showSetupPin \? 'eye-off' : 'eye'/)
    expect(vault).toMatch(/showUnlockPin \? 'eye-off' : 'eye'/)
    expect(vault).toMatch(/t\.vaultShowCredential/)
    expect(vault).toMatch(/t\.vaultHideCredential/)
  })

  it('meets minimum touch targets and avoids transition: all', () => {
    expect(vault).toMatch(/min-height:\s*var\(--touch-min\)/)
    expect(vault).toMatch(/min-width:\s*var\(--touch-min\)/)
    expect(vault).not.toMatch(/transition:\s*all/)
  })

  it('uses semantic status tokens for locked/unlocked presentation', () => {
    expect(vault).toMatch(/color:\s*var\(--success\)/)
    expect(vault).toMatch(/color:\s*var\(--warning\)/)
    expect(vault).not.toMatch(/#10b981/)
    expect(vault).not.toMatch(/#f59e0b/)
  })

  it('aligns with shared media lightbox dialog semantics', () => {
    expect(lightbox).toMatch(/<dialog/)
    expect(lightbox).toMatch(/@keydown\.esc/)
    expect(vault).toMatch(/@close="previewOpen = false"/)
  })
})
