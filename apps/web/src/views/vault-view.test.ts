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
    expect(vault).toMatch(/t\.value\.vaultDownloadFailed/)
    expect(vault).not.toContain("'Kho cá nhân đã tự động khóa'")
    expect(vault).not.toContain("'Đi tới Tệp của tôi'")
    expect(vault).not.toContain("formatApiError(e, 'Download failed')")
  })

  it('shows a localized retryable load-error state before the genuine empty state', () => {
    expect(vault).toContain('const loadError = ref(false)')
    expect(vault).toMatch(/async function loadVaultFiles\(\)[\s\S]*?loadError\.value = false[\s\S]*?await vault\.loadFiles\(\)[\s\S]*?loadError\.value = true/)
    expect(vault).toContain('v-if="loadError" class="vault-empty-box" role="alert"')
    expect(vault).toContain('{{ t.filesLoadFailed }}')
    expect(vault).toContain('@click="loadVaultFiles"')
    expect(vault).toContain('{{ t.retry }}')
    expect(vault).toMatch(/v-if="loadError"[\s\S]*?v-else-if="vault\.files\.length === 0"/)
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

  it('makes vault file rows keyboard-operable with visible focus while isolating nested controls', () => {
    expect(vault).toMatch(/class="vault-file-row tappable"[\s\S]*?role="button"[\s\S]*?tabindex="0"[\s\S]*?:aria-label="file\.name"/)
    expect(vault).toContain('@keydown="onVaultFileKeydown($event, file)"')
    expect(vault).toContain('if (event.target !== event.currentTarget) return')
    expect(vault).toContain("if (event.key !== 'Enter' && event.key !== ' ' && event.key !== 'Spacebar') return")
    expect(vault).toContain('void previewMediaFile(file)')
    expect(vault).toContain('@click.stop="openFileMenu(file)"')
    expect(vault).toMatch(/\.vault-file-row:focus-visible[\s\S]*?outline:\s*2px solid var\(--accent\)[\s\S]*?outline-offset:\s*2px/)
  })
  it('gives each vault file action control a filename-specific accessible name', () => {
    expect(vault).toContain(':aria-label="`${t.vaultFileActions}: ${file.name}`"')
    expect(vault).not.toContain(':aria-label="t.vaultFileActions"')
  })

  it('keeps localized file dates reactive and hides their separator when no date can be rendered', () => {
    expect(vault).toContain("const { t, locale } = useI18n()")
    expect(vault).toContain('v-if="formatVaultDate(file.createdAt, locale)"')
    expect(vault).toContain('{{ formatVaultDate(file.createdAt, locale) }}')
  })

})
