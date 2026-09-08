<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useVaultStore } from '@/stores/vault'
import { useUiStore } from '@/stores/ui'
import { useI18n } from '@/lib/i18n'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { mimeIcon } from '@/lib/mimeIcon'
import { isHeic, getHeicDisplayUrl } from '@/lib/heic'
import Icon from '@/components/AppIcon.vue'
import BottomSheet from '@/components/BottomSheet.vue'
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
const showChangePins = ref(false)
const changePinError = ref('')
const changingPin = ref(false)

// Reset PIN modal (using account password)
const resetPinOpen = ref(false)
const accountPassword = ref('')
const resetNewPin = ref('')
const confirmResetNewPin = ref('')
const showResetPins = ref(false)
const resetPinError = ref('')
const resettingPin = ref(false)

// Actions on a vault file
const selectedFile = ref<VaultFile | null>(null)
const fileActionOpen = ref(false)

// In-app preview
const previewOpen = ref(false)
const previewFile = ref<{ id: string; name: string; mimeType: string; url: string } | null>(null)
const previewDisplayUrl = ref('')
const previewHeicLoading = ref(false)

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
          ui.showToast('Kho cá nhân đã tự động khóa', 'info')
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
  } catch (e) {
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
    setupError.value = formatApiError(e, 'Không thể thiết lập mật khẩu kho')
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
    unlockError.value = formatApiError(e, 'Mật khẩu kho không chính xác')
  }
}

function handleLockNow() {
  vault.lock()
  ui.showToast('Kho cá nhân đã được khóa', 'info')
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
    changePinError.value = formatApiError(e, 'Không thể đổi mật khẩu kho')
  } finally {
    changingPin.value = false
  }
}

async function handleResetPin() {
  resetPinError.value = ''
  if (!accountPassword.value) {
    resetPinError.value = 'Vui lòng nhập mật khẩu tài khoản'
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
    resetPinError.value = formatApiError(e, 'Mật khẩu tài khoản không đúng hoặc có lỗi xảy ra')
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
    if (isHeic(file.name, file.mimeType)) {
      previewHeicLoading.value = true
      previewDisplayUrl.value = ''
      try {
        const resolved = await getHeicDisplayUrl(out.downloadUrl)
        previewDisplayUrl.value = resolved
      } catch {
        previewDisplayUrl.value = out.downloadUrl
      } finally {
        previewHeicLoading.value = false
      }
    } else {
      previewDisplayUrl.value = out.downloadUrl
    }
  } catch (e) {
    ui.showToast(formatApiError(e, 'Không thể xem trước tệp'), 'error')
  }
}

function closePreview() {
  previewOpen.value = false
  previewFile.value = null
  previewDisplayUrl.value = ''
  previewHeicLoading.value = false
}

async function removeSelectedFileFromVault() {
  if (!selectedFile.value) return
  const f = selectedFile.value
  fileActionOpen.value = false
  try {
    await vault.removeFromVault([f.id])
    ui.showToast(t.value.vaultMoveOutSuccess, 'success')
  } catch (e) {
    ui.showToast(formatApiError(e, 'Không thể chuyển tệp ra ngoài'), 'error')
  }
}

async function deleteSelectedFile() {
  if (!selectedFile.value) return
  const f = selectedFile.value
  fileActionOpen.value = false
  const ok = await ui.confirm({
    title: 'Xóa tệp khỏi kho cá nhân?',
    message: `"${f.name}" sẽ được chuyển vào Thùng rác.`,
    confirmLabel: 'Xóa tệp',
    danger: true,
  })
  if (!ok) return

  try {
    await api(`/files/${f.id}`, { method: 'DELETE' })
    await vault.loadFiles()
    ui.showToast('Đã xóa tệp', 'success')
  } catch (e) {
    ui.showToast(formatApiError(e, 'Xóa tệp thất bại'), 'error')
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
      <button type="button" class="back-btn" :title="t.back" aria-label="Quay lại" @click="router.back()">
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
          aria-label="Khóa kho ngay"
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
              placeholder="Nhập mã PIN hoặc mật khẩu (tối thiểu 4 ký tự)"
              maxlength="32"
              required
              autofocus
            />
            <button
              type="button"
              class="eye-btn"
              :aria-label="showSetupPin ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'"
              @click="showSetupPin = !showSetupPin"
            >
              <Icon :name="showSetupPin ? 'eye' : 'eye'" :size="18" />
            </button>
          </div>
        </label>

        <label class="field">
          <span>{{ t.vaultPinConfirm }}</span>
          <div class="input-wrap">
            <input
              v-model="confirmSetupPin"
              :type="showSetupPin ? 'text' : 'password'"
              placeholder="Nhập lại mật khẩu"
              maxlength="32"
              required
            />
          </div>
        </label>

        <button type="submit" class="btn primary submit-btn" :disabled="vault.loading">
          {{ vault.loading ? 'Đang tạo…' : t.vaultCreateAndEnter }}
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
              placeholder="Nhập mật khẩu kho cá nhân"
              maxlength="32"
              required
              autofocus
            />
            <button
              type="button"
              class="eye-btn"
              :aria-label="showUnlockPin ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'"
              @click="showUnlockPin = !showUnlockPin"
            >
              <Icon :name="showUnlockPin ? 'eye' : 'eye'" :size="18" />
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
          <span class="file-count">{{ vault.files.length }} tệp</span>
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
          <span>Đi tới Tệp của tôi</span>
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
            aria-label="Thao tác tệp"
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
      :title="selectedFile?.name || 'Thao tác tệp'"
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
            {{ changingPin ? 'Đang lưu…' : 'Lưu mật khẩu mới' }}
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
            {{ resettingPin ? 'Đang xử lý…' : t.vaultResetAndEnter }}
          </button>
        </div>
      </form>
    </BottomSheet>

    <!-- Preview Modal -->
    <div v-if="previewOpen && previewFile" class="preview-backdrop" @click="closePreview">
      <div class="preview-dialog" @click.stop>
        <header class="preview-header">
          <span class="preview-filename">
            {{ previewFile.name }}
            <span v-if="isHeic(previewFile.name, previewFile.mimeType)" class="heic-tag">HEIC</span>
          </span>
          <button type="button" class="close-btn" aria-label="Đóng" @click="closePreview">
            <Icon name="close" :size="20" />
          </button>
        </header>

        <div class="preview-body">
          <div v-if="previewHeicLoading" class="preview-heic-loading">
            <div class="heic-spinner" />
            <p>{{ t.heicConverting }}</p>
          </div>
          <img
            v-else-if="previewFile.mimeType.startsWith('image/')"
            :src="previewDisplayUrl || previewFile.url"
            :alt="previewFile.name"
            class="preview-img"
          />
          <video
            v-else-if="previewFile.mimeType.startsWith('video/')"
            :src="previewFile.url"
            controls
            autoplay
            class="preview-video"
          />
          <div v-else class="preview-fallback">
            <Icon :name="mimeIcon(previewFile.mimeType)" :size="48" />
            <p>{{ t.previewUnavailable }}</p>
            <button type="button" class="btn primary" @click="downloadFile(previewFile.id)">
              <Icon name="download" :size="18" />
              <span>{{ t.download }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
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
  width: 38px;
  height: 38px;
  border: none;
  border-radius: var(--radius-full, 9999px);
  background: var(--surface-soft);
  color: var(--ink);
  cursor: pointer;
  transition: background 0.15s ease;
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
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

.status-badge.locked {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
}

.lock-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--ink);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
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
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(79, 70, 229, 0.25));
  color: #6366f1;
}

.lock-icon-box {
  background: linear-gradient(135deg, rgba(13, 148, 136, 0.15), rgba(15, 118, 110, 0.25));
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
  right: 10px;
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.form-error {
  margin: 0;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
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
  padding: 6px 10px;
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
  padding: 6px 10px;
  border-radius: var(--radius-md);
  transition: all 0.15s ease;
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
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

.file-icon-box.video {
  background: rgba(139, 92, 246, 0.12);
  color: #8b5cf6;
}

.file-icon-box.doc {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
}

.file-icon-box.archive {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
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
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
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
  color: #ef4444;
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

/* Preview Backdrop */
.preview-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-md);
}

.preview-dialog {
  display: flex;
  flex-direction: column;
  max-width: 90vw;
  max-height: 90vh;
  background: var(--surface);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--hairline);
}

.preview-filename {
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 80vw;
}

.close-btn {
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 4px;
}

.preview-body {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-md);
  overflow: auto;
}

.preview-img {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
  border-radius: var(--radius-md);
}

.preview-video {
  max-width: 100%;
  max-height: 70vh;
  border-radius: var(--radius-md);
}

.preview-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-md);
  padding: var(--space-xl);
  color: var(--muted);
}

.heic-tag {
  display: inline-block;
  padding: 1px 6px;
  margin-left: 6px;
  border-radius: var(--radius-pill);
  background: var(--accent);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  vertical-align: middle;
}

.preview-heic-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xl);
  color: var(--fg-soft, var(--ink));
  font-size: 14px;
}

.heic-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(128, 128, 128, 0.2);
  border-top-color: var(--accent, #3b82f6);
  border-radius: 50%;
  animation: heic-spin 0.8s linear infinite;
}

@keyframes heic-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
