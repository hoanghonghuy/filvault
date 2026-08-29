<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, uploadToPresigned, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import LoadingSkeletonThread from '@/components/LoadingSkeletonThread.vue'
import UploadProgress from '@/components/UploadProgress.vue'
import type { ChatAttachment, ChatConversation, ChatMessage, DownloadURL, UploadSession } from '@/api/types'

const router = useRouter()
const ui = useUiStore()
const chatStore = useChatStore()
const auth = useAuthStore()

const conversations = ref<ChatConversation[]>([])
const selectedId = ref<string | null>(null)
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const searchQuery = ref('')
const searchResults = ref<ChatMessage[] | null>(null)
const media = ref<ChatAttachment[]>([])
const loading = ref(false)
const sending = ref(false)
const error = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const pendingMessage = ref<ChatMessage | null>(null)
const showJump = ref(false)
const threadBodyRef = ref<HTMLElement | null>(null)
const composerRef = ref<HTMLTextAreaElement | null>(null)
const hasMore = ref(false)
const nextBefore = ref<string | null>(null)
const loadingOlder = ref(false)
const lightboxOpen = ref(false)
const lightboxName = ref('')
const lightboxMime = ref('')
const lightboxUrl = ref('')
const lightboxFileId = ref('')
const loadingThread = ref(true)

const selectedConversation = computed(() => conversations.value.find((c) => c.id === selectedId.value) ?? null)
const threadSearchOpen = ref(false)
const uploadProgress = ref<number | null>(null)
const isMobile = ref(false)
type AttachmentUploadStatus = 'queued' | 'uploading' | 'failed' | 'canceled'
type AttachmentUpload = {
  id: string
  file: File
  conversationId: string
  body: string
  status: AttachmentUploadStatus
  error?: string
  controller: AbortController
}
const attachmentQueue = ref<AttachmentUpload[]>([])
const pendingMessageError = ref('')
let activeSelection = 0
function isMobileViewport() {
  return isMobile.value
}
const railFilter = ref('')
const filteredConversations = computed(() => {
  const q = railFilter.value.trim().toLowerCase()
  if (!q) return conversations.value
  return conversations.value.filter((c) =>
    `${c.title} ${c.peer?.name ?? ''} ${c.peer?.email ?? ''}`.toLowerCase().includes(q),
  )
})
const visibleMessages = computed(() => searchResults.value ?? messages.value)

function conversationTitle(conversation: ChatConversation | null): string {
  if (!conversation) return ''
  return conversation.peer?.name || conversation.peer?.email || conversation.title || 'Untitled chat'
}

function avatarClass(title: string): string {
  const palette = ['a', 'b', 'c', 'd', 'e', 'f']
  let sum = 0
  for (const ch of title) sum += ch.codePointAt(0) ?? 0
  return `avatar-color-${palette[sum % palette.length]}`
}

function formatRelativeDay(iso: string): string {
  const date = new Date(iso)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  if (date.toDateString() === today.toDateString()) return formatTime(iso)
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday'
  return date.toLocaleDateString(undefined, { day: 'numeric', month: 'numeric' })
}
const inThread = computed({
  get: () => Boolean(selectedId.value),
  set: (value: boolean) => {
    if (!value) {
      selectedId.value = null
      searchQuery.value = ''
      searchResults.value = null
    }
  },
})

function backToRail() {
  inThread.value = false
}

function updateViewport() {
  isMobile.value = window.matchMedia('(max-width: 767px)').matches
}

function promptNewConversation() {
  void ui.prompt({
    title: 'New direct chat',
    label: 'Verified email address',
    confirmLabel: 'Open chat',
  }).then((email) => {
    if (!email?.trim()) return
    void createDirectConversation(email.trim())
  })
}

function backToVault() {
  void router.push('/')
}

async function loadConversations() {
  error.value = ''
  loading.value = true
  try {
    const out = await api<{ conversations: ChatConversation[] }>('/chat/conversations?includePreview=true')
    conversations.value = out.conversations
    if (!selectedId.value && out.conversations[0] && !isMobileViewport()) {
      await selectConversation(out.conversations[0].id)
    }
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load chats')
  } finally {
    loading.value = false
  }
}

async function createDirectConversation(email: string) {
  error.value = ''
  try {
    const conversation = await chatStore.createDirectConversation(email)
    conversations.value = [...chatStore.conversations]
    await selectConversation(conversation.id)
  } catch (e) {
    error.value = formatApiError(e, 'Could not open direct chat')
  }
}

async function selectConversation(id: string) {
  const selection = activeSelection + 1
  activeSelection = selection
  selectedId.value = id
  void chatStore.openConversation(id)
  searchQuery.value = ''
  searchResults.value = null
  threadSearchOpen.value = false
  loadingThread.value = true
  await Promise.all([loadMessages(id), loadMedia(id)])
  if (selection !== activeSelection || selectedId.value !== id) return
  await nextTick()
  scrollToLatest()
  focusComposer()
}

async function loadMessages(id = selectedId.value) {
  if (!id) return
  const selection = activeSelection
  error.value = ''
  loadingThread.value = true
  try {
    const out = await api<{ messages: ChatMessage[]; hasMore?: boolean; nextBefore?: string }>(
      `/chat/conversations/${id}/messages?limit=50`,
    )
    if (selection !== activeSelection || selectedId.value !== id) return
    messages.value = out.messages
    hasMore.value = out.hasMore ?? false
    nextBefore.value = out.nextBefore ?? null
    pendingMessage.value = null
    showJump.value = false
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load messages')
  } finally {
    loadingThread.value = false
  }
}

async function loadOlder() {
  const id = selectedId.value
  const selection = activeSelection
  if (!id || !hasMore.value || !nextBefore.value || loadingOlder.value) return
  loadingOlder.value = true
  error.value = ''
  try {
    const container = threadBodyRef.value
    const heightBefore = container?.scrollHeight ?? 0
    const offsetBefore = container?.scrollTop ?? 0
    const out = await api<{ messages: ChatMessage[]; hasMore?: boolean; nextBefore?: string }>(
      `/chat/conversations/${id}/messages?limit=50&before=${nextBefore.value}`,
    )
    if (selection !== activeSelection || selectedId.value !== id) return
    messages.value = [...out.messages, ...messages.value]
    hasMore.value = out.hasMore ?? false
    nextBefore.value = out.nextBefore ?? null
    await nextTick()
    if (container) {
      container.scrollTop = container.scrollHeight - heightBefore + offsetBefore
    }
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load older messages')
  } finally {
    loadingOlder.value = false
  }
}

function onThreadScroll() {
  const el = threadBodyRef.value
  if (!el || searchResults.value) return
  const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight
  showJump.value = distanceFromBottom > 160
  if (el.scrollTop <= 0 && hasMore.value && !loadingOlder.value && visibleMessages.value.length) {
    void loadOlder()
  }
}

function scrollToLatest() {
  const el = threadBodyRef.value
  if (!el) return
  el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
}

function autoGrow() {
  const el = composerRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, 140)}px`
}

function shouldShowDate(index: number): boolean {
  const list = visibleMessages.value
  const current = list[index]
  if (!current || index === 0) return true
  const prev = list[index - 1]
  if (!prev) return true
  return new Date(prev.createdAt).toDateString() !== new Date(current.createdAt).toDateString()
}

function formatDayLabel(iso: string): string {
  const date = new Date(iso)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  if (date.toDateString() === today.toDateString()) return 'Today'
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday'
  return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function onComposerKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
    event.preventDefault()
    void sendText()
  }
}

async function searchMessages() {
  if (!selectedId.value || searchQuery.value.trim().length < 2) {
    searchResults.value = null
    return
  }
  error.value = ''
  try {
    const q = encodeURIComponent(searchQuery.value.trim())
    const out = await api<{ messages: ChatMessage[] }>(`/chat/conversations/${selectedId.value}/messages/search?q=${q}`)
    searchResults.value = out.messages
  } catch (e) {
    error.value = formatApiError(e, 'Search failed')
  }
}

function clearSearch() {
  searchResults.value = null
  searchQuery.value = ''
}

async function loadMedia(id = selectedId.value) {
  if (!id) return
  const selection = activeSelection
  try {
    const out = await api<{ media: ChatAttachment[] }>(`/chat/conversations/${id}/media`)
    if (selection !== activeSelection || selectedId.value !== id) return
    media.value = out.media
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load media')
  }
}

function canMutateMessage(message: ChatMessage): boolean {
  return (
    message.senderId === auth.user?.id &&
    !message.removedAt &&
    Date.now() - new Date(message.createdAt).getTime() <= 15 * 60 * 1000
  )
}

async function editMessage(message: ChatMessage) {
  const body = await ui.prompt({
    title: 'Edit message',
    label: 'Message',
    initialValue: message.body,
    confirmLabel: 'Save',
  })
  if (!body?.trim() || body.trim() === message.body || !selectedId.value) return
  try {
    const updated = await api<ChatMessage>(
      `/chat/conversations/${selectedId.value}/messages/${message.id}`,
      { method: 'PATCH', body: JSON.stringify({ body: body.trim() }) },
    )
    messages.value = messages.value.map((item) => item.id === updated.id ? updated : item)
  } catch (e) {
    error.value = formatApiError(e, 'Failed to edit message')
  }
}

async function removeMessage(message: ChatMessage) {
  if (!selectedId.value) return
  const confirmed = await ui.confirm({
    title: 'Remove message?',
    message: 'This message will be replaced with a removal notice for everyone.',
    confirmLabel: 'Remove',
    danger: true,
  })
  if (!confirmed) return
  try {
    await api(`/chat/conversations/${selectedId.value}/messages/${message.id}`, { method: 'DELETE' })
    messages.value = messages.value.map((item) =>
      item.id === message.id ? { ...item, body: '', removedAt: new Date().toISOString() } : item,
    )
  } catch (e) {
    error.value = formatApiError(e, 'Failed to remove message')
  }
}

function attachmentLabel(attachment: ChatAttachment): string {
  if (attachment.availability === 'trashed') return 'File moved to Trash'
  if (attachment.availability === 'purged') return 'File permanently deleted'
  return ''
}

async function sendText() {
  if (!selectedId.value || !draft.value.trim()) return
  const conversationId = selectedId.value
  const body = draft.value.trim()
  const clientMessageId = crypto.randomUUID()
  const optimistic: ChatMessage = {
    id: `pending-${Date.now()}`,
    conversationId,
    senderId: auth.user?.id ?? 'self',
    clientMessageId,
    body,
    createdAt: new Date().toISOString(),
    attachments: [],
  }
  pendingMessage.value = optimistic
  pendingMessageError.value = ''
  draft.value = ''
  autoGrow()
  error.value = ''
  sending.value = true
  try {
    const message = await chatStore.sendMessage(body, clientMessageId, conversationId)
    messages.value = messages.value.some((item) => item.clientMessageId === clientMessageId)
      ? messages.value.map((item) => item.clientMessageId === clientMessageId ? message : item)
      : [...messages.value, message]
    searchResults.value = null
    pendingMessage.value = null
    showJump.value = false
    await loadConversations()
    selectedId.value = message.conversationId
  } catch (e) {
    draft.value = draft.value ? draft.value : body
    pendingMessageError.value = formatApiError(e, 'Failed to send message')
    error.value = formatApiError(e, 'Failed to send message')
  } finally {
    sending.value = false
  }
}

function discardPendingMessage() {
  pendingMessage.value = null
  pendingMessageError.value = ''
}

function retryPendingMessage() {
  if (!pendingMessage.value) return
  draft.value = pendingMessage.value.body
  discardPendingMessage()
  void sendText()
}

function triggerAttachment() {
  fileInputRef.value?.click()
}

async function onAttachmentChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !selectedId.value) return
  const item: AttachmentUpload = {
    id: crypto.randomUUID(),
    file,
    conversationId: selectedId.value,
    body: draft.value.trim(),
    status: 'queued',
    controller: new AbortController(),
  }
  attachmentQueue.value = [...attachmentQueue.value, item]
  draft.value = ''
  autoGrow()
  void processAttachment(item.id)
}

async function processAttachment(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item || item.status === 'uploading' || item.status === 'canceled') return
  item.status = 'uploading'
  item.error = undefined
  sending.value = true
  uploadProgress.value = 0
  try {
    const contentType = item.file.type || 'application/octet-stream'
    const session = await api<UploadSession>('/chat/attachments/upload-sessions', {
      method: 'POST',
      body: JSON.stringify({
        conversationId: item.conversationId,
        name: item.file.name,
        size: item.file.size,
        contentType,
      }),
    })
    await uploadToPresigned(session.uploadUrl, item.file, contentType, (ratio) => {
      uploadProgress.value = ratio
    }, item.controller.signal)
    const message = await api<ChatMessage>(`/chat/attachments/${session.fileId}/complete`, {
      method: 'POST',
      body: JSON.stringify({ conversationId: item.conversationId, body: item.body }),
    })
    if (selectedId.value === item.conversationId) messages.value = [...messages.value, message]
    searchResults.value = null
    await loadMedia(item.conversationId)
    attachmentQueue.value = attachmentQueue.value.filter((entry) => entry.id !== id)
    ui.showToast('Attachment sent')
  } catch (e) {
    item.status = 'failed'
    item.error = formatApiError(e, 'Failed to send attachment')
  } finally {
    uploadProgress.value = null
    sending.value = attachmentQueue.value.some((entry) => entry.status === 'uploading')
  }
}

function retryUpload(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item) return
  item.status = 'queued'
  void processAttachment(id)
}

function cancelUpload(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item) return
  item.status = 'canceled'
  item.controller.abort()
  attachmentQueue.value = attachmentQueue.value.filter((entry) => entry.id !== id)
}

async function openAttachment(fileId: string | null) {
  if (!fileId || !selectedId.value) return
  try {
    const attachment = messages.value.flatMap((message) => message.attachments).find((item) => item.fileId === fileId)
    if (attachment && attachment.availability !== 'available') return
    const out = await api<DownloadURL>(`/chat/conversations/${selectedId.value}/attachments/${attachment?.id ?? fileId}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, 'Download failed')
  }
}

async function openInlineImage(attachment: ChatAttachment) {
  if (attachment.availability !== 'available' || !attachment.fileId || !selectedId.value) return
  error.value = ''
  try {
    const out = await api<DownloadURL>(
      `/chat/conversations/${selectedId.value}/attachments/${attachment.id}/download`,
    )
    lightboxFileId.value = attachment.fileId
    lightboxName.value = attachment.name
    lightboxMime.value = attachment.mimeType
    lightboxUrl.value = out.downloadUrl
    lightboxOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, 'View failed')
  }
}

async function downloadLightbox() {
  if (!lightboxFileId.value) return
  await openAttachment(lightboxFileId.value)
}

function focusComposer() {
  void nextTick(() => composerRef.value?.focus())
}

onMounted(() => {
  updateViewport()
  window.addEventListener('resize', updateViewport)
  window.addEventListener('offline', chatStore.markOffline)
  window.addEventListener('online', chatStore.markOnline)
  void loadConversations()
  chatStore.connectEvents()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewport)
  window.removeEventListener('offline', chatStore.markOffline)
  window.removeEventListener('online', chatStore.markOnline)
  chatStore.stopEvents()
})

watch(visibleMessages, () => {
  void nextTick(() => {
    const el = threadBodyRef.value
    if (el && !loadingOlder.value && !searchResults.value) {
      const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight
      if (distanceFromBottom < 160) el.scrollTop = el.scrollHeight
    }
  })
})

watch(
  () => chatStore.messages[selectedId.value ?? ''],
  (nextMessages) => {
    if (nextMessages && selectedId.value) {
      messages.value = [...nextMessages]
    }
  },
)

watch(
  () => chatStore.conversations,
  (nextConversations) => {
    conversations.value = [...nextConversations]
  },
  { deep: true },
)
</script>

<template>
  <div class="chat-app" :class="{ 'in-thread': inThread }">
    <aside class="chat-rail" aria-label="Conversations">
      <header class="rail-header">
        <span class="chat-mark" aria-hidden="true">F</span>
        <h1>Chats</h1>
        <button type="button" class="icon-btn" aria-label="Back to Filvault" @click="backToVault">
          <Icon name="folder" :size="20" />
        </button>
        <button type="button" class="compose-btn" aria-label="Create chat" @click="promptNewConversation">
          <Icon name="plus" :size="18" />
        </button>
      </header>
      <div class="rail-filter">
        <label class="sr-only" for="rail-filter-input">Filter chats</label>
        <input
          id="rail-filter-input"
          v-model="railFilter"
          type="search"
          placeholder="Search chats"
          autocomplete="off"
        />
      </div>
      <div v-if="error && !inThread" class="alert" role="alert">{{ error }}</div>
      <nav class="conversation-list">
        <p v-if="!filteredConversations.length" class="rail-empty">No chats match "{{ railFilter }}".</p>
    <button
          v-for="conv in filteredConversations"
          :key="conv.id"
          type="button"
          class="conversation-row"
          :class="{ active: conv.id === selectedId }"
          :aria-current="conv.id === selectedId ? 'true' : undefined"
          @click="selectConversation(conv.id)"
        >
          <span class="avatar" :class="avatarClass(conversationTitle(conv))" aria-hidden="true">{{ conversationTitle(conv).slice(0, 1).toUpperCase() }}</span>
          <span class="conversation-meta">
            <span class="conversation-title">{{ conversationTitle(conv) }}</span>
            <span class="conversation-preview">{{ conv.preview?.body || (conv.preview?.attachments?.length ? 'Attachment' : 'No messages yet') }}</span>
          </span>
          <span class="conversation-date">{{ formatRelativeDay(conv.preview?.createdAt ?? conv.updatedAt) }}</span>
        </button>
      </nav>
    </aside>

    <section class="message-thread" aria-live="polite">
      <p v-if="chatStore.connectionState !== 'connected'" class="connection-status" role="status">
        {{ chatStore.connectionState === 'offline' ? 'Offline — messages will retry when connected.' : 'Reconnecting…' }}
      </p>
      <template v-if="selectedConversation">
        <header class="thread-header">
          <button type="button" class="icon-btn back-btn" aria-label="Back" @click="backToRail">
            <Icon name="arrow-left" :size="20" />
          </button>
          <span class="thread-avatar" :class="avatarClass(conversationTitle(selectedConversation))" aria-hidden="true">{{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}</span>
          <h2>{{ conversationTitle(selectedConversation) }}</h2>
          <button class="ghost-btn" type="button" @click="loadMessages()">Refresh</button>
          <button
            class="icon-btn"
            type="button"
            :aria-expanded="threadSearchOpen"
            aria-label="Search messages"
            @click="threadSearchOpen = !threadSearchOpen"
          >
            <Icon name="search" :size="18" />
          </button>
        </header>

        <form v-if="threadSearchOpen" class="chat-search" @submit.prevent="searchMessages">
          <label class="sr-only" for="chat-search">Search messages</label>
          <input id="chat-search" v-model="searchQuery" type="search" placeholder="Search in this chat" />
          <button class="ghost-btn" type="submit" :disabled="searchQuery.trim().length < 2">Search</button>
          <button v-if="searchResults" class="ghost-btn" type="button" @click="clearSearch">Clear</button>
        </form>

        <p v-if="error && inThread" class="alert" role="alert">{{ error }}</p>

        <div
          ref="threadBodyRef"
          class="message-body"
          :aria-busy="loadingThread"
          @scroll.passive="onThreadScroll"
        >
          <LoadingSkeletonThread v-if="loadingThread" />
          <p v-else-if="loadingOlder" class="loading-older">Loading older…</p>
          <div v-if="!loadingThread && visibleMessages.length" class="message-list">
            <TransitionGroup name="msg">
              <template v-for="(message, index) in visibleMessages" :key="message.id">
                <div v-if="shouldShowDate(index)" class="day-separator">
                  <span>{{ formatDayLabel(message.createdAt) }}</span>
                </div>
                <article class="message-bubble" :class="{ outgoing: message.senderId === auth.user?.id }">
                  <p v-if="message.body">{{ message.body }}</p>
                  <template v-for="attachment in message.attachments" :key="attachment.id">
                    <button
                      v-if="attachment.availability === 'available' && attachment.thumbnailUrl && attachment.mimeType.startsWith('image/')"
                      type="button"
                      class="inline-image"
                      :aria-label="`View ${attachment.name}`"
                      @click="openInlineImage(attachment)"
                    >
                      <img
                        :src="attachment.thumbnailUrl"
                        :alt="attachment.name"
                        loading="lazy"
                      />
                    </button>
                    <button
                      v-else
                      type="button"
                      class="attachment-card"
                      :disabled="attachment.availability !== 'available' || !attachment.fileId"
                      @click="attachment.fileId && openAttachment(attachment.fileId)"
                    >
                      <Icon :name="attachment.mimeType.startsWith('image/') ? 'image' : 'file'" :size="18" />
                      <span class="attachment-name">{{ attachment.name }}</span>
                      <span class="attachment-size">{{ formatBytes(attachment.sizeBytes) }}</span>
                      <span v-if="attachment.availability !== 'available'" class="attachment-status">{{ attachmentLabel(attachment) }}</span>
                    </button>
                  </template>
                  <span v-if="message.removedAt" class="message-status">Message removed</span>
                  <span v-else-if="message.editedAt" class="message-status">Edited</span>
                  <span class="bubble-time">{{ formatTime(message.createdAt) }}</span>
                  <div v-if="canMutateMessage(message)" class="message-actions">
                    <button type="button" class="message-action" @click="editMessage(message)">Edit</button>
                    <button type="button" class="message-action danger-text" @click="removeMessage(message)">Remove</button>
                  </div>
                </article>
              </template>
            </TransitionGroup>
            <Transition name="msg">
              <article v-if="pendingMessage" class="message-bubble pending" aria-live="polite">
                <p>{{ pendingMessage.body }}</p>
                <span v-if="pendingMessageError" class="message-status">{{ pendingMessageError }}</span>
                <div v-if="pendingMessageError" class="message-actions">
                  <button type="button" class="message-action" @click="retryPendingMessage">Retry</button>
                  <button type="button" class="message-action" @click="discardPendingMessage">Discard</button>
                </div>
              </article>
            </Transition>
          </div>
          <EmptyState v-else title="No messages yet" description="Start the conversation with a message or an image." icon="chat" />
          <Transition name="msg">
            <button v-if="showJump && !searchResults" type="button" class="jump-latest" @click="scrollToLatest">
              <Icon name="download" :size="16" />
              <span>Latest</span>
            </button>
          </Transition>
        </div>

        <UploadProgress :progress="uploadProgress" />
        <ul v-if="attachmentQueue.length" class="attachment-queue" aria-live="polite">
          <li v-for="item in attachmentQueue" :key="item.id" class="queue-item">
            <span class="queue-name">{{ item.file.name }}</span>
            <span class="queue-status">{{ item.status === 'failed' ? item.error : item.status }}</span>
            <button v-if="item.status === 'failed'" type="button" class="message-action" @click="retryUpload(item.id)">Retry upload</button>
            <button v-if="item.status === 'uploading' || item.status === 'queued'" type="button" class="message-action" @click="cancelUpload(item.id)">Cancel upload</button>
          </li>
        </ul>
        <form class="chat-composer" @submit.prevent="sendText">
          <button type="button" class="icon-btn attach-btn" aria-label="Attach file" :disabled="sending" @click="triggerAttachment">
            <Icon name="plus" :size="18" />
          </button>
          <label class="sr-only" for="chat-message">Message</label>
          <textarea
            id="chat-message"
            ref="composerRef"
            v-model="draft"
            rows="1"
            placeholder="Aa"
            :disabled="sending"
            @input="autoGrow"
            @keydown.enter="onComposerKeydown"
          ></textarea>
          <button
            class="send-btn"
            type="submit"
            aria-label="Send"
            :class="{ active: Boolean(draft.trim()) }"
            :disabled="!draft.trim() || sending"
          >
            <Icon name="check" :size="18" />
          </button>
          <label class="sr-only" for="chat-attachment">Attachment</label>
          <input id="chat-attachment" ref="fileInputRef" type="file" class="sr-only" @change="onAttachmentChange" />
        </form>
      </template>
      <EmptyState v-else title="Select a chat" description="Pick a conversation from the list or create a new one." icon="chat" />
    </section>

    <aside class="media-panel" aria-label="Shared media">
      <h2>Media</h2>
        <button
        v-for="item in media"
        :key="item.id"
        type="button"
        class="media-card"
        :disabled="!item.fileId"
          @click="item.fileId && openAttachment(item.fileId)"
      >
        <img
          v-if="item.thumbnailUrl"
          :src="item.thumbnailUrl"
          :alt="item.name"
          class="media-thumb"
          loading="lazy"
        />
        <Icon v-else :name="item.mimeType.startsWith('image/') ? 'image' : 'file'" :size="18" />
        <span class="media-name">{{ item.name }}</span>
        <small>{{ new Date(item.createdAt).toLocaleDateString() }}</small>
      </button>
      <p v-if="!media.length" class="muted-hint">No shared photos or videos yet.</p>
    </aside>

    <MediaLightbox
      :open="lightboxOpen"
      :name="lightboxName"
      :mime-type="lightboxMime"
      :url="lightboxUrl"
      @download="downloadLightbox"
      @close="lightboxOpen = false"
    />
  </div>
</template>

<style scoped>
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.chat-app {
  display: grid;
  height: 100vh;
  height: 100dvh;
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: minmax(0, 1fr);
  background: var(--surface-soft);
}

.chat-rail,
.message-thread,
.media-panel {
  min-height: 0;
  min-width: 0;
}

.chat-rail {
  display: flex;
  flex-direction: column;
  background: var(--canvas);
  border-right: 1px solid var(--hairline);
}

.rail-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: 60px;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
}

.rail-header h1 {
  flex: 1;
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--ink);
}

.chat-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent-hover);
  font-weight: 700;
}

.icon-btn,
.compose-btn,
.ghost-btn,
.send-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
}

.icon-btn:hover,
.compose-btn:hover,
.ghost-btn:hover {
  background: var(--surface-card);
}

.icon-btn:focus-visible,
.compose-btn:focus-visible,
.ghost-btn:focus-visible,
.send-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.icon-btn:disabled,
.ghost-btn:disabled,
.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.alert {
  margin: 0 var(--space-md) var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--danger);
  border-radius: var(--radius-lg);
  background: var(--danger-soft);
  color: var(--danger);
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-xs) var(--space-sm) var(--space-md);
}

.conversation-row {
  display: flex;
  width: 100%;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  border: none;
  border-radius: var(--radius-lg);
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.conversation-row:hover {
  background: var(--surface-soft);
}

.conversation-row.active {
  background: var(--accent-soft);
}

.conversation-row:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
  font-weight: 600;
}

.avatar-color-a { background: #e0e7ff; color: #3730a3; }
.avatar-color-b { background: #fce7f3; color: #9d174d; }
.avatar-color-c { background: #dcfce7; color: #166534; }
.avatar-color-d { background: #ffedd5; color: #9a3412; }
.avatar-color-e { background: #cffafe; color: #155e75; }
.avatar-color-f { background: #f3e8ff; color: #6b21a8; }

.rail-filter {
  padding: 0 var(--space-md) var(--space-sm);
}

.rail-filter input {
  width: 100%;
  min-height: 36px;
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
  font-size: 13px;
}

.rail-filter input:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.rail-empty {
  margin: 0;
  padding: var(--space-md);
  color: var(--muted);
  font-size: 13px;
  text-align: center;
}

.conversation-meta {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.conversation-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ink);
  font-weight: 600;
}

.conversation-preview {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
  font-size: 13px;
}

.conversation-date {
  color: var(--muted);
  font-size: 12px;
}

.message-thread {
  display: flex;
  flex-direction: column;
  background: var(--canvas);
}

.thread-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: 60px;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
}

.thread-header h2 {
  flex: 1;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 16px;
  font-weight: 600;
  color: var(--ink);
}

.thread-avatar {
  display: none;
}

.back-btn {
  display: inline-flex;
}

.chat-search {
  display: flex;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-md);
  border-bottom: 1px solid var(--hairline);
}

.chat-search input {
  min-width: 0;
  flex: 1;
  min-height: 40px;
  padding: 8px 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  font: inherit;
}

.chat-search input:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.message-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  position: relative;
}

.loading-older {
  margin: 0;
  padding: var(--space-xs);
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.jump-latest {
  position: sticky;
  bottom: var(--space-sm);
  align-self: center;
  display: inline-flex;
  align-items: center;
  gap: var(--space-xxs);
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--canvas);
  color: var(--ink);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(17, 24, 39, 0.12);
}

.jump-latest:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.message-bubble.pending {
  opacity: 0.55;
}

.bubble-time {
  display: block;
  margin-top: 2px;
  text-align: right;
  opacity: 0.65;
  font-size: 11px;
}

.inline-image {
  display: block;
  width: min(260px, 100%);
  margin-top: var(--space-xs);
  padding: 0;
  border: none;
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: zoom-in;
  background: transparent;
}

.inline-image img {
  display: block;
  width: 100%;
  max-height: 320px;
  object-fit: cover;
}

.inline-image:focus-visible {
  outline: 2px solid var(--on-ink, #ffffff);
  outline-offset: 2px;
}

.day-separator {
  align-self: center;
  padding: var(--space-xxs) var(--space-md);
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--muted);
  font-size: 12px;
  font-weight: 600;
}

.msg-enter-active {
  transition: opacity var(--duration-medium) var(--ease-emphasized-decelerate),
    transform var(--duration-medium) var(--ease-emphasized-decelerate);
}

.msg-leave-active {
  transition: opacity var(--duration-short) var(--ease-standard);
}

.msg-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.97);
}

.msg-leave-to {
  opacity: 0;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  padding: var(--space-md);
}

.message-bubble {
  align-self: flex-end;
  max-width: min(78%, 520px);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-xl) var(--radius-xl) var(--radius-sm) var(--radius-xl);
  background: var(--primary-cta, #111827);
  color: var(--on-ink, #ffffff);
}

.message-bubble:not(.outgoing) {
  align-self: flex-start;
  background: var(--surface-card);
  color: var(--ink);
}

.message-status {
  display: inline-block;
  margin-right: var(--space-xs);
  color: inherit;
  font-size: 12px;
  font-style: italic;
  opacity: 0.72;
}

.message-actions {
  display: flex;
  gap: var(--space-xs);
  margin-top: var(--space-xs);
}

.message-action {
  padding: 2px 0;
  border: 0;
  background: transparent;
  color: inherit;
  font-size: 12px;
  text-decoration: underline;
  cursor: pointer;
}

.danger-text {
  color: var(--danger);
}

.attachment-queue {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  margin: 0;
  padding: var(--space-xs) var(--space-md);
  border-top: 1px solid var(--hairline);
  list-style: none;
}

.queue-item {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  min-width: 0;
  color: var(--muted);
  font-size: 12px;
}

.queue-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-status {
  margin-left: auto;
  flex-shrink: 0;
}

.connection-status {
  margin: 0;
  padding: 6px var(--space-md);
  background: var(--surface-soft);
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.message-bubble p {
  margin: 0;
  white-space: pre-wrap;
}

.attachment-card {
  display: flex;
  width: 100%;
  align-items: center;
  gap: var(--space-xs);
  margin-top: var(--space-xs);
  padding: var(--space-xs);
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.14);
  color: inherit;
  cursor: pointer;
}

.attachment-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-size {
  flex-shrink: 0;
  opacity: 0.75;
  font-size: 12px;
}

.chat-composer {
  display: flex;
  position: relative;
  align-items: flex-end;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md) calc(var(--space-sm) + env(safe-area-inset-bottom));
  border-top: 1px solid var(--hairline);
  background: var(--canvas);
}

.chat-composer textarea {
  min-height: 44px;
  max-height: 140px;
  flex: 1;
  padding: 10px 14px;
  resize: none;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  font: inherit;
}

.chat-composer textarea:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.send-btn {
  background: var(--primary-cta, #111827);
  color: var(--on-ink, #ffffff);
  opacity: 0.4;
  transition:
    opacity var(--duration-short) var(--ease-standard),
    transform var(--duration-short) var(--ease-standard);
}

.send-btn.active {
  opacity: 1;
  transform: scale(1.04);
}

.media-panel {
  display: none;
}

.muted-hint {
  color: var(--muted);
  font-size: 14px;
}

@media (min-width: 768px) {
  .chat-app {
    grid-template-columns: 320px minmax(0, 1fr);
    grid-template-rows: minmax(0, 1fr);
  }

  .back-btn {
    display: none;
  }

  .thread-avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    border-radius: var(--radius-pill);
    background: var(--surface-card);
    color: var(--ink);
    font-weight: 600;
  }

  .message-thread {
    border-right: 1px solid var(--hairline);
  }
}

@media (min-width: 1200px) {
  .chat-app {
    grid-template-columns: 300px minmax(0, 1fr) 280px;
  }

  .media-panel {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: var(--space-md);
    background: var(--canvas);
    overflow-y: auto;
  }

  .media-panel h2 {
    margin: 0 0 var(--space-xs);
    font-size: 16px;
    font-weight: 600;
    color: var(--ink);
  }

  .media-card {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 2px var(--space-xs);
    align-items: center;
    padding: var(--space-xs);
    border: 1px solid var(--hairline);
    border-radius: var(--radius-lg);
    background: transparent;
    color: var(--ink);
    cursor: pointer;
    text-align: left;
  }

  .media-card:hover {
    background: var(--surface-soft);
  }

  .media-card:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: -2px;
  }

  .media-thumb {
    width: 40px;
    height: 40px;
    object-fit: cover;
    border-radius: var(--radius-md);
  }

  .media-name,
  .media-card small {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .media-card small {
    grid-column: 2;
    color: var(--muted);
    font-size: 12px;
  }
}

@media (min-width: 768px) and (max-width: 1199px) {
  .chat-app {
    grid-template-columns: 320px minmax(0, 1fr);
  }

  .media-panel {
    display: none;
  }
}

/* Mobile master-detail: show rail OR thread, never both stacked */
@media (max-width: 767px) {
  .media-panel {
    display: none;
  }

  .message-thread {
    display: none;
  }

  .chat-app.in-thread .chat-rail {
    display: none;
  }

  .chat-app.in-thread .message-thread {
    display: flex;
  }
}

</style>
