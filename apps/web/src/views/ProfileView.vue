<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, setTokens, ApiError } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { userInitials } from '@/lib/userInitials'
import { useI18n } from '@/lib/i18n'
import { isHeic, convertHeicBlobToJpeg, checkIsHeicBlob } from '@/lib/heic'
import Icon from '@/components/AppIcon.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()
const { t, locale } = useI18n()

const displayName = ref(auth.user?.displayName ?? '')
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordError = ref('')
const error = ref('')
const savingProfile = ref(false)
const changingPassword = ref(false)
const updatingAvatar = ref(false)
const avatarInputRef = ref<HTMLInputElement | null>(null)

const passwordCopy = computed(() =>
  locale.value === 'vi'
    ? {
        confirm: 'Xác nhận mật khẩu mới',
        requirement: 'Mật khẩu mới phải có ít nhất 8 ký tự.',
        mismatch: 'Mật khẩu xác nhận không khớp.',
        tooShort: 'Mật khẩu mới phải có ít nhất 8 ký tự.',
        failed: 'Không thể đổi mật khẩu',
      }
    : {
        confirm: 'Confirm new password',
        requirement: 'New password must be at least 8 characters.',
        mismatch: 'Password confirmation does not match.',
        tooShort: 'New password must be at least 8 characters.',
        failed: 'Password change failed',
      },
)

const passwordSubmitDisabled = computed(
  () =>
    changingPassword.value ||
    !currentPassword.value ||
    !newPassword.value ||
    !confirmPassword.value,
)

const initials = computed(() =>
  userInitials(auth.user?.displayName ?? '', auth.user?.email ?? ''),
)

function triggerAvatarPick() {
  avatarInputRef.value?.click()
}

async function compressImage(file: File, maxSize = 256): Promise<string> {
  const isImage =
    file.type.startsWith('image/') ||
    /\.(jpe?g|png|webp|gif|bmp|heic|heif)$/i.test(file.name) ||
    isHeic(file.name, file.type)
  if (!isImage) {
    throw new Error('Vui lòng chọn file hình ảnh (JPG, PNG, WebP, HEIC)')
  }

  if (file.size > 25 * 1024 * 1024) {
    throw new Error('Kích thước ảnh quá lớn (tối đa 25MB)')
  }

  let sourceBlob: Blob = file
  const isHeicImage = isHeic(file.name, file.type) || (await checkIsHeicBlob(file))
  if (isHeicImage) {
    try {
      sourceBlob = await convertHeicBlobToJpeg(file, 0.9)
    } catch (err) {
      console.error('HEIC conversion failed:', err)
      throw new Error('Không thể giải mã file ảnh HEIC. Vui lòng chọn ảnh JPG, PNG hoặc thử lại.')
    }
  }

  return new Promise((resolve, reject) => {
    let objectUrl = ''
    try {
      objectUrl = URL.createObjectURL(sourceBlob)
    } catch {
      // Fallback to FileReader if createObjectURL fails
      const reader = new FileReader()
      reader.onerror = () => reject(new Error('Không thể đọc file ảnh từ thiết bị'))
      reader.onload = () => {
        const result = reader.result
        if (typeof result !== 'string' || !result) {
          reject(new Error('Dữ liệu ảnh không hợp lệ'))
          return
        }
        loadImageAndCompress(result, null)
      }
      reader.readAsDataURL(sourceBlob)
      return
    }

    loadImageAndCompress(objectUrl, objectUrl)

    function loadImageAndCompress(src: string, urlToRevoke: string | null) {
      const img = new Image()
      img.onload = () => {
        if (urlToRevoke) {
          URL.revokeObjectURL(urlToRevoke)
        }
        let width = img.naturalWidth || img.width
        let height = img.naturalHeight || img.height
        if (!width || !height) {
          reject(new Error('Không thể xác định kích thước ảnh'))
          return
        }

        if (width > height) {
          if (width > maxSize) {
            height = Math.round((height * maxSize) / width)
            width = maxSize
          }
        } else {
          if (height > maxSize) {
            width = Math.round((width * maxSize) / height)
            height = maxSize
          }
        }

        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          reject(new Error('Không thể xử lý đồ họa ảnh'))
          return
        }

        ctx.drawImage(img, 0, 0, width, height)

        try {
          const webp = canvas.toDataURL('image/webp', 0.85)
          if (webp && webp.startsWith('data:image/webp')) {
            resolve(webp)
            return
          }
        } catch {
          // fallback to jpeg
        }

        try {
          const jpeg = canvas.toDataURL('image/jpeg', 0.85)
          if (jpeg && jpeg.startsWith('data:image/jpeg')) {
            resolve(jpeg)
            return
          }
        } catch {
          // fallback
        }

        resolve(canvas.toDataURL())
      }

      img.onerror = () => {
        if (urlToRevoke) {
          URL.revokeObjectURL(urlToRevoke)
        }
        reject(new Error('Không thể giải mã định dạng ảnh này. Vui lòng chọn ảnh JPG, PNG hoặc WebP.'))
      }

      img.src = src
    }
  })
}

async function handleAvatarSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  error.value = ''
  updatingAvatar.value = true
  try {
    const dataUrl = await compressImage(file, 256)
    if (!dataUrl) {
      throw new Error('Không thể xử lý ảnh')
    }
    await auth.updateAvatar(dataUrl)
    ui.showToast('Đã cập nhật ảnh đại diện', 'success')
  } catch (err) {
    const message =
      err instanceof ApiError
        ? formatApiError(err, 'Không thể tải ảnh lên')
        : (err as Error)?.message || 'Không thể tải ảnh lên'
    error.value = message
    ui.showToast(message, 'error')
  } finally {
    updatingAvatar.value = false
    target.value = ''
  }
}

async function removeAvatar() {
  error.value = ''
  updatingAvatar.value = true
  try {
    await auth.updateAvatar('')
    ui.showToast('Đã xóa ảnh đại diện', 'success')
  } catch (err) {
    const message =
      err instanceof ApiError
        ? formatApiError(err, 'Không thể xóa ảnh')
        : (err as Error)?.message || 'Không thể xóa ảnh'
    error.value = message
    ui.showToast(message, 'error')
  } finally {
    updatingAvatar.value = false
  }
}

async function saveProfile() {
  error.value = ''
  savingProfile.value = true
  try {
    await api<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify({ displayName: displayName.value.trim() }),
    })
    await auth.loadMe()
    ui.showToast(t.value.saveProfile, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Save failed')
  } finally {
    savingProfile.value = false
  }
}

async function changePassword() {
  if (changingPassword.value) return

  passwordError.value = ''
  error.value = ''
  if (newPassword.value.length < 8) {
    passwordError.value = passwordCopy.value.tooShort
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = passwordCopy.value.mismatch
    return
  }

  changingPassword.value = true
  try {
    const tokens = await api<{ accessToken: string; refreshToken: string }>('/users/me/password', {
      method: 'POST',
      body: JSON.stringify({
        currentPassword: currentPassword.value,
        newPassword: newPassword.value,
      }),
    })
    setTokens(tokens.accessToken, tokens.refreshToken)
    await auth.loadMe()
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    ui.showToast(t.value.changePassword, 'success')
  } catch (e) {
    passwordError.value = formatApiError(e, passwordCopy.value.failed)
  } finally {
    changingPassword.value = false
  }
}

async function logout() {
  if (!(await ui.confirmLogout())) return
  await auth.logout()
  window.location.href = '/login'
}
</script>

<template>
  <div class="profile-page">
    <h1 class="page-title desktop-only">{{ t.profile }}</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <section class="card section identity" aria-labelledby="profile-identity-heading">
      <div class="identity-row">
        <div class="avatar-uploader">
          <img
            v-if="auth.user?.avatarUrl"
            :src="auth.user.avatarUrl"
            :alt="auth.user.displayName || 'Avatar'"
            class="avatar-image"
          />
          <span v-else class="avatar" aria-hidden="true">{{ initials }}</span>
          <button
            type="button"
            class="avatar-action-btn"
            :title="'Đổi ảnh đại diện'"
            :aria-label="'Đổi ảnh đại diện'"
            :disabled="updatingAvatar"
            @click="triggerAvatarPick"
          >
            <Icon name="camera" :size="18" />
          </button>
          <input
            ref="avatarInputRef"
            type="file"
            accept="image/png,image/jpeg,image/webp,image/*,.heic,.heif,.HEIC,.HEIF"
            class="sr-only"
            @change="handleAvatarSelected"
          />
        </div>
        <div class="identity-copy">
          <h2 id="profile-identity-heading" class="identity-name">
            {{ auth.user?.displayName || 'Your account' }}
          </h2>
          <p class="identity-email">{{ auth.user?.email }}</p>
          <button
            v-if="auth.user?.avatarUrl"
            type="button"
            class="remove-avatar-btn"
            :disabled="updatingAvatar"
            @click="removeAvatar"
          >
            Xóa ảnh đại diện
          </button>
        </div>
      </div>
      <label class="field">
        <span>{{ t.displayName }}</span>
        <input v-model="displayName" autocomplete="nickname" maxlength="80" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="savingProfile" @click="saveProfile">
        {{ savingProfile ? t.saving : t.saveProfile }}
      </button>
    </section>

    <section class="card section" aria-labelledby="profile-password-heading">
      <h2 id="profile-password-heading" class="section-title">{{ t.password }}</h2>
      <div class="field">
        <label class="field-label" for="profile-current-password">{{ t.currentPassword }}</label>
        <PasswordInput
          id="profile-current-password"
          v-model="currentPassword"
          name="current-password"
          autocomplete="current-password"
          :disabled="changingPassword"
          required
        />
      </div>
      <div class="field">
        <label class="field-label" for="profile-new-password">{{ t.newPassword }}</label>
        <PasswordInput
          id="profile-new-password"
          v-model="newPassword"
          name="new-password"
          autocomplete="new-password"
          :disabled="changingPassword"
          :aria-invalid="Boolean(passwordError)"
          :aria-describedby="passwordError ? 'profile-password-requirement profile-password-error' : 'profile-password-requirement'"
          :minlength="8"
          required
        />
      </div>
      <p id="profile-password-requirement" class="field-hint">{{ passwordCopy.requirement }}</p>
      <div class="field">
        <label class="field-label" for="profile-confirm-password">{{ passwordCopy.confirm }}</label>
        <PasswordInput
          id="profile-confirm-password"
          v-model="confirmPassword"
          name="confirm-password"
          autocomplete="new-password"
          :disabled="changingPassword"
          :aria-invalid="Boolean(passwordError)"
          :aria-describedby="passwordError ? 'profile-password-error' : undefined"
          :minlength="8"
          required
        />
      </div>
      <p
        v-if="passwordError"
        id="profile-password-error"
        class="error password-error"
        role="alert"
        aria-live="assertive"
      >
        {{ passwordError }}
      </p>
      <button
        class="btn ink save-btn"
        type="button"
        :disabled="passwordSubmitDisabled"
        :aria-busy="changingPassword ? 'true' : undefined"
        @click="changePassword"
      >
        {{ changingPassword ? t.changing : t.changePassword }}
      </button>
    </section>

    <section class="card section">
      <button class="btn danger block" type="button" @click="logout">{{ t.logOut }}</button>
    </section>
  </div>
</template>

<style scoped>
.section {
  margin-bottom: var(--space-md);
}

.identity-row {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  margin-bottom: var(--space-md);
}

.avatar-uploader {
  position: relative;
  width: 64px;
  height: 64px;
  flex-shrink: 0;
}

.avatar-image {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-pill);
  object-fit: cover;
  border: 2px solid var(--surface-card);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.avatar-uploader .avatar {
  width: 64px;
  height: 64px;
  font-size: 1.25rem;
}

.avatar-action-btn {
  position: absolute;
  right: -12px;
  bottom: -12px;
  width: var(--touch-min);
  height: var(--touch-min);
  border-radius: var(--radius-pill);
  background: var(--accent);
  color: var(--on-accent, #ffffff);
  border: 2px solid var(--surface-card);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  transition:
    transform var(--duration-short) var(--ease-standard),
    opacity var(--duration-short) var(--ease-standard);
}

.avatar-action-btn:not(:disabled):hover {
  transform: scale(1.06);
}

.avatar-action-btn:focus-visible,
.remove-avatar-btn:focus-visible {
  outline: 3px solid var(--accent);
  outline-offset: 2px;
}

.avatar-action-btn:disabled,
.remove-avatar-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.remove-avatar-btn {
  display: inline-flex;
  align-items: center;
  min-height: var(--touch-min);
  margin: 2px 0 -6px calc(-1 * var(--space-xs));
  padding: 0 var(--space-xs);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--danger, #ef4444);
  font-size: 0.8125rem;
  line-height: 1.25;
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.field-hint {
  margin: calc(-1 * var(--space-xs)) 0 var(--space-sm);
  color: var(--muted);
  font-size: 0.8125rem;
  line-height: 1.4;
}

.password-error {
  margin: calc(-1 * var(--space-xs)) 0 var(--space-sm);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  color: var(--accent-hover);
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.identity-copy {
  min-width: 0;
  flex: 1;
}

.identity-name {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.identity-email {
  margin: 2px 0 0;
  font-size: 14px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.save-btn {
  width: 100%;
}

.desktop-only {
  display: none;
}

@media (max-width: 359px) {
  .identity-row {
    gap: var(--space-sm);
  }
}

@media (min-width: 768px) {
  .save-btn {
    width: auto;
  }

  .desktop-only {
    display: block;
  }
}

@media (prefers-reduced-motion: reduce) {
  .avatar-action-btn {
    transition: none;
  }
}
</style>
