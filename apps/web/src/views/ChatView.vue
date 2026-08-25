<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, uploadToPresigned, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import Icon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import type { ChatConversation, ChatMessage, DownloadURL, UploadSession } from '@/api/types'

const ui = useUiStore()

const conversations = ref<ChatConversation[]>([])
const selectedId = ref<string | null>(null)
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const newTitle = ref('')
const loading = ref(false)
const sending = ref(false)
const error = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const selectedConversation = computed(() => conversations.value.find((c) => c.id === selectedId.value) ?? null)

async function loadConversations() {
  error.value = ''
  loading.value = true
  try {
    const out = await api<{ conversations: ChatConversation[] }>('/chat/conversations')
    conversations.value = out.conversations
    if (!selectedId.value && out.conversations[0]) {
      selectedId.value = out.conversations[0].id
      await loadMessages(out.conversations[0].id)
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
  await loadMessages(id)
}

async function loadMessages(id = selectedId.value) {
  if (!id) return
  error.value = ''
  try {
    const out = await api<{ messages: ChatMessage[] }>(`/chat/conversations/${id}/messages`)
    messages.value = out.messages
  } catch (e) {
    error.value = formatApiError(e, 'Failed to load messages')
  }
}

async function sendText() {
  if (!selectedId.value || !draft.value.trim()) return
  sending.value = true
  error.value = ''
  try {
    const message = await api<ChatMessage>(`/chat/conversations/${selectedId.value}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body: draft.value.trim() }),
    })
    messages.value = [...messages.value, message]
    draft.value = ''
    await loadConversations()
    selectedId.value = message.conversationId
  } catch (e) {
    error.value = formatApiError(e, 'Failed to send message')
  } finally {
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

onMounted(loadConversations)
</script>

<template>
  <div class="chat-page">
    <h1 class="page-title desktop-only">Chat</h1>
    <p v-if="error" class="error" role="alert">{{ error }}</p>

    <div class="chat-layout">
      <aside class="conversation-list" aria-label="Conversations">
        <form class="new-chat" @submit.prevent="createConversation">
          <input v-model="newTitle" type="text" placeholder="New chat title" aria-label="New chat title" />
          <button class="btn ink" type="submit">New</button>
        </form>
        <button
          v-for="conv in conversations"
          :key="conv.id"
          type="button"
          class="conversation-row"
          :class="{ active: conv.id === selectedId }"
          @click="selectConversation(conv.id)"
        >
          <span class="conversation-title">{{ conv.title || 'Untitled chat' }}</span>
          <span class="conversation-date">{{ new Date(conv.updatedAt).toLocaleDateString() }}</span>
        </button>
      </aside>

      <section class="message-thread" aria-live="polite">
        <header class="thread-header">
          <h2>{{ selectedConversation?.title || 'Chat' }}</h2>
          <button class="btn ghost" type="button" :disabled="!selectedId" @click="loadMessages()">Refresh</button>
        </header>

        <div v-if="messages.length" class="message-list">
          <article v-for="message in messages" :key="message.id" class="message-bubble">
            <p v-if="message.body">{{ message.body }}</p>
            <button
              v-for="attachment in message.attachments"
              :key="attachment.id"
              type="button"
              class="attachment-card"
              @click="openAttachment(attachment.fileId)"
            >
              <Icon :name="attachment.mimeType.startsWith('image/') ? 'image' : 'file'" :size="18" />
              <span>{{ attachment.name }}</span>
              <span class="attachment-size">{{ formatBytes(attachment.sizeBytes) }}</span>
            </button>
          </article>
        </div>
        <EmptyState v-else title="No messages yet" description="Start the conversation with a message or an image." icon="chat" />

        <form class="chat-composer" @submit.prevent="sendText">
          <button type="button" class="btn icon-only" aria-label="Attach file" :disabled="!selectedId || sending" @click="triggerAttachment">
            <Icon name="plus" :size="18" />
          </button>
          <label class="sr-only" for="chat-message">Message</label>
          <textarea id="chat-message" v-model="draft" rows="1" placeholder="Message" :disabled="!selectedId || sending" />
          <button class="btn ink" type="submit" :disabled="!selectedId || sending || !draft.trim()">Send</button>
          <label class="sr-only" for="chat-attachment">Attachment</label>
          <input id="chat-attachment" ref="fileInputRef" type="file" class="sr-only" @change="onAttachmentChange" />
        </form>
      </section>
    </div>
  </div>
</template>

<style scoped>
.chat-page {
  min-height: 100%;
}

.chat-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-md);
}

.conversation-list,
.message-thread {
  min-width: 0;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-xl);
  background: var(--canvas);
}

.conversation-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-sm);
}

.new-chat {
  display: flex;
  gap: var(--space-xs);
}

.new-chat input {
  min-width: 0;
  flex: 1;
}

.conversation-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-xxs);
  min-height: var(--touch-min);
  padding: var(--space-sm);
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
}

.conversation-row.active {
  border-color: var(--accent);
  background: var(--accent-soft);
}

.conversation-title {
  font-weight: 600;
}

.conversation-date,
.attachment-size {
  color: var(--muted);
  font-size: 0.8rem;
}

.message-thread {
  display: flex;
  min-height: 560px;
  flex-direction: column;
  overflow: hidden;
}

.thread-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  padding: var(--space-sm);
  border-bottom: 1px solid var(--hairline);
}

.thread-header h2 {
  margin: 0;
  color: var(--ink);
  font-size: 1rem;
}

.message-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: var(--space-sm);
  overflow-y: auto;
  padding: var(--space-md);
}

.message-bubble {
  align-self: flex-end;
  max-width: min(78%, 520px);
  padding: var(--space-sm);
  border-radius: var(--radius-xl) var(--radius-xl) var(--radius-sm) var(--radius-xl);
  background: var(--accent);
  color: var(--on-accent);
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
  border: 1px solid color-mix(in srgb, var(--on-accent) 24%, transparent);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--on-accent) 12%, transparent);
  color: inherit;
  cursor: pointer;
}

.chat-composer {
  display: flex;
  align-items: flex-end;
  gap: var(--space-xs);
  padding: var(--space-sm);
  border-top: 1px solid var(--hairline);
  background: var(--canvas);
}

.chat-composer textarea {
  min-height: var(--touch-min);
  max-height: 140px;
  flex: 1;
  resize: vertical;
}

.desktop-only {
  display: none;
}

@media (min-width: 900px) {
  .chat-layout {
    grid-template-columns: 280px minmax(0, 1fr);
  }

  .desktop-only {
    display: block;
  }
}
</style>
