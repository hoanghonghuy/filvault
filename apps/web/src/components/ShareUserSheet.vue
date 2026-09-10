<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { api } from '@/api/client'
import { formatShareUserError } from '@/api/errors'
import type { CreateUserShareResponse, ShareResourceType } from '@/api/types'
import BottomSheet from '@/components/BottomSheet.vue'
import { useAuthStore } from '@/stores/auth'

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
  shared: [payload: { invited: boolean; email: string }]
}>()

type Phase = 'form' | 'submitting' | 'success'

const auth = useAuthStore()
const email = ref('')
const phase = ref<Phase>('form')
const error = ref('')
const successInvited = ref(false)
const successRecipientLabel = ref('')

const emailInputRef = ref<HTMLInputElement | null>(null)
const errorRef = ref<HTMLElement | null>(null)
const successRef = ref<HTMLElement | null>(null)
const doneButtonRef = ref<HTMLButtonElement | null>(null)

const trimmedEmail = computed(() => email.value.trim())
const canSubmit = computed(() => phase.value === 'form' && trimmedEmail.value.length > 0)

const permissionHint =
  "They'll get read access to this item. Unknown emails receive an invite to sign up."

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
    error.value = 'Enter a valid email address.'
    await focusRef(errorRef)
    return
  }

  if (isSelfShare(value)) {
    error.value = "You can't share with yourself."
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
    emit('shared', { invited: successInvited.value, email: value })
    await focusRef(successRef)
  } catch (e) {
    phase.value = 'form'
    error.value = formatShareUserError(e, 'Could not share')
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
        if (phase.value === 'form') {
          emailInputRef.value?.focus()
        }
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
          Invitation sent to <strong>{{ successRecipientLabel }}</strong>. They'll get read access once
          they sign up with that email.
        </template>
        <template v-else>
          Shared with <strong>{{ successRecipientLabel }}</strong>. They can view and download this
          item with read access.
        </template>
      </p>
      <div class="actions">
        <button ref="doneButtonRef" type="button" class="btn block ink" @click="onClose">Done</button>
        <button type="button" class="btn block ghost" @click="onShareAnother">Share with another</button>
      </div>
    </div>

    <form v-else class="share-form" novalidate @submit.prevent="onSubmit">
      <p class="hint">{{ permissionHint }}</p>

      <label class="field">
        <span class="field-label">Email</span>
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

      <p
        v-if="error"
        id="share-user-error"
        ref="errorRef"
        class="error"
        role="alert"
        tabindex="-1"
      >
        {{ error }}
      </p>

      <div class="actions">
        <button
          type="submit"
          class="btn block ink"
          :disabled="phase === 'submitting' || !trimmedEmail"
          :aria-busy="phase === 'submitting' ? 'true' : undefined"
        >
          {{ phase === 'submitting' ? 'Sharing…' : 'Share' }}
        </button>
        <button type="button" class="btn block ghost" :disabled="phase === 'submitting'" @click="onClose">
          Close
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

.actions {
  display: flex;
  flex-direction: column;
}

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
