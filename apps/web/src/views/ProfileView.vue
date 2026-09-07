<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, setTokens } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { userInitials } from '@/lib/userInitials'
import { useI18n } from '@/lib/i18n'
import Icon from '@/components/AppIcon.vue'
import type { User } from '@/api/types'

const auth = useAuthStore()
const ui = useUiStore()
const { t } = useI18n()

const displayName = ref(auth.user?.displayName ?? '')
const currentPassword = ref('')
const newPassword = ref('')
const error = ref('')
const savingProfile = ref(false)
const changingPassword = ref(false)
const updatingAvatar = ref(false)
const avatarInputRef = ref<HTMLInputElement | null>(null)

const initials = computed(() =>
  userInitials(auth.user?.displayName ?? '', auth.user?.email ?? ''),
)

function triggerAvatarPick() {
  avatarInputRef.value?.click()
}

function compressImage(file: File, maxSize = 256): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    const reader = new FileReader()
    reader.onload = (e) => {
      img.src = e.target?.result as string
    }
    reader.onerror = reject
    img.onload = () => {
      const canvas = document.createElement('canvas')
      let width = img.width
      let height = img.height
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
      canvas.width = width
      canvas.height = height
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        resolve(img.src)
        return
      }
      ctx.drawImage(img, 0, 0, width, height)
      resolve(canvas.toDataURL('image/webp', 0.85))
    }
    img.onerror = reject
    reader.readAsDataURL(file)
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
    await auth.updateAvatar(dataUrl)
    ui.showToast('Đã cập nhật ảnh đại diện', 'success')
  } catch (err) {
    error.value = formatApiError(err, 'Không thể tải ảnh lên')
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
    error.value = formatApiError(err, 'Không thể xóa ảnh')
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
  error.value = ''
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
    ui.showToast(t.value.changePassword, 'success')
  } catch (e) {
    error.value = formatApiError(e, 'Password change failed')
  } finally {
    changingPassword.value = false
  }
}

async function logout() {
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
            <Icon name="camera" :size="15" />
          </button>
          <input
            ref="avatarInputRef"
            type="file"
            accept="image/*"
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
      <label class="field">
        <span>{{ t.currentPassword }}</span>
        <input v-model="currentPassword" type="password" autocomplete="current-password" />
      </label>
      <label class="field">
        <span>{{ t.newPassword }}</span>
        <input v-model="newPassword" type="password" minlength="8" autocomplete="new-password" />
      </label>
      <button class="btn ink save-btn" type="button" :disabled="changingPassword" @click="changePassword">
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
  right: -4px;
  bottom: -4px;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-pill);
  background: var(--accent);
  color: var(--on-accent, #ffffff);
  border: 2px solid var(--surface-card);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  transition: transform var(--duration-short) var(--ease-standard);
}

.avatar-action-btn:hover {
  transform: scale(1.1);
}

.remove-avatar-btn {
  display: inline-block;
  margin-top: 4px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--danger, #ef4444);
  font-size: 12px;
  cursor: pointer;
  text-decoration: underline;
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

@media (min-width: 768px) {
  .save-btn {
    width: auto;
  }

  .desktop-only {
    display: block;
  }
}
</style>
