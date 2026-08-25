<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, uploadToPresigned, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import Icon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import LoadingSkeletonThread from '@/components/LoadingSkeletonThread.vue'
import type { ChatAttachment, ChatConversation, ChatMessage, DownloadURL, UploadSession } from '@/api/types'

const router = useRouter()
const ui = useUiStore()

const conversations = ref<ChatConversation[]>([])
const selectedId = ref<string | null>(null)
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const searchQuery = ref('')
const searchResults = ref<ChatMessage[] | null>(null)
const media = ref<ChatAttachment[]>([])
const newTitle = ref('')
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
const visibleMessages = computed(() => searchResults.value ?? messages.value)
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

function backToVault() {
  void router.push('/')
}

async function loadConversations() {
  error.value = ''
  loading.value = true
  try {
    const out = await api<{ conversations: ChatConversation[] }>('/chat/conversations?includePreview=true')
    conversations.value = out.conversations
    if (!selectedId.value && out.conversations[0]) {
      await selectConversation(out.conversations[0].id)
    }
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load chats')
  } finally {
    loading.value = false
  }
}

async function createConversation() {
  const title = newTitle.value.trim() || 'New chat'
  error.value = ''
  try {
    const conv = await api<ChatConversation>('/chat/conversations', {
      method: 'POST',
      body: JSON.stringify({ title }),
    })
    conversations.value = [conv, ...conversations.value]
    newTitle.value = ''
    await selectConversation(conv.id)
  } catch (e) {
    error.value = formatApiError(e, 'Failed to create chat')
  }
}

async function selectConversation(id: string) {
  selectedId.value = id
  searchQuery.value = ''
  searchResults.value = null
  loadingThread.value = true
  await Promise.all([loadMessages(id), loadMedia(id)])
  focusComposer()
}

async function loadMessages(id = selectedId.value) {
  if (!id) return
  error.value = ''
  loadingThread.value = true
  try {
    const out = await api<{ messages: ChatMessage[]; hasMore?: boolean; nextBefore?: string }>(
      `/chat/conversations/${id}/messages?limit=50`,
    )
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
  try {
    const out = await api<{ media: ChatAttachment[] }>(`/chat/conversations/${id}/media`)
    media.value = out.media
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load media')
  }
}

async function sendText() {
  if (!selectedId.value || !draft.value.trim()) return
  const conversationId = selectedId.value
  const body = draft.value.trim()
  const optimistic: ChatMessage = {
    id: `pending-${Date.now()}`,
    conversationId,
    body,
    createdAt: new Date().toISOString(),
    attachments: [],
  }
  pendingMessage.value = optimistic
  draft.value = ''
  autoGrow()
  error.value = ''
  sending.value = true
  try {
    const message = await api<ChatMessage>(`/chat/conversations/${conversationId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    })
    messages.value = [...messages.value, message]
    searchResults.value = null
    pendingMessage.value = null
    showJump.value = false
    await loadConversations()
    selectedId.value = message.conversationId
  } catch (e) {
    draft.value = draft.value ? draft.value : body
    error.value = formatApiError(e, 'Failed to send message')
  } finally {
    pendingMessage.value = null
    sending.value = false
  }
}

function triggerAttachment() {
  fileInputRef.value?.click()
}

async function onAttachmentChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !selectedId.value) return
  sending.value = true
  error.value = ''
  try {
    const session = await api<UploadSession>('/chat/attachments/upload-sessions', {
      method: 'POST',
      body: JSON.stringify({
        conversationId: selectedId.value,
        name: file.name,
        size: file.size,
        contentType: file.type || 'application/octet-stream',
      }),
    })
    await uploadToPresigned(session.uploadUrl, file, file.type || 'application/octet-stream')
    const message = await api<ChatMessage>(`/chat/attachments/${session.fileId}/complete`, {
      method: 'POST',
      body: JSON.stringify({ conversationId: selectedId.value, body: draft.value.trim() }),
    })
    messages.value = [...messages.value, message]
    searchResults.value = null
    await loadMedia()
    draft.value = ''
    ui.showToast('Attachment sent')
  } catch (e) {
    error.value = formatApiError(e, 'Failed to send attachment')
  } finally {
    sending.value = false
  }
}

async function openAttachment(fileId: string) {
  try {
    const out = await api<DownloadURL>(`/files/${fileId}/download`)
    window.open(out.downloadUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, 'Download failed')
  }
}

async function openInlineImage(attachment: ChatAttachment) {
  error.value = ''
  try {
    const out = await api<DownloadURL>(`/files/${attachment.fileId}/download`)
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

onMounted(loadConversations)

watch(visibleMessages, () => {
  void nextTick(() => {
    const el = threadBodyRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
})
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
      </header>
      <form class="new-chat" @submit.prevent="createConversation">
        <label class="sr-only" for="new-chat-title">New chat title</label>
        <input id="new-chat-title" v-model="newTitle" type="text" placeholder="New chat title" />
        <button class="compose-btn" type="submit" aria-label="Create chat">
          <Icon name="pencil" :size="18" />
        </button>
      </form>
      <div v-if="error && !inThread" class="alert" role="alert">{{ error }}</div>
      <nav class="conversation-list">
        <button
          v-for="conv in conversations"
          :key="conv.id"
          type="button"
          class="conversation-row"
          :class="{ active: conv.id === selectedId }"
          :aria-current="conv.id === selectedId ? 'true' : undefined"
          @click="selectConversation(conv.id)"
        >
          <span class="avatar" aria-hidden="true">{{ conv.title.slice(0, 1).toUpperCase() }}</span>
          <span class="conversation-meta">
            <span class="conversation-title">{{ conv.title || 'Untitled chat' }}</span>
            <span class="conversation-preview">{{ conv.preview?.body || (conv.preview?.attachments?.length ? 'Attachment' : 'No messages yet') }}</span>
            <span class="conversation-date">{{ new Date(conv.updatedAt).toLocaleDateString() }}</span>
          </span>
        </button>
      </nav>
    </aside>

    <section class="message-thread" aria-live="polite">
      <template v-if="selectedConversation">
        <header class="thread-header">
          <button type="button" class="icon-btn back-btn" aria-label="Back" @click="backToRail">
            <Icon name="close" :size="20" />
          </button>
          <span class="thread-avatar" aria-hidden="true">{{ selectedConversation.title.slice(0, 1).toUpperCase() }}</span>
          <h2>{{ selectedConversation.title || 'Untitled chat' }}</h2>
          <button class="ghost-btn" type="button" @click="loadMessages()">Refresh</button>
        </header>

        <form class="chat-search" @submit.prevent="searchMessages">
          <label class="sr-only" for="chat-search">Search messages</label>
          <input id="chat-search" v-model="searchQuery" type="search" placeholder="Search in this chat" />
          <button class="ghost-btn" type="submit" :disabled="searchQuery.trim().length < 2">Search</button>
          <button v-if="searchResults" class="ghost-btn" type="button" @click="clearSearch">Clear</button>
        </form>

        <p v-if="error && inThread" class="alert" role="alert">{{ error }}</p>

        <div ref="threadBodyRef" class="message-body" @scroll.passive="onThreadScroll">
          <LoadingSkeletonThread v-if="loadingThread" />
          <p v-else-if="loadingOlder" class="loading-older">Loading older…</p>
          <div v-if="!loadingThread && visibleMessages.length" class="message-list">
            <TransitionGroup name="msg">
              <template v-for="(message, index) in visibleMessages" :key="message.id">
                <div v-if="shouldShowDate(index)" class="day-separator">
                  <span>{{ formatDayLabel(message.createdAt) }}</span>
                </div>
                <article class="message-bubble">
                  <p v-if="message.body">{{ message.body }}</p>
                  <template v-for="attachment in message.attachments" :key="attachment.id">
                    <button
                      v-if="attachment.thumbnailUrl && attachment.mimeType.startsWith('image/')"
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
                      @click="openAttachment(attachment.fileId)"
                    >
                      <Icon :name="attachment.mimeType.startsWith('image/') ? 'image' : 'file'" :size="18" />
                      <span class="attachment-name">{{ attachment.name }}</span>
                      <span class="attachment-size">{{ formatBytes(attachment.sizeBytes) }}</span>
                    </button>
                  </template>
                  <span class="bubble-time">{{ formatTime(message.createdAt) }}</span>
                </article>
              </template>
            </TransitionGroup>
            <Transition name="msg">
              <article v-if="pendingMessage" class="message-bubble pending" aria-live="polite">
                <p>{{ pendingMessage.body }}</p>
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
        @click="openAttachment(item.fileId)"
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

.new-chat {
  display: flex;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
}

.new-chat input {
  min-width: 0;
  flex: 1;
  min-height: 44px;
  padding: 12px 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  font: inherit;
}

.new-chat input:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
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
  min-height: 32px;
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

@media (min-width: 1024px) {
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
