<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, uploadToPresigned, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import { useChatStore } from '@/stores/chat'
import { useCallStore } from '@/stores/call'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import LoadingSkeletonThread from '@/components/LoadingSkeletonThread.vue'
import UploadProgress from '@/components/UploadProgress.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import { userInitials } from '@/lib/userInitials'
import { useI18n } from '@/lib/i18n'
import { generateUUID } from '@/lib/uuid'
import type { ChatAttachment, ChatConversation, ChatMessage, DownloadURL, UploadSession } from '@/api/types'

const router = useRouter()
const ui = useUiStore()
const chatStore = useChatStore()
const callStore = useCallStore()
const auth = useAuthStore()
const { t, locale, setLocale } = useI18n()

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
  if (date.toDateString() === yesterday.toDateString()) return t.value.yesterday
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
    title: t.value.newChat,
    label: t.value.verifiedEmail,
    confirmLabel: t.value.openChat,
  }).then((email) => {
    if (!email?.trim()) return
    void createDirectConversation(email.trim())
  })
}

function backToVault() {
  void router.push('/files')
}

const appMenuOpen = ref(false)
const isDarkMode = ref(document.documentElement.dataset.theme === 'dark')

function toggleDarkMode() {
  if (isDarkMode.value) {
    delete document.documentElement.dataset.theme
    localStorage.removeItem('filvault.theme')
    isDarkMode.value = false
  } else {
    document.documentElement.dataset.theme = 'dark'
    localStorage.setItem('filvault.theme', 'dark')
    isDarkMode.value = true
  }
}

function toggleLanguage() {
  const next = locale.value === 'vi' ? 'en' : 'vi'
  setLocale(next)
}

function navigateTo(path: string) {
  appMenuOpen.value = false
  void router.push(path)
}

async function handleLogout() {
  appMenuOpen.value = false
  await auth.logout()
  void router.push('/login')
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

function conversationPreview(conv: ChatConversation): string {
  if (conv.preview?.body) {
    return conv.preview.body
  }
  if (conv.preview?.attachments?.length) {
    return t.value.attachment
  }
  return t.value.noMessagesYet
}

function isFirstInCluster(index: number): boolean {
  const list = visibleMessages.value
  const current = list[index]
  if (!current) return false
  if (index === 0) return true
  const prev = list[index - 1]
  if (!prev) return true
  if (shouldShowDate(index)) return true
  return prev.senderId !== current.senderId
}

function isLastInCluster(index: number): boolean {
  const list = visibleMessages.value
  const current = list[index]
  if (!current) return false
  if (index === list.length - 1) return true
  const next = list[index + 1]
  if (!next) return true
  if (shouldShowDate(index + 1)) return true
  return next.senderId !== current.senderId
}

function sendQuickLike() {
  draft.value = '👍'
  void sendText()
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
  if (sending.value) return
  if (!selectedId.value || !draft.value.trim()) return
  const conversationId = selectedId.value
  const body = draft.value.trim()
  const clientMessageId = generateUUID()
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
  sending.value = true
  await nextTick()
  autoGrow()
  error.value = ''
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
    id: generateUUID(),
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

async function openAttachment(attachmentId: string) {
  if (!selectedId.value) return
  try {
    const out = await api<DownloadURL>(`/chat/conversations/${selectedId.value}/attachments/${attachmentId}/download`)
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
    lightboxFileId.value = attachment.id
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

function onVisualViewportResize() {
  if (!isMobile.value) return
  if (inThread.value) {
    void nextTick(() => {
      scrollToLatest()
    })
  }
}

onMounted(() => {
  updateViewport()
  window.addEventListener('resize', updateViewport)
  window.addEventListener('offline', chatStore.markOffline)
  window.addEventListener('online', chatStore.markOnline)
  if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', onVisualViewportResize)
    window.visualViewport.addEventListener('scroll', onVisualViewportResize)
  }
  void loadConversations()
  chatStore.connectEvents()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewport)
  window.removeEventListener('offline', chatStore.markOffline)
  window.removeEventListener('online', chatStore.markOnline)
  if (window.visualViewport) {
    window.visualViewport.removeEventListener('resize', onVisualViewportResize)
    window.visualViewport.removeEventListener('scroll', onVisualViewportResize)
  }
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
    // Only sync from store on initial load (when local messages are empty)
    // Avoid overwriting older messages loaded via pagination
    if (nextMessages && selectedId.value && messages.value.length === 0) {
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
        <button
          type="button"
          class="chat-mark-btn"
          :aria-label="t.appMenu"
          :title="t.appMenu"
          @click="appMenuOpen = true"
        >
          <span class="chat-mark" aria-hidden="true">F</span>
        </button>
        <h1>{{ t.chats }}</h1>
        <button type="button" class="icon-btn" aria-label="Back to Filvault" @click="backToVault">
          <Icon name="folder" :size="20" />
        </button>
        <button type="button" class="compose-btn" :aria-label="t.newChat" @click="promptNewConversation">
          <Icon name="plus" :size="18" />
        </button>
      </header>
      <div class="rail-filter">
        <label class="sr-only" for="rail-filter-input">{{ t.filterChats }}</label>
        <input
          id="rail-filter-input"
          v-model="railFilter"
          type="search"
          :placeholder="t.searchChats"
          autocomplete="off"
        />
      </div>
      <div v-if="error && !inThread" class="alert" role="alert">{{ error }}</div>
      <nav class="conversation-list">
        <div v-if="!filteredConversations.length" class="rail-empty">
          <p v-if="railFilter.trim()" class="rail-empty-query">{{ t.noChatsMatch }} "{{ railFilter }}".</p>
          <div v-else class="rail-empty-state">
            <p class="rail-empty-title">{{ t.noConversationsYet }}</p>
            <p class="rail-empty-hint">{{ t.startNewChatHint }}</p>
            <button type="button" class="btn ink rail-start-btn" @click="promptNewConversation">
              <Icon name="plus" :size="16" />
              <span>{{ t.startNewChat }}</span>
            </button>
          </div>
        </div>
        <button
          v-for="conv in filteredConversations"
          :key="conv.id"
          type="button"
          class="conversation-row"
          :class="{ active: conv.id === selectedId }"
          :aria-current="conv.id === selectedId ? 'true' : undefined"
          @click="selectConversation(conv.id)"
        >
          <span class="avatar-wrap">
            <span class="avatar" :class="avatarClass(conversationTitle(conv))" aria-hidden="true">{{ conversationTitle(conv).slice(0, 1).toUpperCase() }}</span>
            <span class="online-indicator" aria-hidden="true" />
          </span>
          <span class="conversation-meta">
            <span class="conversation-title">{{ conversationTitle(conv) }}</span>
            <span class="conversation-preview">{{ conversationPreview(conv) }}</span>
          </span>
          <span class="conversation-date">{{ formatRelativeDay(conv.preview?.createdAt ?? conv.updatedAt) }}</span>
        </button>
      </nav>
    </aside>

    <section class="message-thread" aria-live="polite">
      <p v-if="chatStore.connectionState !== 'connected'" class="connection-status" role="status">
        {{ chatStore.connectionState === 'offline' ? t.offlineStatus : t.reconnecting }}
      </p>
      <template v-if="selectedConversation">
        <header class="thread-header">
          <!-- aria-label="Back" -->
          <button type="button" class="icon-btn back-btn" :aria-label="t.back" @click="backToRail">
            <Icon name="arrow-left" :size="20" />
          </button>
          <div class="thread-peer-info">
            <span class="thread-avatar" :class="avatarClass(conversationTitle(selectedConversation))" aria-hidden="true">{{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}</span>
            <div class="thread-peer-meta">
              <h2>{{ conversationTitle(selectedConversation) }}</h2>
              <span class="thread-status">{{ t.activeNow }}</span>
            </div>
          </div>
          <div class="thread-actions">
            <button
              class="icon-btn"
              type="button"
              aria-label="Gọi thoại"
              title="Gọi thoại"
              @click="callStore.startCall(selectedConversation.id, { isVideo: false })"
            >
              <Icon name="phone" :size="18" />
            </button>
            <button
              class="icon-btn"
              type="button"
              aria-label="Gọi video"
              title="Gọi video"
              @click="callStore.startCall(selectedConversation.id, { isVideo: true })"
            >
              <Icon name="camera" :size="18" />
            </button>
            <button
              class="icon-btn"
              type="button"
              :aria-expanded="threadSearchOpen"
              :aria-label="t.search"
              @click="threadSearchOpen = !threadSearchOpen"
            >
              <Icon name="search" :size="18" />
            </button>
          </div>
        </header>

        <form v-if="threadSearchOpen" class="chat-search" @submit.prevent="searchMessages">
          <label class="sr-only" for="chat-search">{{ t.search }}</label>
          <input id="chat-search" v-model="searchQuery" type="search" :placeholder="t.searchInChat" />
          <button class="ghost-btn" type="submit" :disabled="searchQuery.trim().length < 2">{{ t.search }}</button>
          <button v-if="searchResults" class="ghost-btn" type="button" @click="clearSearch">{{ t.clear }}</button>
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
                <div
                  class="message-row"
                  :class="{
                    outgoing: message.senderId === auth.user?.id,
                    'cluster-start': isFirstInCluster(index),
                    'cluster-last': isLastInCluster(index),
                    'cluster-middle': !isFirstInCluster(index) && !isLastInCluster(index),
                    'cluster-single': isFirstInCluster(index) && isLastInCluster(index),
                  }"
                >
                  <span
                    v-if="message.senderId !== auth.user?.id && isLastInCluster(index)"
                    class="row-avatar"
                    :class="avatarClass(conversationTitle(selectedConversation))"
                    aria-hidden="true"
                  >
                    {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
                  </span>
                  <span
                    v-else-if="message.senderId !== auth.user?.id"
                    class="row-avatar-spacer"
                    aria-hidden="true"
                  />
                  <article class="message-bubble" :class="{ outgoing: message.senderId === auth.user?.id, 'has-like': message.body === '👍' }">
                    <p v-if="message.body" :class="{ 'like-bubble': message.body === '👍' }">{{ message.body }}</p>
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
                        @click="openAttachment(attachment.id)"
                      >
                        <Icon :name="attachment.mimeType.startsWith('image/') ? 'image' : 'file'" :size="18" />
                        <span class="attachment-name">{{ attachment.name }}</span>
                        <span class="attachment-size">{{ formatBytes(attachment.sizeBytes) }}</span>
                        <span v-if="attachment.availability !== 'available'" class="attachment-status">{{ attachmentLabel(attachment) }}</span>
                      </button>
                    </template>
                    <span v-if="message.removedAt" class="message-status">{{ t.messageRemoved }}</span>
                    <span v-else-if="message.editedAt" class="message-status">{{ t.edited }}</span>
                    <span class="bubble-time">{{ formatTime(message.createdAt) }}</span>
                    <div v-if="canMutateMessage(message)" class="message-actions">
                      <button type="button" class="message-action" @click="editMessage(message)">{{ t.edit }}</button>
                      <button type="button" class="message-action danger-text" @click="removeMessage(message)">{{ t.remove }}</button>
                    </div>
                  </article>
                </div>
              </template>
            </TransitionGroup>
            <Transition name="msg">
              <div v-if="pendingMessage" class="message-row outgoing cluster-single">
                <article class="message-bubble pending outgoing" :class="{ 'has-like': pendingMessage.body === '👍' }" aria-live="polite">
                  <p :class="{ 'like-bubble': pendingMessage.body === '👍' }">{{ pendingMessage.body }}</p>
                  <span v-if="pendingMessageError" class="message-status">{{ pendingMessageError }}</span>
                  <div v-if="pendingMessageError" class="message-actions">
                    <button type="button" class="message-action" @click="retryPendingMessage">{{ t.retry }}<!-- Retry --></button>
                    <button type="button" class="message-action" @click="discardPendingMessage">{{ t.closeSelection }}<!-- Discard --></button>
                  </div>
                </article>
              </div>
            </Transition>
          </div>
          <EmptyState v-else :title="t.noMessagesYet" description="" icon="chat" />
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
            v-if="!draft.trim()"
            type="button"
            class="like-btn"
            :aria-label="t.sendLike"
            :disabled="sending"
            @click="sendQuickLike"
          >
            <Icon name="thumb-up" :size="20" />
          </button>
          <button
            v-show="Boolean(draft.trim())"
            class="send-btn"
            type="submit"
            aria-label="Send"
            :class="{ active: Boolean(draft.trim()) }"
            :disabled="!draft.trim() || sending"
            @click.prevent="sendText"
          >
            <Icon name="send" :size="18" />
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
          @click="openAttachment(item.id)"
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

    <BottomSheet :open="appMenuOpen" :title="t.filvaultMenu" @close="appMenuOpen = false">
      <div class="app-menu-content">
        <!-- User Profile Card -->
        <button type="button" class="menu-profile-card" @click="navigateTo('/profile')">
          <span class="menu-avatar" aria-hidden="true">{{ userInitials(auth.user?.displayName ?? '', auth.user?.email ?? '') }}</span>
          <span class="menu-profile-info">
            <span class="menu-profile-name">{{ auth.user?.displayName || t.myAccount }}</span>
            <span class="menu-profile-email">{{ auth.user?.email }}</span>
            <span class="menu-status-badge">
              <span class="status-dot" aria-hidden="true" />
              <span>{{ t.activeStatus }}: {{ t.activeStatusOn }}</span>
            </span>
          </span>
          <Icon name="arrow-right" :size="16" class="menu-profile-arrow" />
        </button>

        <!-- Section 1: Điều hướng Filvault -->
        <div class="menu-section">
          <span class="menu-section-label">Filvault</span>
          <div class="menu-items-group">
            <button type="button" class="menu-row-item" @click="navigateTo('/files')">
              <span class="menu-item-icon badge-folder"><Icon name="folder" :size="18" /></span>
              <span class="menu-item-text">{{ t.navFiles }}</span>
            </button>
            <button type="button" class="menu-row-item" @click="navigateTo('/photos')">
              <span class="menu-item-icon badge-photos"><Icon name="photos" :size="18" /></span>
              <span class="menu-item-text">{{ t.navPhotos }}</span>
            </button>
            <button type="button" class="menu-row-item" @click="navigateTo('/trash')">
              <span class="menu-item-icon badge-trash"><Icon name="trash" :size="18" /></span>
              <span class="menu-item-text">{{ t.navTrash }}</span>
            </button>
          </div>
        </div>

        <!-- Section 2: Tùy chọn giao diện -->
        <div class="menu-section">
          <span class="menu-section-label">{{ t.appearance }} &amp; {{ t.navSettings }}</span>
          <div class="menu-items-group">
            <button type="button" class="menu-row-item" @click="toggleDarkMode">
              <span class="menu-item-icon badge-theme">
                <Icon :name="isDarkMode ? 'sun' : 'moon'" :size="18" />
              </span>
              <span class="menu-item-text">{{ t.darkMode }}</span>
              <span class="menu-toggle-state">{{ isDarkMode ? 'Bật' : 'Tắt' }}</span>
            </button>
            <button type="button" class="menu-row-item" @click="toggleLanguage">
              <span class="menu-item-icon badge-lang"><Icon name="chat" :size="18" /></span>
              <span class="menu-item-text">{{ t.language }}</span>
              <span class="menu-toggle-state">{{ locale === 'vi' ? 'Tiếng Việt' : 'English' }}</span>
            </button>
            <button type="button" class="menu-row-item" @click="navigateTo('/settings')">
              <span class="menu-item-icon badge-settings"><Icon name="settings" :size="18" /></span>
              <span class="menu-item-text">{{ t.navSettings }}</span>
            </button>
          </div>
        </div>

        <!-- Section 3: Messenger Features (Mở rộng dần sau này) -->
        <div class="menu-section">
          <span class="menu-section-label">{{ t.messengerFeatures }}</span>
          <div class="menu-items-group">
            <div class="menu-row-item disabled-feature">
              <span class="menu-item-icon badge-status"><span class="status-dot" /></span>
              <span class="menu-item-text">{{ t.activeStatus }}</span>
              <span class="menu-badge-green">{{ t.activeStatusOn }}</span>
            </div>
            <div class="menu-row-item disabled-feature">
              <span class="menu-item-icon badge-archive"><Icon name="archive" :size="18" /></span>
              <span class="menu-item-text">{{ t.archivedChats }}</span>
              <span class="menu-badge-muted">{{ t.noConversationsYet }}</span>
            </div>
          </div>
        </div>

        <!-- Logout -->
        <div class="menu-logout-wrap">
          <button type="button" class="menu-logout-btn" @click="handleLogout">
            <Icon name="log-out" :size="18" />
            <span>{{ t.logOut }}</span>
          </button>
        </div>
      </div>
    </BottomSheet>
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

.rail-filter {
  padding: var(--space-xs) var(--space-md);
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
}

.rail-filter input {
  width: 100%;
  min-height: 40px;
  padding: 8px 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
  font: inherit;
  font-size: 14px;
  box-sizing: border-box;
}

.rail-filter input:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.chat-mark-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: var(--radius-md);
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  touch-action: manipulation;
  outline: none;
  transition: transform var(--duration-short) var(--ease-standard);
}

.chat-mark-btn:hover {
  transform: scale(1.08);
}

.chat-mark-btn:active {
  transform: scale(0.95);
}

.chat-mark-btn:focus:not(:focus-visible) {
  outline: none;
}

.chat-mark-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
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
  font-size: 0.875rem;
  letter-spacing: -0.02em;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  transition:
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard);
}

.chat-mark-btn:hover .chat-mark,
.chat-mark-btn:active .chat-mark {
  background: var(--accent);
  color: var(--on-accent);
}

.app-menu-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: 0 var(--space-xs) var(--space-sm);
}

.menu-profile-card {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-xl);
  background: var(--surface-card);
  color: var(--ink);
  cursor: pointer;
  text-align: left;
  transition: background var(--duration-short) var(--ease-standard);
}

.menu-profile-card:hover {
  background: var(--surface-soft);
}

.menu-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-pill);
  background: linear-gradient(135deg, #0084ff 0%, #0099ff 100%);
  color: #ffffff;
  font-size: 18px;
  font-weight: 700;
  flex-shrink: 0;
}

.menu-profile-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.menu-profile-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-profile-email {
  font-size: 12px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin-top: 3px;
  font-size: 11px;
  color: #22c55e;
  font-weight: 500;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-pill);
  background: #22c55e;
  flex-shrink: 0;
}

.menu-profile-arrow {
  color: var(--muted);
  flex-shrink: 0;
}

.menu-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.menu-section-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--muted);
  padding: 0 4px;
}

.menu-items-group {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-lg);
  background: var(--surface-card);
  overflow: hidden;
}

.menu-row-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: 11px 14px;
  border: none;
  border-bottom: 1px solid var(--hairline);
  background: transparent;
  color: var(--ink);
  font-size: 14px;
  cursor: pointer;
  text-align: left;
  transition: background var(--duration-short) var(--ease-standard);
}

.menu-row-item:last-child {
  border-bottom: none;
}

.menu-row-item:hover:not(.disabled-feature) {
  background: var(--surface-soft);
}

.menu-row-item.disabled-feature {
  cursor: default;
  opacity: 0.85;
}

.menu-item-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.badge-folder { background: #e0f2fe; color: #0284c7; }
.badge-photos { background: #fae8ff; color: #a21caf; }
.badge-trash { background: #fee2e2; color: #dc2626; }
.badge-theme { background: #fef3c7; color: #d97706; }
.badge-lang { background: #e0e7ff; color: #4338ca; }
.badge-settings { background: var(--surface-soft); color: var(--ink); }
.badge-status { background: #dcfce7; color: #16a34a; }
.badge-archive { background: var(--surface-soft); color: var(--muted); }

:global([data-theme='dark']) .badge-folder { background: rgba(2, 132, 199, 0.22); color: #38bdf8; }
:global([data-theme='dark']) .badge-photos { background: rgba(162, 28, 175, 0.22); color: #e879f9; }
:global([data-theme='dark']) .badge-trash { background: rgba(220, 38, 38, 0.22); color: #f87171; }
:global([data-theme='dark']) .badge-theme { background: rgba(217, 119, 6, 0.22); color: #fbbf24; }
:global([data-theme='dark']) .badge-lang { background: rgba(67, 56, 202, 0.22); color: #818cf8; }
:global([data-theme='dark']) .badge-status { background: rgba(22, 163, 74, 0.22); color: #4ade80; }

.menu-item-text {
  flex: 1;
  font-weight: 500;
}

.menu-toggle-state {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
}

.menu-badge-green {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  background: #dcfce7;
  color: #16a34a;
  font-weight: 600;
}

.menu-badge-muted {
  font-size: 11px;
  color: var(--muted);
}

.menu-logout-wrap {
  margin-top: 4px;
}

.menu-logout-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  width: 100%;
  padding: 10px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--danger, #ef4444);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background var(--duration-short) var(--ease-standard);
}

.menu-logout-btn:hover {
  background: rgba(239, 68, 68, 0.08);
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

.avatar-wrap {
  position: relative;
  display: inline-flex;
  flex-shrink: 0;
}

.online-indicator {
  position: absolute;
  right: 1px;
  bottom: 1px;
  width: 12px;
  height: 12px;
  border-radius: var(--radius-pill);
  background: #22c55e;
  border: 2px solid var(--canvas);
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
  font-weight: 600;
  font-size: 16px;
}

.avatar-color-a { background: #e0e7ff; color: #3730a3; }
.avatar-color-b { background: #fce7f3; color: #9d174d; }
.avatar-color-c { background: #dcfce7; color: #166534; }
.avatar-color-d { background: #ffedd5; color: #9a3412; }
.avatar-color-e { background: #cffafe; color: #155e75; }
.avatar-color-f { background: #f3e8ff; color: #6b21a8; }

.rail-empty {
  margin: 0;
  padding: var(--space-md);
  color: var(--muted);
  font-size: 13px;
  text-align: center;
}

.rail-empty-query {
  margin: 0;
}

.rail-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xl) var(--space-sm);
}

.rail-empty-title {
  margin: 0;
  color: var(--ink);
  font-size: 15px;
  font-weight: 600;
}

.rail-empty-hint {
  margin: 0 0 var(--space-xs);
  color: var(--muted);
  font-size: 13px;
  line-height: 1.4;
  max-width: 260px;
}

.rail-start-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-pill);
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
  gap: var(--space-xs);
  min-height: 56px;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
}

.thread-peer-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.thread-peer-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
  overflow: hidden;
}

.thread-peer-meta h2 {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
  line-height: 1.25;
}

.thread-status {
  font-size: 11px;
  color: #22c55e;
  font-weight: 500;
  line-height: 1.25;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.thread-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
  font-weight: 600;
  font-size: 14px;
}

.thread-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.thread-actions .icon-btn {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  color: #0084ff;
}

.back-btn {
  display: inline-flex;
  flex-shrink: 0;
}

.chat-search {
  display: flex;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-md);
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
}

.chat-search input {
  min-width: 0;
  flex: 1;
  min-height: 40px;
  padding: 8px 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--ink);
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
  gap: 2px;
  padding: var(--space-md);
}

.message-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  width: 100%;
}

.message-row.outgoing {
  justify-content: flex-end;
}

.message-row.cluster-start,
.message-row.cluster-single {
  margin-top: 8px;
}

.message-row:first-child {
  margin-top: 0;
}

.row-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 2px;
}

.row-avatar-spacer {
  width: 28px;
  flex-shrink: 0;
}

.message-bubble {
  max-width: min(74%, 480px);
  padding: 8px 14px;
  border-radius: 18px;
  background: var(--surface-card);
  color: var(--ink);
  font-size: 15px;
  line-height: 1.36;
  word-break: break-word;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.message-bubble.outgoing {
  background: linear-gradient(135deg, #0084ff 0%, #0099ff 100%);
  color: #ffffff;
}

/* Incoming bubble corners */
.message-row:not(.outgoing).cluster-single .message-bubble {
  border-radius: 18px;
}

.message-row:not(.outgoing).cluster-start .message-bubble {
  border-radius: 18px 18px 18px 4px;
}

.message-row:not(.outgoing).cluster-middle .message-bubble {
  border-radius: 4px 18px 18px 4px;
}

.message-row:not(.outgoing).cluster-last .message-bubble {
  border-radius: 4px 18px 18px 18px;
}

/* Outgoing bubble corners */
.message-row.outgoing.cluster-single .message-bubble {
  border-radius: 18px;
}

.message-row.outgoing.cluster-start .message-bubble {
  border-radius: 18px 18px 4px 18px;
}

.message-row.outgoing.cluster-middle .message-bubble {
  border-radius: 18px 4px 4px 18px;
}

.message-row.outgoing.cluster-last .message-bubble {
  border-radius: 18px 4px 18px 18px;
}

.message-bubble.has-like {
  background: transparent !important;
  box-shadow: none !important;
  padding: 2px 0 !important;
}

.message-bubble.has-like .bubble-time {
  display: none;
}

.like-bubble {
  font-size: 34px !important;
  line-height: 1.1 !important;
  display: inline-block;
  user-select: none;
  animation: messenger-pop 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes messenger-pop {
  0% { transform: scale(0.6); opacity: 0; }
  80% { transform: scale(1.18); }
  100% { transform: scale(1); opacity: 1; }
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
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-lg);
  background: rgba(0, 0, 0, 0.04);
  color: inherit;
  cursor: pointer;
}

.message-bubble.outgoing .attachment-card {
  border: 1px solid rgba(255, 255, 255, 0.32);
  background: rgba(255, 255, 255, 0.16);
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
  gap: 6px;
  padding: 8px var(--space-md) calc(8px + env(safe-area-inset-bottom));
  border-top: 1px solid var(--hairline);
  background: var(--canvas);
}

.attach-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: #0084ff;
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition: background var(--duration-short) var(--ease-standard);
}

.attach-btn:hover {
  background: rgba(0, 132, 255, 0.08);
}

.attach-btn:focus-visible {
  outline: 2px solid #0084ff;
  outline-offset: 2px;
}

.chat-composer textarea {
  min-height: 38px;
  max-height: 120px;
  flex: 1;
  padding: 8px 14px;
  resize: none;
  border: 1px solid var(--hairline);
  border-radius: 20px;
  background: var(--surface-soft);
  color: var(--ink);
  font: inherit;
  font-size: 15px;
  line-height: 1.35;
}

.chat-composer textarea:focus-visible {
  outline: 2px solid #0084ff;
  outline-offset: -1px;
}

.like-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  border: none;
  border-radius: var(--radius-pill);
  background: transparent;
  color: #0084ff;
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition:
    transform var(--duration-short) var(--ease-standard),
    background var(--duration-short) var(--ease-standard);
}

.like-btn:hover {
  transform: scale(1.15);
  background: rgba(0, 132, 255, 0.08);
}

.like-btn:active {
  transform: scale(0.92);
}

.like-btn:focus-visible {
  outline: 2px solid #0084ff;
  outline-offset: 2px;
}

.send-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  border: none;
  border-radius: var(--radius-pill);
  background: #0084ff;
  color: #ffffff;
  opacity: 0.4;
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition:
    opacity var(--duration-short) var(--ease-standard),
    transform var(--duration-short) var(--ease-standard);
}

.send-btn.active {
  opacity: 1;
  transform: scale(1.05);
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

  .thread-header {
    padding: calc(6px + env(safe-area-inset-top)) 10px 6px;
    gap: 6px;
    min-height: 52px;
  }

  .thread-peer-info {
    gap: 8px;
  }

  .thread-avatar {
    width: 34px;
    height: 34px;
    font-size: 13px;
  }

  .thread-peer-meta h2 {
    font-size: 14px;
  }

  .thread-status {
    font-size: 10.5px;
  }

  .thread-actions {
    gap: 2px;
  }

  .thread-actions .icon-btn {
    width: 34px;
    height: 34px;
  }
}

</style>
