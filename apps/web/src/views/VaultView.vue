<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useVaultStore } from '@/stores/vault'
import { useUiStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { mimeIcon } from '@/lib/mimeIcon'
import Icon from '@/components/AppIcon.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import type { VaultFile } from '@/api/types'

const router = useRouter()
const vault = useVaultStore()
const ui = useUiStore()
const { t } = useI18n()

// Setup state
const setupPin = ref('')
const confirmSetupPin = ref('')
const showSetupPin = ref(false)
const setupError = ref('')

// Unlock state
const unlockPin = ref('')
const showUnlockPin = ref(false)
const unlockError = ref('')

// Change PIN modal
const changePinOpen = ref(false)
const currentPin = ref('')
const newPin = ref('')
const confirmNewPin = ref('')
const changePinError = ref('')
const changingPin = ref(false)

// Reset PIN modal (using account password)
const resetPinOpen = ref(false)
const accountPassword = ref('')
const resetNewPin = ref('')
const confirmResetNewPin = ref('')
const resetPinError = ref('')
const resettingPin = ref(false)

// Actions on a vault file
const selectedFile = ref<VaultFile | null>(null)
const fileActionOpen = ref(false)

// In-app preview
const previewOpen = ref(false)
const previewFile = ref<{ id: string; name: string; mimeType: string; url: string } | null>(null)

const totalVaultSize = computed(() => {
  return vault.files.reduce((acc, f) => acc + (f.sizeBytes || 0), 0)
})

let hiddenTimeout: ReturnType<typeof setTimeout> | null = null

function onVisibilityChange() {
  if (document.hidden) {
    if (vault.isUnlocked) {
      hiddenTimeout = setTimeout(() => {
        if (vault.isUnlocked) {
          vault.lock()
          ui.showToast(t.value.vaultAutoLockedToast, 'info')
        }
      }, 2 * 60 * 1000)
    }
  } else {
    if (hiddenTimeout) {
      clearTimeout(hiddenTimeout)
      hiddenTimeout = null
    }
  }
}

onMounted(async () => {
  document.addEventListener('visibilitychange', onVisibilityChange)
  try {
    await vault.fetchStatus()
    if (vault.isUnlocked) {
      await vault.loadFiles()
    }
  } catch {
    // ignore
  }
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  if (hiddenTimeout) {
    clearTimeout(hiddenTimeout)
    hiddenTimeout = null
  }
})

async function handleSetup() {
  setupError.value = ''
  if (setupPin.value.trim().length < 4) {
    setupError.value = t.value.vaultPinMinLength
    return
  }
  if (setupPin.value !== confirmSetupPin.value) {
    setupError.value = t.value.vaultPinMismatch
    return
  }

  try {
    await vault.setup(setupPin.value.trim())
    ui.showToast(t.value.vaultUnlockedNotice, 'success')
    setupPin.value = ''
    confirmSetupPin.value = ''
  } catch (e) {
    setupError.value = formatApiError(e, t.value.vaultSetupFailed)
  }
}

async function handleUnlock() {
  unlockError.value = ''
  if (!unlockPin.value.trim()) return

  try {
    await vault.unlock(unlockPin.value.trim())
    ui.showToast(t.value.vaultUnlockedNotice, 'success')
    unlockPin.value = ''
  } catch (e) {
    unlockError.value = formatApiError(e, t.value.vaultUnlockFailed)
  }
}

function handleLockNow() {
  vault.lock()
  ui.showToast(t.value.vaultLockedToast, 'info')
}

async function handleChangePin() {
  changePinError.value = ''
  if (newPin.value.trim().length < 4) {
    changePinError.value = t.value.vaultPinMinLength
    return
  }
  if (newPin.value !== confirmNewPin.value) {
    changePinError.value = t.value.vaultPinMismatch
    return
  }

  changingPin.value = true
  try {
    await vault.changePin(currentPin.value.trim(), newPin.value.trim())
    ui.showToast(t.value.vaultPinChanged, 'success')
    changePinOpen.value = false
    currentPin.value = ''
    newPin.value = ''
    confirmNewPin.value = ''
  } catch (e) {
    changePinError.value = formatApiError(e, t.value.vaultChangePinFailed)
  } finally {
    changingPin.value = false
  }
}

async function handleResetPin() {
  resetPinError.value = ''
  if (!accountPassword.value) {
    resetPinError.value = t.value.vaultResetPasswordRequired
    return
  }
  if (resetNewPin.value.trim().length < 4) {
    resetPinError.value = t.value.vaultPinMinLength
    return
  }
  if (resetNewPin.value !== confirmResetNewPin.value) {
    resetPinError.value = t.value.vaultPinMismatch
    return
  }

  resettingPin.value = true
  try {
    await vault.resetPin(accountPassword.value, resetNewPin.value.trim())
    ui.showToast(t.value.vaultUnlockedNotice, 'success')
    resetPinOpen.value = false
    accountPassword.value = ''
    resetNewPin.value = ''
    confirmResetNewPin.value = ''
  } catch (e) {
    resetPinError.value = formatApiError(e, t.value.vaultResetFailed)
  } finally {
    resettingPin.value = false
  }
}

function openFileMenu(f: VaultFile) {
  selectedFile.value = f
  fileActionOpen.value = true
}

async function downloadFile(id: string) {
  try {
    const out = await api<{ downloadUrl: string }>(`/files/${id}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    ui.showToast(formatApiError(e, 'Download failed'), 'error')
  }
}

async function previewMediaFile(file: VaultFile) {
  try {
    const out = await api<{ downloadUrl: string }>(`/files/${file.id}/download`)
    previewFile.value = { id: file.id, name: file.name, mimeType: file.mimeType, url: out.downloadUrl }
    previewOpen.value = true
  } catch (e) {
    ui.showToast(formatApiError(e, t.value.vaultPreviewFailed), 'error')
  }
}

async function removeSelectedFileFromVault() {
  if (!selectedFile.value) return
  const f = selectedFile.value
  fileActionOpen.value = false
  try {
    await vault.removeFromVault([f.id])
    ui.showToast(t.value.vaultMoveOutSuccess, 'success')
  } catch (e) {
    ui.showToast(formatApiError(e, t.value.vaultMoveOutFailed), 'error')
  }
}

async function deleteSelectedFile() {
  if (!selectedFile.value) return
  const f = selectedFile.value
  fileActionOpen.value = false
  const ok = await ui.confirm({
    title: t.value.vaultDeleteTitle,
    message: `"${f.name}" ${t.value.vaultDeleteMessage}`,
    confirmLabel: t.value.vaultDeleteConfirm,
    danger: true,
  })
  if (!ok) return

  try {
    await api(`/files/${f.id}`, { method: 'DELETE' })
    await vault.loadFiles()
    ui.showToast(t.value.vaultFileDeleted, 'success')
  } catch (e) {
    ui.showToast(formatApiError(e, t.value.vaultDeleteFailed), 'error')
  }
}

function formatDate(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="vault-page">
    <!-- Top Header -->
    <header class="vault-header">
      <button type="button" class="back-btn" :title="t.back" :aria-label="t.back" @click="router.back()">
        <Icon name="arrow-left" :size="20" />
      </button>
      <div class="header-titles">
        <h1 class="page-title">{{ t.vaultTitle }}</h1>
        <span v-if="vault.isUnlocked" class="status-badge unlocked">
          <Icon name="unlock" :size="13" />
          {{ t.vaultUnlockedNotice }}
        </span>
        <span v-else class="status-badge locked">
          <Icon name="lock" :size="13" />
          {{ t.vaultLockedTitle }}
        </span>
      </div>

      <div class="header-actions">
        <button
          v-if="vault.isUnlocked"
          type="button"
          class="lock-btn"
          :title="t.vaultLockNow"
          :aria-label="t.vaultLockNowAria"
          @click="handleLockNow"
        >
          <Icon name="lock" :size="16" />
          <span class="desktop-only">{{ t.vaultLockNow }}</span>
        </button>
      </div>
    </header>

    <!-- Main Content State 1: Setup PIN (First time) -->
    <div v-if="!vault.isInitialized" class="vault-state-card setup-card">
      <div class="shield-icon-box">
        <Icon name="shield" :size="40" class="shield-icon" />
      </div>
      <h2 class="state-title">{{ t.vaultSetupTitle }}</h2>
      <p class="state-desc">{{ t.vaultSetupDesc }}</p>

      <form class="vault-form" @submit.prevent="handleSetup">
        <p v-if="setupError" class="form-error" role="alert">{{ setupError }}</p>

        <label class="field">
          <span>{{ t.vaultPin }}</span>
          <div class="input-wrap">
            <input
              v-model="setupPin"
              :type="showSetupPin ? 'text' : 'password'"
              :placeholder="t.vaultPinPlaceholder"
              maxlength="32"
              required
              autofocus
            />
            <button
              type="button"
              class="eye-btn"
              :aria-label="showSetupPin ? t.vaultHideCredential : t.vaultShowCredential"
              :aria-pressed="showSetupPin"
              @click="showSetupPin = !showSetupPin"
            >
              <Icon :name="showSetupPin ? 'eye-off' : 'eye'" :size="18" />
            </button>
          </div>
        </label>

        <label class="field">
          <span>{{ t.vaultPinConfirm }}</span>
          <div class="input-wrap">
            <input
              v-model="confirmSetupPin"
              :type="showSetupPin ? 'text' : 'password'"
              :placeholder="t.vaultPinConfirmPlaceholder"
              maxlength="32"
              required
            />
          </div>
        </label>

        <button type="submit" class="btn primary submit-btn" :disabled="vault.loading">
          {{ vault.loading ? t.vaultCreating : t.vaultCreateAndEnter }}
        </button>
      </form>
    </div>

    <!-- Main Content State 2: Locked (Enter PIN) -->
    <div v-else-if="!vault.isUnlocked" class="vault-state-card locked-card">
      <div class="lock-icon-box">
        <Icon name="lock" :size="44" class="lock-icon" />
      </div>
      <h2 class="state-title">{{ t.vaultLockedTitle }}</h2>
      <p class="state-desc">{{ t.vaultLockedDesc }}</p>

      <form class="vault-form" @submit.prevent="handleUnlock">
        <p v-if="unlockError" class="form-error" role="alert">{{ unlockError }}</p>

        <label class="field">
          <div class="input-wrap">
            <input
              v-model="unlockPin"
              :type="showUnlockPin ? 'text' : 'password'"
              :placeholder="t.vaultUnlockPlaceholder"
              maxlength="32"
              required
              autofocus
            />
            <button
              type="button"
              class="eye-btn"
              :aria-label="showUnlockPin ? t.vaultHideCredential : t.vaultShowCredential"
              :aria-pressed="showUnlockPin"
              @click="showUnlockPin = !showUnlockPin"
            >
              <Icon :name="showUnlockPin ? 'eye-off' : 'eye'" :size="18" />
            </button>
          </div>
        </label>

        <button type="submit" class="btn primary submit-btn" :disabled="vault.loading || !unlockPin">
          {{ vault.loading ? t.vaultUnlocking : t.vaultUnlock }}
        </button>

        <div class="form-footer">
          <button type="button" class="link-btn" @click="resetPinOpen = true">
            {{ t.vaultForgotPassword }}
          </button>
        </div>
      </form>
    </div>

    <!-- Main Content State 3: Unlocked (File Manager) -->
    <div v-else class="vault-unlocked-surface">
      <!-- Toolbar -->
      <div class="vault-toolbar">
        <div class="vault-meta-info">
          <span class="file-count">{{ vault.files.length }} {{ t.vaultFilesUnit }}</span>
          <span class="dot-sep">•</span>
          <span class="total-size">{{ formatBytes(totalVaultSize) }}</span>
        </div>

        <div class="toolbar-actions">
          <button type="button" class="btn text-btn" @click="changePinOpen = true">
            <Icon name="settings" :size="16" />
            <span>{{ t.vaultChangePin }}</span>
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="vault.files.length === 0" class="vault-empty-box">
        <div class="empty-icon-wrap">
          <Icon name="shield" :size="48" />
        </div>
        <h3 class="empty-title">{{ t.vaultEmpty }}</h3>
        <p class="empty-desc">{{ t.vaultEmptyDesc }}</p>
        <button type="button" class="btn primary" @click="router.push('/files')">
          <Icon name="folder" :size="18" />
          <span>{{ t.vaultGoToMyFiles }}</span>
        </button>
      </div>

      <!-- Files List -->
      <div v-else class="vault-file-list">
        <div
          v-for="file in vault.files"
          :key="file.id"
          class="vault-file-row tappable"
          @click="previewMediaFile(file)"
        >
          <div class="file-icon-box" :class="mimeIcon(file.mimeType)">
            <Icon :name="mimeIcon(file.mimeType)" :size="24" />
          </div>

          <div class="file-info">
            <h4 class="file-name" :title="file.name">{{ file.name }}</h4>
            <div class="file-subtext">
              <span>{{ formatBytes(file.sizeBytes) }}</span>
              <span class="dot-sep">•</span>
              <span>{{ formatDate(file.createdAt) }}</span>
            </div>
          </div>

          <button
            type="button"
            class="more-btn"
            :aria-label="t.vaultFileActions"
            @click.stop="openFileMenu(file)"
          >
            <Icon name="more" :size="20" />
          </button>
        </div>
      </div>
    </div>

    <!-- Bottom Sheet: Single File Actions -->
    <BottomSheet
      :open="fileActionOpen"
      :title="selectedFile?.name || t.vaultFileActions"
      @close="fileActionOpen = false"
    >
      <div class="actions-menu">
        <button
          type="button"
          class="sheet-row"
          @click="selectedFile && previewMediaFile(selectedFile); fileActionOpen = false"
        >
          <Icon name="eye" :size="20" />
          <span>{{ t.preview }}</span>
        </button>
        <button
          type="button"
          class="sheet-row"
          @click="selectedFile && downloadFile(selectedFile.id); fileActionOpen = false"
        >
          <Icon name="download" :size="20" />
          <span>{{ t.download }}</span>
        </button>
        <button
          type="button"
          class="sheet-row"
          @click="removeSelectedFileFromVault"
        >
          <Icon name="arrow-right" :size="20" />
          <span>{{ t.vaultMoveOutOfVault }}</span>
        </button>
        <button
          type="button"
          class="sheet-row danger"
          @click="deleteSelectedFile"
        >
          <Icon name="trash" :size="20" />
          <span>{{ t.delete }}</span>
        </button>
      </div>
    </BottomSheet>

    <!-- Bottom Sheet: Change PIN -->
    <BottomSheet
      :open="changePinOpen"
      :title="t.vaultChangePin"
      @close="changePinOpen = false"
    >
      <form class="modal-form" @submit.prevent="handleChangePin">
        <p v-if="changePinError" class="form-error" role="alert">{{ changePinError }}</p>

        <label class="field">
          <span>{{ t.vaultCurrentPin }}</span>
          <input v-model="currentPin" type="password" required maxlength="32" />
        </label>

        <label class="field">
          <span>{{ t.vaultNewPin }}</span>
          <input v-model="newPin" type="password" required maxlength="32" />
        </label>

        <label class="field">
          <span>{{ t.vaultPinConfirm }}</span>
          <input v-model="confirmNewPin" type="password" required maxlength="32" />
        </label>

        <div class="modal-actions">
          <button type="button" class="btn text-btn" @click="changePinOpen = false">{{ t.cancel }}</button>
          <button type="submit" class="btn primary" :disabled="changingPin">
            {{ changingPin ? t.saving : t.vaultSaveNewPassword }}
          </button>
        </div>
      </form>
    </BottomSheet>

    <!-- Bottom Sheet: Reset PIN (using account password) -->
    <BottomSheet
      :open="resetPinOpen"
      :title="t.vaultResetTitle"
      @close="resetPinOpen = false"
    >
      <form class="modal-form" @submit.prevent="handleResetPin">
        <p class="modal-desc">{{ t.vaultResetDesc }}</p>
        <p v-if="resetPinError" class="form-error" role="alert">{{ resetPinError }}</p>

        <label class="field">
          <span>{{ t.vaultAccountPassword }}</span>
          <input v-model="accountPassword" type="password" required autocomplete="current-password" />
        </label>

        <label class="field">
          <span>{{ t.vaultNewPin }}</span>
          <input v-model="resetNewPin" type="password" required maxlength="32" />
        </label>

        <label class="field">
          <span>{{ t.vaultPinConfirm }}</span>
          <input v-model="confirmResetNewPin" type="password" required maxlength="32" />
        </label>

        <div class="modal-actions">
          <button type="button" class="btn text-btn" @click="resetPinOpen = false">{{ t.cancel }}</button>
          <button type="submit" class="btn primary" :disabled="resettingPin">
            {{ resettingPin ? t.vaultProcessing : t.vaultResetAndEnter }}
          </button>
        </div>
      </form>
    </BottomSheet>

    <MediaLightbox
      :open="previewOpen"
      :name="previewFile?.name ?? ''"
      :mime-type="previewFile?.mimeType ?? ''"
      :url="previewFile?.url ?? ''"
      @download="previewFile ? downloadFile(previewFile.id) : undefined"
      @close="previewOpen = false"
    />
  </div>
</template>

<style scoped>
.vault-page {
  display: flex;
  flex-direction: column;
  max-width: 800px;
  margin: 0 auto;
  padding: var(--space-md) var(--space-shell);
}

.vault-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--touch-min);
  min-height: var(--touch-min);
  border: none;
  border-radius: var(--radius-full, 9999px);
  background: var(--surface-soft);
  color: var(--ink);
  cursor: pointer;
  transition: background var(--duration-short) var(--ease-standard);
}

.back-btn:hover {
  background: var(--surface-card);
}

.header-titles {
  flex: 1;
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex-wrap: wrap;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--ink);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.unlocked {
  background: color-mix(in srgb, var(--success) 12%, transparent);
  color: var(--success);
}

.status-badge.locked {
  background: color-mix(in srgb, var(--warning) 12%, transparent);
  color: var(--warning);
}

.lock-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: var(--touch-min);
  padding: 0 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--ink);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background var(--duration-short) var(--ease-standard),
    border-color var(--duration-short) var(--ease-standard);
}

.lock-btn:hover {
  background: var(--surface-soft);
  border-color: var(--hairline-strong, #cbd5e1);
}

/* State Cards */
.vault-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 440px;
  margin: var(--space-xl) auto;
  padding: var(--space-xl) var(--space-lg);
  border-radius: var(--radius-xl);
  background: var(--surface);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  border: 1px solid var(--hairline);
}

.shield-icon-box,
.lock-icon-box {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 24px;
  margin-bottom: var(--space-md);
}

.shield-icon-box {
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  color: var(--accent);
}

.lock-icon-box {
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  color: var(--accent);
}

.state-title {
  margin: 0 0 var(--space-xs);
  font-size: 20px;
  font-weight: 700;
  color: var(--ink);
}

.state-desc {
  margin: 0 0 var(--space-lg);
  font-size: 14px;
  color: var(--muted);
  line-height: 1.5;
}

.vault-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  text-align: left;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
}

.input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.input-wrap input {
  width: 100%;
  height: 44px;
  padding: 0 44px 0 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  color: var(--ink);
  font-size: 15px;
  transition: border-color 0.15s ease;
}

.input-wrap input:focus {
  outline: none;
  border-color: var(--accent);
  background: var(--surface);
}

.eye-btn {
  position: absolute;
  right: 4px;
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  min-width: var(--touch-min);
  min-height: var(--touch-min);
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.form-error {
  margin: 0;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: var(--danger-soft);
  color: var(--danger);
  font-size: 13px;
  font-weight: 500;
}

.submit-btn {
  height: 46px;
  margin-top: var(--space-xs);
  font-size: 15px;
  font-weight: 600;
}

.form-footer {
  text-align: center;
  margin-top: var(--space-xs);
}

.link-btn {
  background: none;
  border: none;
  color: var(--accent);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  min-height: var(--touch-min);
  padding: 0 10px;
}

.link-btn:hover {
  text-decoration: underline;
}

/* Unlocked View */
.vault-unlocked-surface {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.vault-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--hairline);
}

.vault-meta-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--muted);
  font-weight: 500;
}

.dot-sep {
  opacity: 0.6;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.text-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--muted);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  min-height: var(--touch-min);
  padding: 0 10px;
  border-radius: var(--radius-md);
  transition:
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard);
}

.text-btn:hover {
  background: var(--surface-soft);
  color: var(--ink);
}

/* Empty Box */
.vault-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--space-xl) var(--space-md);
  margin: var(--space-xl) auto;
  max-width: 400px;
}

.empty-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--surface-soft);
  color: var(--muted);
  margin-bottom: var(--space-md);
}

.empty-title {
  margin: 0 0 var(--space-xs);
  font-size: 18px;
  font-weight: 700;
  color: var(--ink);
}

.empty-desc {
  margin: 0 0 var(--space-lg);
  font-size: 14px;
  color: var(--muted);
  line-height: 1.5;
}

/* File List */
.vault-file-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.vault-file-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  padding: 12px 14px;
  border-radius: var(--radius-lg);
  background: var(--surface);
  border: 1px solid var(--hairline-soft, #f1f5f9);
  transition: transform 0.1s ease, box-shadow 0.15s ease;
  cursor: pointer;
}

.vault-file-row:hover {
  background: var(--surface-soft);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.file-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  color: var(--ink);
  flex-shrink: 0;
}

.file-icon-box.image {
  background: color-mix(in srgb, var(--success) 12%, transparent);
  color: var(--success);
}

.file-icon-box.video {
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  color: var(--accent);
}

.file-icon-box.doc {
  background: color-mix(in srgb, var(--primary-cta) 12%, transparent);
  color: var(--primary-cta);
}

.file-icon-box.archive {
  background: color-mix(in srgb, var(--warning) 12%, transparent);
  color: var(--warning);
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  margin: 0 0 4px;
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-subtext {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--muted);
}

.more-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--touch-min);
  min-height: var(--touch-min);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}

.more-btn:hover {
  background: var(--surface-card);
  color: var(--ink);
}

/* Modals & Sheets */
.actions-menu {
  display: flex;
  flex-direction: column;
  padding: var(--space-xs) 0;
}

.sheet-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  min-height: 48px;
  padding: 0 var(--space-md);
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  font-size: 15px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.sheet-row:hover {
  background: var(--surface-soft);
}

.sheet-row.danger {
  color: var(--danger);
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: var(--space-sm) 0 var(--space-lg);
}

.modal-desc {
  margin: 0;
  font-size: 14px;
  color: var(--muted);
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-sm);
  margin-top: var(--space-sm);
}
</style>
