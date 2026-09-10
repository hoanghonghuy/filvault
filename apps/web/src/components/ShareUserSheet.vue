<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { api } from '@/api/client'
import { formatShareUserError } from '@/api/errors'
import type { CreateUserShareResponse, ShareResourceType } from '@/api/types'
import BottomSheet from '@/components/BottomSheet.vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/lib/i18n'

const props = withDefaults(
  defineProps<{
    open: boolean
    name: string
    resourceId: string | null
    resourceType?: ShareResourceType
  }>(),
  {
    resourceType: 'file',
  },
)

const emit = defineEmits<{
  close: []
}>()

type Phase = 'form' | 'submitting' | 'success'

const auth = useAuthStore()
const { locale } = useI18n()
const email = ref('')
const phase = ref<Phase>('form')
const error = ref('')
const successInvited = ref(false)
const successRecipientLabel = ref('')

const emailInputRef = ref<HTMLInputElement | null>(null)
const errorRef = ref<HTMLElement | null>(null)
const successRef = ref<HTMLElement | null>(null)

const copy = computed(() =>
  locale.value === 'vi'
    ? {
        permission: 'Người nhận sẽ có quyền đọc mục này. Email chưa đăng ký sẽ nhận lời mời tạo tài khoản.',
        email: 'Email',
        invalidEmail: 'Nhập địa chỉ email hợp lệ.',
        selfShare: 'Bạn không thể chia sẻ cho chính mình.',
        conflict: 'Mục này đã được chia sẻ cho người dùng đó.',
        notFound: 'Không tìm thấy mục để chia sẻ.',
        network: 'Không thể kết nối máy chủ. Hãy thử lại sau ít phút.',
        fallback: 'Không thể chia sẻ mục này.',
        sharing: 'Đang chia sẻ…',
        share: 'Chia sẻ',
        close: 'Đóng',
        done: 'Xong',
        shareAnother: 'Chia sẻ cho người khác',
        invitedPrefix: 'Đã gửi lời mời tới',
        invitedSuffix: 'Họ sẽ có quyền đọc sau khi đăng ký bằng email này.',
        sharedPrefix: 'Đã chia sẻ với',
        sharedSuffix: 'Họ có thể xem và tải mục này với quyền đọc.',
      }
    : {
        permission: "They'll get read access to this item. Unknown emails receive an invite to sign up.",
        email: 'Email',
        invalidEmail: 'Enter a valid email address.',
        selfShare: "You can't share with yourself.",
        conflict: 'This item is already shared with that user.',
        notFound: 'Item not found.',
        network: "Can't reach the server. Try again in a moment.",
        fallback: 'Could not share this item.',
        sharing: 'Sharing…',
        share: 'Share',
        close: 'Close',
        done: 'Done',
        shareAnother: 'Share with another',
        invitedPrefix: 'Invitation sent to',
        invitedSuffix: "They'll get read access once they sign up with that email.",
        sharedPrefix: 'Shared with',
        sharedSuffix: 'They can view and download this item with read access.',
      },
)

const trimmedEmail = computed(() => email.value.trim())
const canSubmit = computed(() => phase.value === 'form' && trimmedEmail.value.length > 0)

function resetState() {
  email.value = ''
  phase.value = 'form'
  error.value = ''
  successInvited.value = false
  successRecipientLabel.value = ''
}

function isValidEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
}

function isSelfShare(value: string): boolean {
  const ownEmail = auth.user?.email?.trim().toLowerCase()
  return Boolean(ownEmail && ownEmail === value.toLowerCase())
}

async function focusRef(target: { value: HTMLElement | null }) {
  await nextTick()
  target.value?.focus()
}

async function onSubmit() {
  if (!canSubmit.value || !props.resourceId) return

  const value = trimmedEmail.value
  error.value = ''

  if (!isValidEmail(value)) {
    error.value = copy.value.invalidEmail
    await focusRef(errorRef)
    return
  }

  if (isSelfShare(value)) {
    error.value = copy.value.selfShare
    await focusRef(errorRef)
    return
  }

  phase.value = 'submitting'
  try {
    const res = await api<CreateUserShareResponse>('/shares', {
      method: 'POST',
      body: JSON.stringify({
        resourceType: props.resourceType,
        resourceId: props.resourceId,
        email: value,
      }),
    })

    successInvited.value = Boolean(res.invited)
    if (res.invited) {
      successRecipientLabel.value = value
    } else {
      const recipient = res.recipient
      successRecipientLabel.value = recipient?.displayName?.trim() || recipient?.email || value
    }

    phase.value = 'success'
    await focusRef(successRef)
  } catch (e) {
    phase.value = 'form'
    error.value = formatShareUserError(e, copy.value.fallback, {
      conflict: copy.value.conflict,
      notFound: copy.value.notFound,
      validation: copy.value.invalidEmail,
      network: copy.value.network,
    })
    await focusRef(errorRef)
  }
}

function onShareAnother() {
  email.value = ''
  phase.value = 'form'
  error.value = ''
  successInvited.value = false
  successRecipientLabel.value = ''
  void focusRef(emailInputRef)
}

function onClose() {
  emit('close')
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      resetState()
      void nextTick(() => {
        if (phase.value === 'form') emailInputRef.value?.focus()
      })
    }
  },
)
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="onClose">
    <div v-if="phase === 'success'" class="success-state">
      <p ref="successRef" class="success-message" tabindex="-1" role="status" aria-live="polite">
        <template v-if="successInvited">
          {{ copy.invitedPrefix }} <strong>{{ successRecipientLabel }}</strong>. {{ copy.invitedSuffix }}
        </template>
        <template v-else>
          {{ copy.sharedPrefix }} <strong>{{ successRecipientLabel }}</strong>. {{ copy.sharedSuffix }}
        </template>
      </p>
      <div class="actions">
        <button type="button" class="btn block ink" @click="onClose">{{ copy.done }}</button>
        <button type="button" class="btn block ghost" @click="onShareAnother">{{ copy.shareAnother }}</button>
      </div>
    </div>

    <form v-else class="share-form" novalidate @submit.prevent="onSubmit">
      <p class="hint">{{ copy.permission }}</p>

      <label class="field">
        <span class="field-label">{{ copy.email }}</span>
        <input
          ref="emailInputRef"
          v-model="email"
          type="email"
          name="share-email"
          autocomplete="email"
          inputmode="email"
          enterkeyhint="send"
          placeholder="name@example.com"
          :disabled="phase === 'submitting'"
          :aria-invalid="error ? 'true' : undefined"
          :aria-describedby="error ? 'share-user-error' : undefined"
        />
      </label>

      <p v-if="error" id="share-user-error" ref="errorRef" class="error" role="alert" tabindex="-1">
        {{ error }}
      </p>

      <div class="actions">
        <button
          type="submit"
          class="btn block ink"
          :disabled="phase === 'submitting' || !trimmedEmail"
          :aria-busy="phase === 'submitting' ? 'true' : undefined"
        >
          {{ phase === 'submitting' ? copy.sharing : copy.share }}
        </button>
        <button type="button" class="btn block ghost" :disabled="phase === 'submitting'" @click="onClose">
          {{ copy.close }}
        </button>
      </div>
    </form>
  </BottomSheet>
</template>

<style scoped>
.hint {
  margin: 0 0 var(--space-md);
  font-size: 0.8125rem;
  line-height: 1.45;
  color: var(--muted);
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  margin-bottom: var(--space-md);
}

.field-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--muted);
}

.actions,
.success-state {
  display: flex;
  flex-direction: column;
}

.success-message {
  margin: 0 0 var(--space-md);
  font-size: 0.9375rem;
  line-height: 1.5;
  color: var(--ink);
}

.success-message:focus {
  outline: none;
}

.success-message:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}
</style>
