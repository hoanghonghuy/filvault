import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { API_BASE, api, getAccessToken, refreshAccessToken } from '@/api/client'
import type { ChatConversation, ChatMessage, ChatMessageReaction } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useCallStore } from '@/stores/call'
import { generateUUID } from '@/lib/uuid'

type ConnectionState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'offline'
type ChatEvent = { type: string; payload: string }

export const useChatStore = defineStore('chat', () => {
  const auth = useAuthStore()
  const conversations = ref<ChatConversation[]>([])
  const messages = ref<Record<string, ChatMessage[]>>({})
  const selectedId = ref<string | null>(null)
  const connectionState = ref<ConnectionState>('idle')
  const lastEventId = ref(0)
  const isLoadingConversations = ref(false)
  const seenEventIds = new Set<number>()
  let eventController: AbortController | null = null
  let reconnectTimer: number | null = null
  const typingUsers = ref<Record<string, { userId: string; userName: string; timer?: number }>>({})
  let lastTypingTime = 0

  const selectedConversation = computed(() =>
    conversations.value.find((conversation) => conversation.id === selectedId.value) ?? null,
  )

  async function loadConversations(): Promise<void> {
    if (isLoadingConversations.value) return
    isLoadingConversations.value = true
    try {
      const out = await api<{ conversations: ChatConversation[] }>('/chat/conversations?includePreview=true')
      conversations.value = out.conversations
    } finally {
      isLoadingConversations.value = false
    }
  }

  async function openConversation(id: string): Promise<void> {
    selectedId.value = id
    const out = await api<{ messages: ChatMessage[] }>(`/chat/conversations/${id}/messages?limit=50`)
    if (selectedId.value !== id) return
    messages.value[id] = out.messages
  }

  async function createDirectConversation(recipientEmail: string): Promise<ChatConversation> {
    const conversation = await api<ChatConversation>('/chat/direct-conversations', {
      method: 'POST',
      body: JSON.stringify({ recipientEmail }),
    })
    conversations.value = [conversation, ...conversations.value.filter((item) => item.id !== conversation.id)]
    await openConversation(conversation.id)
    return conversation
  }

  async function sendMessage(
    body: string,
    clientMessageId = generateUUID(),
    conversationId = selectedId.value,
  ): Promise<ChatMessage> {
    if (!conversationId) throw new Error('No conversation selected')
    const message = await api<ChatMessage>(`/chat/conversations/${conversationId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body, clientMessageId }),
    })
    messages.value[conversationId] = reconcileMessage(messages.value[conversationId] ?? [], message)
    return message
  }

  function reconcileMessage(current: ChatMessage[], incoming: ChatMessage): ChatMessage[] {
    const index = current.findIndex(
      (message) =>
        message.id === incoming.id ||
        (incoming.clientMessageId && message.clientMessageId === incoming.clientMessageId),
    )
    if (index < 0) return [...current, incoming]
    const next = [...current]
    next[index] = incoming
    return next
  }

  function applyEvent(event: ChatEvent): void {
    if (event.type === 'call.signal') {
      try {
        const signalData =
          typeof event.payload === 'string' ? JSON.parse(event.payload) : event.payload
        useCallStore().handleSignal(signalData)
      } catch {
        // ignore
      }
      return
    }
    if (event.type === 'message.read') {
      try {
        const data =
          typeof event.payload === 'string' ? JSON.parse(event.payload) : event.payload
        const { conversationId, userId, lastReadMessageId, lastReadAt } = data
        const conv = conversations.value.find((c) => c.id === conversationId)
        if (conv && conv.peer?.id === userId) {
          conv.peerLastReadMessageId = lastReadMessageId
          conv.peerLastReadAt = lastReadAt
        }
      } catch {
        // ignore
      }
      return
    }
    if (event.type === 'typing.indicator') {
      try {
        const data =
          typeof event.payload === 'string' ? JSON.parse(event.payload) : event.payload
        const { conversationId, userId, userName, typing } = data
        if (!conversationId || !userId || userId === auth.user?.id) {
          return
        }
        if (typing) {
          if (typingUsers.value[conversationId]?.timer) {
            window.clearTimeout(typingUsers.value[conversationId].timer)
          }
          const timer = window.setTimeout(() => {
            delete typingUsers.value[conversationId]
          }, 4000)
          typingUsers.value[conversationId] = { userId, userName, timer }
        } else {
          if (typingUsers.value[conversationId]?.timer) {
            window.clearTimeout(typingUsers.value[conversationId].timer)
          }
          delete typingUsers.value[conversationId]
        }
      } catch {
        // ignore
      }
      return
    }
    if (event.type === 'presence.changed') {
      try {
        const data =
          typeof event.payload === 'string' ? JSON.parse(event.payload) : event.payload
        const { userId, status, lastSeenAt } = data
        if (!userId || userId === auth.user?.id) return
        conversations.value.forEach((conv) => {
          if (conv.peer?.id === userId) {
            conv.peerStatus = status
            if (lastSeenAt) {
              conv.peerLastSeenAt = lastSeenAt
              if (conv.peer) {
                conv.peer.lastSeenAt = lastSeenAt
              }
            }
          }
        })
      } catch {
        // ignore
      }
      return
    }
    if (event.type === 'message.reaction') {
      try {
        const data =
          typeof event.payload === 'string' ? JSON.parse(event.payload) : event.payload
        const { conversationId, messageId, reactions } = data
        const currentList = messages.value[conversationId]
        if (currentList) {
          const msg = currentList.find((m) => m.id === messageId)
          if (msg) {
            msg.reactions = reactions
          }
        }
      } catch {
        // ignore
      }
      return
    }
    try {
      const payload = JSON.parse(event.payload) as { aggregateId?: string }
      if (selectedId.value && payload.aggregateId) {
        const current = messages.value[selectedId.value] ?? []
        if (event.type === 'message.removed') {
          messages.value[selectedId.value] = current.map((message) =>
            message.id === payload.aggregateId
              ? { ...message, body: '', removedAt: new Date().toISOString() }
              : message,
          )
        }
        if (event.type === 'message.created' || event.type === 'message.edited') {
          void openConversation(selectedId.value)
        }
      }
      if (event.type === 'message.created') {
        void loadConversations()
      }
    } catch {
      // The durable event is still acknowledged by its sequence cursor.
    }
  }

  async function markAsRead(conversationId: string, messageId?: string): Promise<void> {
    const conv = conversations.value.find((c) => c.id === conversationId)
    if (conv) {
      conv.unreadCount = 0
    }
    try {
      await api(`/chat/conversations/${conversationId}/read`, {
        method: 'POST',
        body: JSON.stringify({ messageId: messageId ?? '' }),
      })
    } catch {
      // ignore
    }
  }

  async function sendTyping(conversationId: string, typing: boolean): Promise<void> {
    const now = Date.now()
    if (typing && now - lastTypingTime < 2000) return
    if (typing) lastTypingTime = now
    try {
      await api(`/chat/conversations/${conversationId}/typing`, {
        method: 'POST',
        body: JSON.stringify({ typing }),
      })
    } catch {
      // ignore
    }
  }

  async function toggleReaction(
    conversationId: string,
    messageId: string,
    reaction: string,
  ): Promise<ChatMessageReaction[]> {
    const res = await api<{ reactions: ChatMessageReaction[] }>(
      `/chat/conversations/${conversationId}/messages/${messageId}/reactions`,
      {
        method: 'POST',
        body: JSON.stringify({ reaction }),
      },
    )
    const currentList = messages.value[conversationId]
    if (currentList) {
      const msg = currentList.find((m) => m.id === messageId)
      if (msg) {
        msg.reactions = res.reactions
      }
    }
    return res.reactions
  }

  function stopEvents(): void {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    eventController?.abort()
    eventController = null
    seenEventIds.clear()
    connectionState.value = 'idle'
  }

  function markOffline(): void {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    eventController?.abort()
    eventController = null
    connectionState.value = 'offline'
  }

  function markOnline(): void {
    if (connectionState.value === 'offline') connectEvents()
  }

  function scheduleReconnect(): void {
    if (reconnectTimer !== null) return
    connectionState.value = navigator.onLine ? 'reconnecting' : 'offline'
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connectEvents()
    }, 1000 + Math.round(Math.random() * 2000))
  }

  function connectEvents(): void {
    eventController?.abort()
    eventController = null
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    const token = getAccessToken()
    if (!token) {
      connectionState.value = 'offline'
      return
    }
    connectionState.value = 'connecting'
    const controller = new AbortController()
    eventController = controller
    void readEvents(token, controller.signal)
  }

  async function readEvents(token: string, signal: AbortSignal): Promise<void> {
    try {
      const response = await fetch(`${API_BASE}/chat/events?after=${lastEventId.value}`, {
        headers: { Authorization: `Bearer ${token}`, Accept: 'text/event-stream' },
        signal,
      })
          if (response.status === 401 && !signal.aborted && (await refreshAccessToken())) {
            connectEvents()
            return
          }
          if (!response.ok || !response.body) throw new Error(`SSE failed (${response.status})`)
      connectionState.value = 'connected'
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (!signal.aborted) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const frames = buffer.split('\n\n')
        buffer = frames.pop() ?? ''
        for (const frame of frames) {
          const id = frame.match(/^id:\s*(\d+)$/m)?.[1]
          const sequence = id ? Number(id) : 0
          if (sequence > 0 && (seenEventIds.has(sequence) || sequence <= lastEventId.value)) continue
          if (sequence > 0) {
            seenEventIds.add(sequence)
            lastEventId.value = sequence
            if (seenEventIds.size > 1000) {
              const oldest = seenEventIds.values().next().value
              if (typeof oldest === 'number') seenEventIds.delete(oldest)
            }
          }
          const type = frame.match(/^event:\s*(.+)$/m)?.[1]
          const data = frame.match(/^data:\s*(.+)$/m)?.[1]
          if (type && data) applyEvent({ type, payload: data })
          if (type) {
            void loadConversations()
            if (selectedId.value) void openConversation(selectedId.value)
          }
        }
      }
      if (!signal.aborted) scheduleReconnect()
    } catch {
      if (!signal.aborted) scheduleReconnect()
    }
  }

  return {
    conversations,
    messages,
    selectedId,
    selectedConversation,
    connectionState,
    loadConversations,
    openConversation,
    createDirectConversation,
    sendMessage,
    applyEvent,
    connectEvents,
    stopEvents,
    markOffline,
    markOnline,
    typingUsers,
    markAsRead,
    sendTyping,
    toggleReaction,
  }
})

