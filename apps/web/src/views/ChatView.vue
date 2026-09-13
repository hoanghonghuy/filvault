<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, uploadToPresigned, formatBytes, normalizePresignedUrl } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useUiStore } from '@/stores/ui'
import { useChatStore } from '@/stores/chat'
import { useCallStore } from '@/stores/call'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/AppIcon.vue'
import EmptyState from '@/components/EmptyState.vue'
import MediaLightbox from '@/components/MediaLightbox.vue'
import ChatInlineImage from '@/components/ChatInlineImage.vue'
import LoadingSkeletonThread from '@/components/LoadingSkeletonThread.vue'
import LoadingSkeletonChatRail from '@/components/LoadingSkeletonChatRail.vue'
import UploadProgress from '@/components/UploadProgress.vue'
import BottomSheet from '@/components/BottomSheet.vue'
import EmojiPicker from '@/components/EmojiPicker.vue'
import ChatBubblePickerModal from '@/components/ChatBubblePickerModal.vue'
import { resolveContentType } from '@/lib/mimeIcon'
import { isHeic, getHeicDisplayUrl, convertHeicBlobToJpeg } from '@/lib/heic'
import { getBubbleStyle, activeBubbleStyleId, type ChatBubbleStyle } from '@/lib/chatBubbles'
import { userInitials } from '@/lib/userInitials'
import {
  isStickerMessage,
  parseStickerSymbol,
  formatStickerMessage,
  type Sticker,
} from '@/lib/stickers'
import { useI18n } from '@/lib/i18n'
import { useTheme } from '@/lib/theme'
import { generateUUID } from '@/lib/uuid'
import type { UploadItemStatus } from '@/lib/uploadQueue'
import { useLongPress } from '@/lib/useLongPress'
import { resolveWallpaperTheme, type WallpaperTheme } from '@/lib/wallpaperPalette'
import type { ChatAttachment, ChatConversation, ChatMessage, DownloadURL, Timeline, TimelineItem, UploadSession } from '@/api/types'

const router = useRouter()
const route = useRoute()
const ui = useUiStore()
const chatStore = useChatStore()
const callStore = useCallStore()
const auth = useAuthStore()
const { t, locale, setLocale } = useI18n()
const { resolvedIsDark } = useTheme()

const conversations = ref<ChatConversation[]>([])
const selectedId = ref<string | null>((route.params.id as string) || null)
const selectedConversation = computed(() => conversations.value.find((c) => c.id === selectedId.value) ?? null)
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const searchQuery = ref('')
const searchResults = ref<ChatMessage[] | null>(null)
const media = ref<ChatAttachment[]>([])
const loading = ref(false)
const sendingText = ref(false)
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
const loadingThread = ref(Boolean(route.params.id))

const threadInfoOpen = ref(false)
const messageMenuOpen = ref(false)
const activeMessage = ref<ChatMessage | null>(null)
const convMenuOpen = ref(false)
const activeConv = ref<ChatConversation | null>(null)
const mutedConversations = ref<Record<string, boolean>>({})
const nicknames = ref<Record<string, string>>({})

const chatThemes = [
  { id: 'blue', name: 'Messenger Blue', color: '#0084ff', gradient: 'linear-gradient(135deg, #0084ff 0%, #0099ff 100%)' },
  { id: 'purple', name: 'Hoàng hôn Tím', color: '#8b5cf6', gradient: 'linear-gradient(135deg, #8b5cf6 0%, #ec4899 100%)' },
  { id: 'emerald', name: 'Ngọc lục bảo', color: '#10b981', gradient: 'linear-gradient(135deg, #059669 0%, #10b981 100%)' },
  { id: 'rose', name: 'Hồng ngọt ngào', color: '#f43f5e', gradient: 'linear-gradient(135deg, #f43f5e 0%, #fb7185 100%)' },
  { id: 'amber', name: 'Hổ phách ấm', color: '#f59e0b', gradient: 'linear-gradient(135deg, #d97706 0%, #f59e0b 100%)' },
  { id: 'cyan', name: 'Đại dương xanh', color: '#06b6d4', gradient: 'linear-gradient(135deg, #0891b2 0%, #06b6d4 100%)' },
  { id: 'indigo', name: 'Chàm huyền bí', color: '#6366f1', gradient: 'linear-gradient(135deg, #4f46e5 0%, #6366f1 100%)' },
  { id: 'coral', name: 'Cam san hô', color: '#f97316', gradient: 'linear-gradient(135deg, #ea580c 0%, #fb923c 100%)' },
  { id: 'slate', name: 'Than chì thanh lịch', color: '#64748b', gradient: 'linear-gradient(135deg, #475569 0%, #64748b 100%)' },
] as const
const activeTheme = ref<string>('blue')
const currentTheme = computed(() => chatThemes.find((entry) => entry.id === activeTheme.value) ?? chatThemes[0]!)
const activeWallpaperTheme = ref<WallpaperTheme | null>(null)

const bubblePickerOpen = ref(false)
const currentBubbleStyle = computed<ChatBubbleStyle>(() => getBubbleStyle(activeBubbleStyleId.value))

function outgoingBubbleStyle(message: ChatMessage) {
  if (message.senderId !== auth.user?.id) return undefined
  if (isStickerMessage(message.body) || message.body === '👍') return undefined
  if (currentBubbleStyle.value.id === 'default') return undefined
  return {
    background: currentBubbleStyle.value.bg,
    color: currentBubbleStyle.value.color,
    border: currentBubbleStyle.value.border || 'none',
  }
}

const threadThemeStyle = computed(() => {
  const theme = activeWallpaperTheme.value
  if (!theme) {
    return {
      '--chat-bubble-outgoing': currentTheme.value.gradient,
      '--chat-accent': currentTheme.value.color,
    }
  }

  return {
    '--chat-bubble-outgoing': theme.gradient,
    '--chat-accent': theme.primary,
    '--chat-accent-secondary': theme.secondary,
    '--chat-surface-tint': theme.surfaceTint,
    '--chat-border-tint': theme.borderTint,
    '--chat-header-bg': theme.headerBg,
    '--chat-composer-bg': theme.composerBg,
    '--chat-scrim-overlay': theme.scrimOverlay,
  }
})

export interface ChatWallpaperPreset {
  id: string
  name: string
  value: string
  preview: string
  isDark?: boolean
}

const CHAT_WALLPAPER_PRESETS: ChatWallpaperPreset[] = [
  {
    id: 'none',
    name: 'Mặc định',
    value: '',
    preview: 'var(--canvas)',
  },

  // --- Chế độ Sáng (Light Mode Presets) ---
  {
    id: 'doodle',
    name: 'Doodle họa tiết',
    value: `radial-gradient(circle at 50% 50%, rgba(99, 102, 241, 0.08) 0%, transparent 60%), radial-gradient(circle at 10% 20%, rgba(236, 72, 153, 0.08) 0%, transparent 40%), url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='%239C92AC' fill-opacity='0.09' fill-rule='evenodd'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/svg%3E")`,
    preview: 'linear-gradient(135deg, #e0e7ff 0%, #fce7f3 100%)',
  },
  {
    id: 'sakura',
    name: 'Hoa anh đào',
    value: 'radial-gradient(circle at 30% 20%, rgba(251, 113, 133, 0.25) 0%, transparent 50%), radial-gradient(circle at 70% 80%, rgba(192, 132, 252, 0.2) 0%, transparent 50%), linear-gradient(135deg, #fdf2f8 0%, #fae8ff 50%, #f5f3ff 100%)',
    preview: 'linear-gradient(135deg, #fdf2f8 0%, #f43f5e 100%)',
  },
  {
    id: 'sky_breeze',
    name: 'Mây trời sáng',
    value: 'radial-gradient(ellipse at top, rgba(186, 230, 253, 0.45) 0%, transparent 60%), radial-gradient(ellipse at bottom, rgba(224, 231, 255, 0.45) 0%, transparent 50%), linear-gradient(140deg, #f0f9ff 0%, #e0f2fe 50%, #eff6ff 100%)',
    preview: 'linear-gradient(135deg, #bae6fd 0%, #e0f2fe 100%)',
  },
  {
    id: 'matcha',
    name: 'Trà xanh Matcha',
    value: 'radial-gradient(circle at 15% 15%, rgba(134, 239, 172, 0.35) 0%, transparent 50%), radial-gradient(circle at 85% 85%, rgba(187, 247, 208, 0.4) 0%, transparent 50%), linear-gradient(135deg, #f0fdf4 0%, #dcfce7 50%, #f7fee7 100%)',
    preview: 'linear-gradient(135deg, #86efac 0%, #bbf7d0 100%)',
  },
  {
    id: 'terracotta',
    name: 'Gốm ấm áp',
    value: 'radial-gradient(circle at top right, rgba(253, 186, 116, 0.35) 0%, transparent 60%), radial-gradient(circle at bottom left, rgba(254, 215, 170, 0.4) 0%, transparent 50%), linear-gradient(135deg, #fff7ed 0%, #ffedd5 50%, #fef3c7 100%)',
    preview: 'linear-gradient(135deg, #fdba74 0%, #fed7aa 100%)',
  },
  {
    id: 'lavender_mist',
    name: 'Oải hương sương mai',
    value: 'radial-gradient(circle at 30% 20%, rgba(216, 180, 254, 0.35) 0%, transparent 60%), radial-gradient(circle at 70% 80%, rgba(233, 213, 255, 0.35) 0%, transparent 50%), linear-gradient(135deg, #faf5ff 0%, #f3e8ff 50%, #f5f3ff 100%)',
    preview: 'linear-gradient(135deg, #d8b4fe 0%, #e9d5ff 100%)',
  },
  {
    id: 'notebook_grid',
    name: 'Giấy kẻ ô',
    value: `radial-gradient(circle at 50% 50%, rgba(59, 130, 246, 0.05) 0%, transparent 70%), linear-gradient(#e2e8f0 1px, transparent 1px), linear-gradient(90deg, #e2e8f0 1px, #f8fafc 1px)`,
    preview: 'linear-gradient(135deg, #f1f5f9 0%, #e2e8f0 100%)',
  },
  {
    id: 'geometric_pastel',
    name: 'Hình học tối giản',
    value: 'radial-gradient(circle at 20% 80%, rgba(244, 114, 182, 0.2) 0%, transparent 40%), radial-gradient(circle at 80% 20%, rgba(96, 165, 250, 0.2) 0%, transparent 40%), linear-gradient(135deg, #fdf4ff 0%, #eff6ff 100%)',
    preview: 'linear-gradient(135deg, #f472b6 0%, #60a5fa 100%)',
  },

  // --- Chế độ Tối (Dark Mode Presets) ---
  {
    id: 'aurora',
    name: 'Cực quang xanh',
    value: 'radial-gradient(ellipse at 20% 20%, rgba(16, 185, 129, 0.35) 0%, transparent 50%), radial-gradient(ellipse at 80% 80%, rgba(59, 130, 246, 0.35) 0%, transparent 50%), linear-gradient(160deg, #091a24 0%, #0d2838 50%, #07131b 100%)',
    preview: 'linear-gradient(135deg, #0d2838 0%, #10b981 100%)',
    isDark: true,
  },
  {
    id: 'sunset',
    name: 'Hoàng hôn',
    value: 'radial-gradient(circle at top right, rgba(244, 63, 94, 0.3) 0%, transparent 60%), radial-gradient(circle at bottom left, rgba(251, 146, 60, 0.3) 0%, transparent 60%), linear-gradient(145deg, #1e112a 0%, #2d1b4e 50%, #140924 100%)',
    preview: 'linear-gradient(135deg, #2d1b4e 0%, #f43f5e 100%)',
    isDark: true,
  },
  {
    id: 'cosmos',
    name: 'Vũ trụ huyền ảo',
    value: 'radial-gradient(circle at 50% 30%, rgba(139, 92, 246, 0.35) 0%, transparent 60%), radial-gradient(circle at 80% 90%, rgba(236, 72, 153, 0.25) 0%, transparent 50%), linear-gradient(180deg, #0a0818 0%, #130e2e 50%, #070512 100%)',
    preview: 'linear-gradient(135deg, #130e2e 0%, #8b5cf6 100%)',
    isDark: true,
  },
  {
    id: 'emerald',
    name: 'Rừng nhiệt đới',
    value: 'radial-gradient(circle at 20% 80%, rgba(5, 150, 105, 0.3) 0%, transparent 60%), radial-gradient(circle at 80% 20%, rgba(52, 211, 153, 0.2) 0%, transparent 50%), linear-gradient(150deg, #061a14 0%, #0a2f24 50%, #04100c 100%)',
    preview: 'linear-gradient(135deg, #0a2f24 0%, #10b981 100%)',
    isDark: true,
  },
  {
    id: 'cyber',
    name: 'Đêm Neon',
    value: 'radial-gradient(circle at top left, rgba(6, 182, 212, 0.35) 0%, transparent 55%), radial-gradient(circle at bottom right, rgba(236, 72, 153, 0.35) 0%, transparent 55%), linear-gradient(135deg, #09090b 0%, #18181b 50%, #09090b 100%)',
    preview: 'linear-gradient(135deg, #06b6d4 0%, #ec4899 100%)',
    isDark: true,
  },
  {
    id: 'amoled_carbon',
    name: 'Lưới Carbon AMOLED',
    value: 'radial-gradient(circle at 50% 50%, rgba(255, 255, 255, 0.05) 0%, transparent 70%), linear-gradient(45deg, #121214 25%, transparent 25%), linear-gradient(-45deg, #121214 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #121214 75%), linear-gradient(-45deg, transparent 75%, #121214 75%), #070709',
    preview: 'linear-gradient(135deg, #09090b 0%, #27272a 100%)',
    isDark: true,
  },
  {
    id: 'obsidian_gold',
    name: 'Hắc diện kim',
    value: 'radial-gradient(circle at 75% 25%, rgba(234, 179, 8, 0.25) 0%, transparent 50%), radial-gradient(circle at 20% 80%, rgba(161, 98, 7, 0.2) 0%, transparent 55%), linear-gradient(150deg, #09090b 0%, #171510 50%, #090805 100%)',
    preview: 'linear-gradient(135deg, #171510 0%, #eab308 100%)',
    isDark: true,
  },
  {
    id: 'ocean_abyss',
    name: 'Vực thẳm đại dương',
    value: 'radial-gradient(ellipse at 50% 10%, rgba(14, 165, 233, 0.28) 0%, transparent 60%), radial-gradient(ellipse at 80% 90%, rgba(2, 132, 199, 0.22) 0%, transparent 50%), linear-gradient(165deg, #030712 0%, #082f49 50%, #020617 100%)',
    preview: 'linear-gradient(135deg, #082f49 0%, #0ea5e9 100%)',
    isDark: true,
  },
  {
    id: 'twilight_violet',
    name: 'Hoàng hôn Tím đêm',
    value: 'radial-gradient(circle at top left, rgba(168, 85, 247, 0.3) 0%, transparent 55%), radial-gradient(circle at bottom right, rgba(236, 72, 153, 0.25) 0%, transparent 55%), linear-gradient(140deg, #0f0728 0%, #1e1035 50%, #0b051b 100%)',
    preview: 'linear-gradient(135deg, #1e1035 0%, #a855f7 100%)',
    isDark: true,
  },
]

const STORAGE_KEY_WALLPAPERS = 'filvault.chat.wallpapers'

function loadSavedWallpapers(): Record<string, string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_WALLPAPERS)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveWallpapersToStorage(data: Record<string, string>) {
  try {
    localStorage.setItem(STORAGE_KEY_WALLPAPERS, JSON.stringify(data))
  } catch (e) {
    console.warn('Failed to save chat wallpapers to localStorage', e)
  }
}

export interface ChatCustomWallpaper {
  id: string
  url: string
  name: string
  source: 'uploaded' | 'chat_media' | 'storage'
  sourceId?: string
  createdAt: string
}

const STORAGE_KEY_CUSTOM_WALLPAPERS = 'filvault.chat.custom_wallpapers'

function loadSavedCustomWallpapers(): Record<string, ChatCustomWallpaper[]> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_CUSTOM_WALLPAPERS)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveCustomWallpapersToStorage(data: Record<string, ChatCustomWallpaper[]>) {
  try {
    localStorage.setItem(STORAGE_KEY_CUSTOM_WALLPAPERS, JSON.stringify(data))
  } catch (e) {
    console.warn('Failed to save custom wallpapers to localStorage', e)
  }
}

const chatWallpapers = ref<Record<string, string>>(loadSavedWallpapers())
const chatCustomWallpapers = ref<Record<string, ChatCustomWallpaper[]>>(loadSavedCustomWallpapers())
const wallpaperFileInputRef = ref<HTMLInputElement | null>(null)

const currentConversationCustomWallpapers = computed<ChatCustomWallpaper[]>(() => {
  if (!selectedConversation.value) return []
  return chatCustomWallpapers.value[selectedConversation.value.id] || []
})

function addCustomWallpaperToConversation(conversationId: string, item: ChatCustomWallpaper) {
  const current = chatCustomWallpapers.value[conversationId] || []
  const filtered = current.filter((w) => w.url !== item.url && w.id !== item.id)
  const nextList = [item, ...filtered].slice(0, 16)
  const nextMap = { ...chatCustomWallpapers.value, [conversationId]: nextList }
  chatCustomWallpapers.value = nextMap
  saveCustomWallpapersToStorage(nextMap)
}

function removeCustomWallpaperItem(conversationId: string, wallpaperId: string) {
  const current = chatCustomWallpapers.value[conversationId] || []
  const target = current.find((w) => w.id === wallpaperId)
  if (!target) return
  const nextList = current.filter((w) => w.id !== wallpaperId)
  const nextMap = { ...chatCustomWallpapers.value, [conversationId]: nextList }
  chatCustomWallpapers.value = nextMap
  saveCustomWallpapersToStorage(nextMap)

  if (threadWallpaper.value === target.url) {
    setConversationWallpaper(conversationId, '')
  }
  ui.showToast(t.value.wallpaperRemoved)
}

// Storage & Media Picker for Chat Wallpaper
const storagePickerOpen = ref(false)
const storagePickerTab = ref<'chat' | 'personal'>('chat')
const personalPhotos = ref<TimelineItem[]>([])
const loadingPersonalPhotos = ref(false)
const personalPhotosError = ref('')
const personalPhotosLoaded = ref(false)

async function loadPersonalPhotos() {
  if (loadingPersonalPhotos.value) return
  loadingPersonalPhotos.value = true
  personalPhotosError.value = ''
  try {
    const data = await api<Timeline>('/photos/timeline')
    const allItems: TimelineItem[] = []
    for (const group of data.groups) {
      allItems.push(...group.items.filter((item) => item.mimeType.startsWith('image/')))
    }
    personalPhotos.value = allItems
    personalPhotosLoaded.value = true
  } catch (e) {
    personalPhotosError.value = formatApiError(e, t.value.personalPhotosLoadError)
  } finally {
    loadingPersonalPhotos.value = false
  }
}

function openMediaStoragePicker() {
  storagePickerOpen.value = true
  if (storagePickerTab.value === 'personal' && !personalPhotosLoaded.value) {
    void loadPersonalPhotos()
  }
}

function switchStoragePickerTab(tab: 'chat' | 'personal') {
  storagePickerTab.value = tab
  if (tab === 'personal' && !personalPhotosLoaded.value) {
    void loadPersonalPhotos()
  }
}

async function selectPhotoFromChatMedia(photo: ChatAttachment) {
  if (!selectedConversation.value) return
  storagePickerOpen.value = false
  const convId = selectedConversation.value.id
  let photoUrl = photo.thumbnailUrl
  if (!photoUrl || photoUrl.startsWith('/api')) {
    try {
      const out = await api<DownloadURL>(`/chat/conversations/${convId}/attachments/${photo.id}/download`)
      photoUrl = out.downloadUrl
    } catch {
      photoUrl = photo.thumbnailUrl || ''
    }
  }
  if (!photoUrl) {
    ui.showToast(t.value.wallpaperImageUrlError)
    return
  }
  const customItem: ChatCustomWallpaper = {
    id: generateUUID(),
    url: photoUrl,
    name: photo.name || photo.originalName || 'Ảnh đoạn chat',
    source: 'chat_media',
    sourceId: photo.id,
    createdAt: new Date().toISOString(),
  }
  addCustomWallpaperToConversation(convId, customItem)
  setConversationWallpaper(convId, photoUrl)
  ui.showToast(t.value.wallpaperUpdated)
}

async function selectPhotoFromPersonalStorage(item: TimelineItem) {
  if (!selectedConversation.value) return
  storagePickerOpen.value = false
  const convId = selectedConversation.value.id
  let photoUrl = item.thumbnailUrl
  try {
    const out = await api<DownloadURL>(`/files/${item.id}/download`)
    photoUrl = out.downloadUrl
  } catch {
    // fallback
  }
  if (!photoUrl) {
    ui.showToast(t.value.wallpaperImageUrlError)
    return
  }
  const customItem: ChatCustomWallpaper = {
    id: generateUUID(),
    url: photoUrl,
    name: item.name || 'Ảnh kho cá nhân',
    source: 'storage',
    sourceId: item.id,
    createdAt: new Date().toISOString(),
  }
  addCustomWallpaperToConversation(convId, customItem)
  setConversationWallpaper(convId, photoUrl)
  ui.showToast(t.value.wallpaperUpdated)
}

const threadWallpaper = computed(() => {
  if (!selectedConversation.value) return null
  return chatWallpapers.value[selectedConversation.value.id] || null
})

const threadWallpaperBackground = computed(() => {
  if (!threadWallpaper.value) return ''
  const preset = CHAT_WALLPAPER_PRESETS.find((p) => p.id === threadWallpaper.value)
  if (preset) return preset.value
  return `url("${threadWallpaper.value}")`
})

const currentWallpaperPresetName = computed(() => {
  if (!threadWallpaper.value || threadWallpaper.value === 'none') return 'Mặc định'
  const preset = CHAT_WALLPAPER_PRESETS.find((p) => p.id === threadWallpaper.value)
  if (preset) return preset.name
  return 'Ảnh tùy chỉnh'
})

const activePreviewStyle = computed(() => {
  if (!threadWallpaper.value) return { background: 'var(--canvas)' }
  const preset = CHAT_WALLPAPER_PRESETS.find((p) => p.id === threadWallpaper.value)
  if (preset) {
    return preset.value ? { backgroundImage: preset.value } : { background: 'var(--canvas)' }
  }
  return {
    backgroundImage: `url("${threadWallpaper.value}")`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  }
})

watch(
  threadWallpaper,
  async (wp) => {
    activeWallpaperTheme.value = await resolveWallpaperTheme(wp)
  },
  { immediate: true },
)

async function setConversationWallpaper(conversationId: string, wallpaperVal: string) {
  if (wallpaperVal === 'none' || !wallpaperVal) {
    const next = { ...chatWallpapers.value }
    delete next[conversationId]
    chatWallpapers.value = next
    saveWallpapersToStorage(next)
    activeWallpaperTheme.value = null
  } else {
    const next = { ...chatWallpapers.value, [conversationId]: wallpaperVal }
    chatWallpapers.value = next
    saveWallpapersToStorage(next)
    activeWallpaperTheme.value = await resolveWallpaperTheme(wallpaperVal)
  }
}

function selectWallpaperPreset(presetId: string) {
  if (!selectedConversation.value) return
  setConversationWallpaper(selectedConversation.value.id, presetId === 'none' ? '' : presetId)
  ui.showToast(presetId === 'none' ? t.value.wallpaperRemoved : t.value.wallpaperUpdated)
}

function removeConversationWallpaper() {
  if (!selectedConversation.value) return
  setConversationWallpaper(selectedConversation.value.id, '')
  ui.showToast(t.value.wallpaperRemoved)
}

const wallpaperCategoryFilter = ref<'all' | 'light' | 'dark'>('all')

const filteredWallpaperPresets = computed(() => {
  if (wallpaperCategoryFilter.value === 'light') {
    return CHAT_WALLPAPER_PRESETS.filter((p) => !p.isDark)
  }
  if (wallpaperCategoryFilter.value === 'dark') {
    return CHAT_WALLPAPER_PRESETS.filter((p) => p.isDark || p.id === 'none')
  }
  return CHAT_WALLPAPER_PRESETS
})

function openWallpaperSubPage() {
  chatInfoCurrentView.value = 'wallpaper'
  wallpaperCategoryFilter.value = resolvedIsDark.value ? 'dark' : 'light'
}

function triggerWallpaperFileInput() {
  wallpaperFileInputRef.value?.click()
}

function compressImageForWallpaper(file: File, maxDimension = 1280, quality = 0.82): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = reject
    reader.onload = () => {
      const img = new Image()
      img.onerror = reject
      img.onload = () => {
        let width = img.width
        let height = img.height
        if (width > maxDimension || height > maxDimension) {
          if (width > height) {
            height = Math.round((height * maxDimension) / width)
            width = maxDimension
          } else {
            width = Math.round((width * maxDimension) / height)
            height = maxDimension
          }
        }
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          resolve(reader.result as string)
          return
        }
        ctx.drawImage(img, 0, 0, width, height)
        resolve(canvas.toDataURL('image/jpeg', quality))
      }
      img.src = reader.result as string
    }
    reader.readAsDataURL(file)
  })
}

async function handleWallpaperUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file || !selectedConversation.value) return

  if (!file.type.startsWith('image/')) {
    ui.showToast(t.value.wallpaperInvalidImage)
    return
  }

  const convId = selectedConversation.value.id
  try {
    const dataUrl = await compressImageForWallpaper(file)
    const customItem: ChatCustomWallpaper = {
      id: generateUUID(),
      url: dataUrl,
      name: file.name || 'Ảnh tải lên',
      source: 'uploaded',
      createdAt: new Date().toISOString(),
    }
    addCustomWallpaperToConversation(convId, customItem)
    setConversationWallpaper(convId, dataUrl)
    ui.showToast(t.value.wallpaperUpdated)

    // Asynchronously backup original image to user's Filvault personal storage
    void (async () => {
      try {
        const contentType = resolveContentType(file) || file.type || 'image/jpeg'
        const session = await api<UploadSession>('/files/upload-sessions', {
          method: 'POST',
          body: JSON.stringify({
            name: file.name,
            size: file.size,
            contentType,
            folderId: null,
          }),
        })
        await uploadToPresigned(session.uploadUrl, file, contentType)
        await api(`/files/${session.fileId}/complete`, { method: 'POST', body: '{}' })
      } catch (uploadErr) {
        console.warn('Background backup of wallpaper to personal storage skipped/failed:', uploadErr)
      }
    })()
  } catch {
    ui.showToast(t.value.wallpaperProcessError)
  } finally {
    target.value = ''
  }
}

const stickerPickerOpen = ref(false)

function toggleStickerPicker() {
  stickerPickerOpen.value = !stickerPickerOpen.value
}

function handleSendSticker(sticker: Sticker) {
  draft.value = formatStickerMessage(sticker)
  stickerPickerOpen.value = false
  void sendText()
}

function handleInsertEmoji(emoji: string) {
  const textarea = composerRef.value
  if (!textarea) {
    draft.value += emoji
    return
  }
  const start = textarea.selectionStart ?? draft.value.length
  const end = textarea.selectionEnd ?? draft.value.length
  draft.value = draft.value.slice(0, start) + emoji + draft.value.slice(end)
  void nextTick(() => {
    const pos = start + emoji.length
    textarea.setSelectionRange(pos, pos)
    onComposerInput()
  })
}

const activeMediaTab = ref<'media' | 'file' | 'link'>('media')
const highlightedMessageId = ref<string | null>(null)

// In-thread-info search state
const infoSearchOpen = ref(false)
const infoSearchQuery = ref('')
const infoSearchResults = ref<ChatMessage[]>([])
const infoSearching = ref(false)
const infoSearchDone = ref(false)
const infoSearchInputRef = ref<HTMLInputElement | null>(null)

interface SharedLinkItem {
  url: string
  domain: string
  messageId: string
  createdAt: string
  senderName: string
}

const sharedPhotos = computed(() =>
  media.value.filter(
    (item) =>
      item.availability === 'available' &&
      (item.mimeType.startsWith('image/') || item.mimeType.startsWith('video/')),
  ),
)

const sharedFiles = computed(() =>
  media.value.filter(
    (item) =>
      item.availability === 'available' &&
      !item.mimeType.startsWith('image/') &&
      !item.mimeType.startsWith('video/'),
  ),
)

const URL_REGEX = /(https?:\/\/[^\s<>"{}|\\^`]+)/gi

function messageSenderName(msg: ChatMessage): string {
  if (msg.senderId === auth.user?.id) return 'Bạn'
  const c = conversations.value.find((item) => item.id === msg.conversationId)
  if (c?.peer?.name) return c.peer.name
  if (c?.peer?.email) return c.peer.email
  return 'Người gửi'
}

const sharedLinks = computed<SharedLinkItem[]>(() => {
  const links: SharedLinkItem[] = []
  const seen = new Set<string>()
  for (const m of messages.value) {
    if (!m.body || m.removedAt) continue
    const matches = m.body.match(URL_REGEX)
    if (matches) {
      for (const rawUrl of matches) {
        const cleanUrl = rawUrl.replace(/[.,;!?)]+$/, '')
        const key = `${m.id}-${cleanUrl}`
        if (seen.has(key)) continue
        seen.add(key)
        let domain = cleanUrl
        try {
          domain = new URL(cleanUrl).hostname
        } catch {
          // keep cleanUrl
        }
        links.push({
          url: cleanUrl,
          domain,
          messageId: m.id,
          createdAt: m.createdAt,
          senderName: messageSenderName(m),
        })
      }
    }
  }
  return links.reverse()
})

let infoSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null

function onInfoSearchInput() {
  if (infoSearchDebounceTimer) clearTimeout(infoSearchDebounceTimer)
  infoSearchDebounceTimer = setTimeout(() => {
    void performInfoSearch()
  }, 300)
}

async function performInfoSearch() {
  const q = infoSearchQuery.value.trim()
  if (!selectedId.value || q.length < 2) {
    infoSearchResults.value = []
    infoSearchDone.value = false
    return
  }
  infoSearching.value = true
  infoSearchDone.value = false
  try {
    const encoded = encodeURIComponent(q)
    const out = await api<{ messages: ChatMessage[] }>(
      `/chat/conversations/${selectedId.value}/messages/search?q=${encoded}`,
    )
    infoSearchResults.value = out.messages
    infoSearchDone.value = true
  } catch (e) {
    error.value = formatApiError(e, t.value.chatSearchFailed)
  } finally {
    infoSearching.value = false
  }
}

function clearInfoSearch() {
  infoSearchQuery.value = ''
  infoSearchResults.value = []
  infoSearchDone.value = false
}

type ChatInfoView = 'main' | 'search' | 'media' | 'wallpaper'
const chatInfoCurrentView = ref<ChatInfoView>('main')

const infoSectionsOpen = ref<Record<string, boolean>>({
  customization: true,
  media: true,
  privacy: true,
})

function toggleInfoSection(section: 'customization' | 'media' | 'privacy') {
  infoSectionsOpen.value[section] = !infoSectionsOpen.value[section]
}

function messageSenderAvatar(msg: ChatMessage): string | null {
  if (msg.senderId === auth.user?.id) return auth.user?.avatarUrl || null
  const c = conversations.value.find((item) => item.id === msg.conversationId)
  return c?.peer?.avatarUrl || null
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

function formatSearchSnippet(body: string, query: string): string {
  if (!query.trim()) return escapeHtml(body)
  const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedQuery})`, 'gi')
  const escapedBody = escapeHtml(body)
  return escapedBody.replace(regex, '<mark class="search-match">$1</mark>')
}

const chatInfoTitle = computed(() => {
  if (chatInfoCurrentView.value === 'search') return t.value.search
  if (chatInfoCurrentView.value === 'media') return t.value.sharedMedia
  if (chatInfoCurrentView.value === 'wallpaper') return t.value.chatWallpaper
  return t.value.chatInfo
})

function openInfoSearchView() {
  chatInfoCurrentView.value = 'search'
  infoSearchOpen.value = true
  void nextTick(() => {
    infoSearchInputRef.value?.focus()
  })
}

function openMediaSubPage(tab: 'media' | 'file' | 'link' = 'media') {
  activeMediaTab.value = tab
  chatInfoCurrentView.value = 'media'
}

function returnToMainInfo() {
  chatInfoCurrentView.value = 'main'
  infoSearchOpen.value = false
}

function closeChatInfo() {
  threadInfoOpen.value = false
  chatInfoCurrentView.value = 'main'
  infoSearchOpen.value = false
}

function openChatInfo() {
  if (isDesktop.value) {
    desktopInfoOpen.value = true
  } else {
    threadInfoOpen.value = true
  }
}

function toggleChatInfo() {
  if (isDesktop.value) {
    desktopInfoOpen.value = !desktopInfoOpen.value
  } else {
    threadInfoOpen.value = !threadInfoOpen.value
  }
}

function closeDesktopInfo() {
  if (chatInfoCurrentView.value !== 'main') {
    chatInfoCurrentView.value = 'main'
    infoSearchOpen.value = false
  } else {
    desktopInfoOpen.value = false
  }
}

function scrollToMessage(messageId: string) {
  const el = document.getElementById(`msg-${messageId}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    highlightedMessageId.value = messageId
    setTimeout(() => {
      if (highlightedMessageId.value === messageId) {
        highlightedMessageId.value = null
      }
    }, 2500)
  }
}

async function jumpToMessageFromInfo(msg: ChatMessage) {
  closeChatInfo()
  if (!messages.value.some((m) => m.id === msg.id)) {
    searchResults.value = [msg]
  }
  await nextTick()
  scrollToMessage(msg.id)
}

async function jumpToMessageById(messageId: string) {
  closeChatInfo()
  const target = messages.value.find((m) => m.id === messageId)
  if (!target) {
    return
  }
  await nextTick()
  scrollToMessage(messageId)
}

function handleClickOutsideStickerPicker(event: MouseEvent | PointerEvent) {
  if (!stickerPickerOpen.value) return
  const target = event.target as HTMLElement | null
  if (!target) return
  if (target.closest('.sticker-picker-drawer') || target.closest('.sticker-toggle-btn')) {
    return
  }
  stickerPickerOpen.value = false
}

const threadSearchOpen = ref(false)
const isMobile = ref(
  typeof window !== 'undefined'
    ? window.matchMedia('(max-width: 767px)').matches || window.matchMedia('(pointer: coarse)').matches
    : false,
)
const isDesktop = ref(false)
const desktopInfoOpen = ref(
  typeof localStorage !== 'undefined'
    ? localStorage.getItem('filvault.chat.desktopInfoOpen') !== 'false'
    : true,
)

watch(desktopInfoOpen, (val) => {
  try {
    localStorage.setItem('filvault.chat.desktopInfoOpen', String(val))
  } catch {
    // ignore
  }
})

// Desktop Resizable Panels Configuration
const DEFAULT_RAIL_WIDTH = 320
const MIN_RAIL_WIDTH = 240
const MAX_RAIL_WIDTH = 520

const DEFAULT_INFO_WIDTH = 340
const MIN_INFO_WIDTH = 260
const MAX_INFO_WIDTH = 540

const savedRailWidth = typeof localStorage !== 'undefined' ? Number(localStorage.getItem('filvault.chat.railWidth')) : NaN
const railWidth = ref(!isNaN(savedRailWidth) && savedRailWidth >= MIN_RAIL_WIDTH && savedRailWidth <= MAX_RAIL_WIDTH ? savedRailWidth : DEFAULT_RAIL_WIDTH)

const savedInfoWidth = typeof localStorage !== 'undefined' ? Number(localStorage.getItem('filvault.chat.infoWidth')) : NaN
const infoWidth = ref(!isNaN(savedInfoWidth) && savedInfoWidth >= MIN_INFO_WIDTH && savedInfoWidth <= MAX_INFO_WIDTH ? savedInfoWidth : DEFAULT_INFO_WIDTH)

const isResizingRail = ref(false)
const isResizingInfo = ref(false)

function startRailResize(e: PointerEvent) {
  if (e.button !== 0) return
  isResizingRail.value = true
  const startX = e.clientX
  const startW = railWidth.value

  const onPointerMove = (moveEv: PointerEvent) => {
    const delta = moveEv.clientX - startX
    const availableWidth = typeof window !== 'undefined' ? window.innerWidth : 1200
    const maxAllowedRail = Math.max(MIN_RAIL_WIDTH, Math.min(MAX_RAIL_WIDTH, availableWidth - (desktopInfoOpen.value ? infoWidth.value : 0) - 300))
    const newW = Math.max(MIN_RAIL_WIDTH, Math.min(maxAllowedRail, Math.round(startW + delta)))
    railWidth.value = newW
  }

  const onPointerUp = () => {
    isResizingRail.value = false
    try {
      localStorage.setItem('filvault.chat.railWidth', String(railWidth.value))
    } catch {
      // ignore
    }
    window.removeEventListener('pointermove', onPointerMove)
    window.removeEventListener('pointerup', onPointerUp)
  }

  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
}

function resetRailWidth() {
  railWidth.value = DEFAULT_RAIL_WIDTH
  try {
    localStorage.setItem('filvault.chat.railWidth', String(DEFAULT_RAIL_WIDTH))
  } catch {
    // ignore
  }
}

function startInfoResize(e: PointerEvent) {
  if (e.button !== 0) return
  isResizingInfo.value = true
  const startX = e.clientX
  const startW = infoWidth.value

  const onPointerMove = (moveEv: PointerEvent) => {
    const delta = startX - moveEv.clientX
    const availableWidth = typeof window !== 'undefined' ? window.innerWidth : 1200
    const maxAllowedInfo = Math.max(MIN_INFO_WIDTH, Math.min(MAX_INFO_WIDTH, availableWidth - railWidth.value - 300))
    const newW = Math.max(MIN_INFO_WIDTH, Math.min(maxAllowedInfo, Math.round(startW + delta)))
    infoWidth.value = newW
  }

  const onPointerUp = () => {
    isResizingInfo.value = false
    try {
      localStorage.setItem('filvault.chat.infoWidth', String(infoWidth.value))
    } catch {
      // ignore
    }
    window.removeEventListener('pointermove', onPointerMove)
    window.removeEventListener('pointerup', onPointerUp)
  }

  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
}

function resetInfoWidth() {
  infoWidth.value = DEFAULT_INFO_WIDTH
  try {
    localStorage.setItem('filvault.chat.infoWidth', String(DEFAULT_INFO_WIDTH))
  } catch {
    // ignore
  }
}

const RESIZE_STEP = 16

function clampRailWidth(width: number): number {
  const availableWidth = typeof window !== 'undefined' ? window.innerWidth : 1200
  const maxAllowedRail = Math.max(
    MIN_RAIL_WIDTH,
    Math.min(MAX_RAIL_WIDTH, availableWidth - (desktopInfoOpen.value ? infoWidth.value : 0) - 300),
  )
  return Math.max(MIN_RAIL_WIDTH, Math.min(maxAllowedRail, Math.round(width)))
}

function clampInfoWidth(width: number): number {
  const availableWidth = typeof window !== 'undefined' ? window.innerWidth : 1200
  const maxAllowedInfo = Math.max(
    MIN_INFO_WIDTH,
    Math.min(MAX_INFO_WIDTH, availableWidth - railWidth.value - 300),
  )
  return Math.max(MIN_INFO_WIDTH, Math.min(maxAllowedInfo, Math.round(width)))
}

function persistRailWidth() {
  try {
    localStorage.setItem('filvault.chat.railWidth', String(railWidth.value))
  } catch {
    // ignore
  }
}

function persistInfoWidth() {
  try {
    localStorage.setItem('filvault.chat.infoWidth', String(infoWidth.value))
  } catch {
    // ignore
  }
}

function onRailResizerKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowRight') {
    e.preventDefault()
    railWidth.value = clampRailWidth(railWidth.value + RESIZE_STEP)
    persistRailWidth()
  } else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    railWidth.value = clampRailWidth(railWidth.value - RESIZE_STEP)
    persistRailWidth()
  } else if (e.key === 'Home') {
    e.preventDefault()
    resetRailWidth()
  } else if (e.key === 'End') {
    e.preventDefault()
    railWidth.value = clampRailWidth(MAX_RAIL_WIDTH)
    persistRailWidth()
  }
}

function onInfoResizerKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowLeft') {
    e.preventDefault()
    infoWidth.value = clampInfoWidth(infoWidth.value + RESIZE_STEP)
    persistInfoWidth()
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    infoWidth.value = clampInfoWidth(infoWidth.value - RESIZE_STEP)
    persistInfoWidth()
  } else if (e.key === 'Home') {
    e.preventDefault()
    resetInfoWidth()
  } else if (e.key === 'End') {
    e.preventDefault()
    infoWidth.value = clampInfoWidth(MAX_INFO_WIDTH)
    persistInfoWidth()
  }
}

type AttachmentUploadStatus = 'queued' | 'uploading' | 'failed' | 'canceled'
type AttachmentUpload = {
  id: string
  file: File
  conversationId: string
  body: string
  status: AttachmentUploadStatus
  error?: string
  controller: AbortController
  previewUrl?: string
  progress?: number
  isImage?: boolean
}
const attachmentQueue = ref<AttachmentUpload[]>([])
const activeAttachmentQueue = computed(() => {
  if (!selectedId.value) return []
  return attachmentQueue.value.filter(
    (entry) => entry.conversationId === selectedId.value && entry.status !== 'canceled',
  )
})
const MEDIA_TAB_IDS = ['media', 'file', 'link'] as const
type MediaTabId = (typeof MEDIA_TAB_IDS)[number]

function mapAttachmentStatusToProgress(status: AttachmentUploadStatus): UploadItemStatus {
  if (status === 'canceled') return 'cancelled'
  if (status === 'failed') return 'failed'
  if (status === 'uploading') return 'uploading'
  return 'queued'
}

const attachmentUploadProgressItems = computed(() =>
  attachmentQueue.value
    .filter((entry) => entry.status !== 'canceled')
    .map((entry) => ({
      id: entry.id,
      name: entry.file.name,
      status: mapAttachmentStatusToProgress(entry.status),
      progress: entry.progress ?? 0,
      error: entry.error,
    })),
)

const attachmentAggregateProgress = computed(() => {
  const active = attachmentQueue.value.filter((entry) => entry.status !== 'canceled')
  if (active.length === 0) return null
  const uploading = active.find((entry) => entry.status === 'uploading')
  if (uploading) return uploading.progress ?? 0
  if (active.some((entry) => entry.status === 'queued')) return 0
  return null
})

const hasActiveAttachmentUpload = computed(() =>
  attachmentQueue.value.some((entry) => entry.status === 'uploading'),
)

function processNextInAttachmentQueue() {
  if (hasActiveAttachmentUpload.value) return
  const next = attachmentQueue.value.find((entry) => entry.status === 'queued')
  if (next) void processAttachment(next.id)
}

function selectMediaTab(tabId: MediaTabId) {
  activeMediaTab.value = tabId
}

function focusMediaTab(tabId: MediaTabId, groupPrefix = 'shared-media') {
  void nextTick(() => {
    document.getElementById(`${groupPrefix}-tab-${tabId}`)?.focus()
  })
}

function onMediaTabKeydown(e: KeyboardEvent, tabId: MediaTabId, groupPrefix = 'shared-media') {
  const idx = MEDIA_TAB_IDS.indexOf(tabId)
  if (idx < 0) return
  let nextIdx = idx
  if (e.key === 'ArrowRight') nextIdx = (idx + 1) % MEDIA_TAB_IDS.length
  else if (e.key === 'ArrowLeft') nextIdx = (idx - 1 + MEDIA_TAB_IDS.length) % MEDIA_TAB_IDS.length
  else if (e.key === 'Home') nextIdx = 0
  else if (e.key === 'End') nextIdx = MEDIA_TAB_IDS.length - 1
  else return
  e.preventDefault()
  const nextTab = MEDIA_TAB_IDS[nextIdx]!
  selectMediaTab(nextTab)
  focusMediaTab(nextTab, groupPrefix)
}
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
  const nick = nicknames.value[conversation.id]
  if (nick) return nick
  return conversation.peer?.name || conversation.peer?.email || conversation.title || t.value.untitledChat
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
  if (route.params.id) {
    void router.push('/chat')
  }
}

const QUICK_REACTIONS = ['❤️', '😆', '😮', '😢', '😡', '👍'] as const
const activeBurstMessageId = ref<string | null>(null)
const capsulePos = ref<{ top: number; left: number; above: boolean } | null>(null)
const activeBubbleRect = ref<{ top: number; left: number; width: number; height: number } | null>(null)

let lastTapTime = 0
let lastTapMsgId = ''

function handleMessageBubbleClick(message: ChatMessage) {
  const now = Date.now()
  if (lastTapMsgId === message.id && now - lastTapTime < 320) {
    lastTapTime = 0
    lastTapMsgId = ''
    triggerHeartBurst(message)
    return
  }
  lastTapTime = now
  lastTapMsgId = message.id
}

function triggerHeartBurst(message: ChatMessage) {
  activeBurstMessageId.value = message.id
  setTimeout(() => {
    if (activeBurstMessageId.value === message.id) {
      activeBurstMessageId.value = null
    }
  }, 900)
  void selectReaction('❤️', message)
}

async function selectReaction(emoji: string, targetMessage?: ChatMessage) {
  const msg = targetMessage ?? activeMessage.value
  if (!msg || !selectedId.value) return
  messageMenuOpen.value = false
  try {
    const updated = await chatStore.toggleReaction(selectedId.value, msg.id, emoji)
    msg.reactions = updated
    const local = messages.value.find((m) => m.id === msg.id)
    if (local) {
      local.reactions = updated
    }
  } catch (err) {
    ui.showToast(formatApiError(err, t.value.reactionFailed), 'error')
  }
}

const reactionPickerOpen = ref(false)
const reactionPickerTargetMessage = ref<ChatMessage | null>(null)

const targetMessageReactionList = computed(() => {
  if (!reactionPickerTargetMessage.value?.reactions) return []
  return reactionPickerTargetMessage.value.reactions
    .filter((r) => r.reacted)
    .map((r) => r.reaction)
})

function openReactionPicker() {
  const msg = activeMessage.value
  messageMenuOpen.value = false
  if (!msg) return
  reactionPickerTargetMessage.value = msg
  reactionPickerOpen.value = true
}

function handlePickReactionEmoji(emoji: string) {
  const msg = reactionPickerTargetMessage.value
  reactionPickerOpen.value = false
  if (!msg || !selectedId.value) return
  void selectReaction(emoji, msg)
}

function triggerReplyMessage() {
  const msg = activeMessage.value
  messageMenuOpen.value = false
  if (!msg) return
  const preview = msg.body ? (msg.body.length > 50 ? msg.body.slice(0, 50) + '…' : msg.body) : 'Tệp đính kèm'
  draft.value = `> ${preview}\n`
  focusComposer(true)
}

function isReactedWith(message: ChatMessage | null, emoji: string): boolean {
  if (!message?.reactions) return false
  const r = message.reactions.find((item) => item.reaction === emoji)
  return Boolean(r?.reacted)
}

function hasUserReacted(message: ChatMessage): boolean {
  return Boolean(message.reactions?.some((r) => r.reacted))
}

function totalReactionCount(message: ChatMessage): number {
  if (!message.reactions) return 0
  return message.reactions.reduce((sum, r) => sum + r.count, 0)
}

function reactionTooltip(message: ChatMessage): string {
  if (!message.reactions) return ''
  return message.reactions.map((r) => `${r.reaction} (${r.count})`).join(', ')
}

const {
  start: onMessageTouchStartRaw,
  move: onMessageTouchMove,
  end: onMessageTouchEnd,
  cancel: onMessageTouchCancel,
} = useLongPress({
  onLongPress: (payload) => {
    if (payload) {
      const data = payload as { message: ChatMessage; el?: HTMLElement }
      if (data.message) {
        openMessageMenu(data.message, data.el)
      } else {
        openMessageMenu(payload as ChatMessage)
      }
    }
  },
})

function onMessageTouch(e: TouchEvent, message: ChatMessage) {
  onMessageTouchStartRaw(e, { message, el: e.currentTarget as HTMLElement })
}

function openMessageMenu(message: ChatMessage, el?: HTMLElement) {
  activeMessage.value = message
  if (el) {
    const rect = el.getBoundingClientRect()
    activeBubbleRect.value = {
      top: rect.top,
      left: rect.left,
      width: rect.width,
      height: rect.height,
    }
    const above = rect.top > 120
    capsulePos.value = {
      top: above ? Math.max(12, rect.top - 62) : rect.bottom + 10,
      left: Math.min(Math.max(12, rect.left + rect.width / 2 - 160), window.innerWidth - 332),
      above,
    }
  } else {
    activeBubbleRect.value = null
    capsulePos.value = {
      top: Math.max(60, window.innerHeight / 2 - 100),
      left: Math.max(16, (window.innerWidth - 320) / 2),
      above: true,
    }
  }
  messageMenuOpen.value = true
}

async function copyMessageText() {
  if (!activeMessage.value?.body) return
  try {
    await navigator.clipboard.writeText(activeMessage.value.body)
    ui.showToast(t.value.copiedMessage)
  } catch {
    ui.showToast(t.value.copyFailed, 'error')
  } finally {
    messageMenuOpen.value = false
  }
}

function triggerEditMessage() {
  const msg = activeMessage.value
  messageMenuOpen.value = false
  if (msg) void editMessage(msg)
}

function triggerRemoveMessage() {
  const msg = activeMessage.value
  messageMenuOpen.value = false
  if (msg) void removeMessage(msg)
}

const {
  start: onConvTouchStart,
  move: onConvTouchMove,
  end: onConvTouchEnd,
  cancel: onConvTouchCancel,
  shouldIgnoreClick: ignoreConvClick,
} = useLongPress({
  onLongPress: (payload) => {
    if (payload) openConvMenu(payload as ChatConversation)
  },
})

function openConvMenu(conv: ChatConversation) {
  activeConv.value = conv
  convMenuOpen.value = true
}

function handleConvRowClick(convId: string) {
  if (ignoreConvClick.value) return
  if (selectedId.value === convId && route.params.id === convId) return
  void selectConversation(convId)
  void router.push(`/chat/${convId}`)
}

function toggleMuteConversation(convId?: string) {
  const id = convId ?? selectedId.value
  if (!id) return
  mutedConversations.value[id] = !mutedConversations.value[id]
  ui.showToast(mutedConversations.value[id] ? t.value.muteChat : t.value.unmuteChat)
  convMenuOpen.value = false
}

async function deleteConversation(convId?: string) {
  const id = convId ?? selectedId.value
  if (!id) return
  const confirmed = await ui.confirm({
    title: t.value.deleteChat,
    message: t.value.deleteChatMessage,
    confirmLabel: t.value.remove,
    danger: true,
  })
  if (!confirmed) return
  convMenuOpen.value = false
  threadInfoOpen.value = false
  conversations.value = conversations.value.filter((c) => c.id !== id)
  if (selectedId.value === id) {
    selectedId.value = null
    messages.value = []
    if (route.params.id === id) {
      void router.replace('/chat')
    }
  }
  if (chatWallpapers.value[id]) {
    const next = { ...chatWallpapers.value }
    delete next[id]
    chatWallpapers.value = next
    saveWallpapersToStorage(next)
  }
  ui.showToast(t.value.deleteChatSuccess)
}

async function changeNickname() {
  if (!selectedConversation.value) return
  threadInfoOpen.value = false
  const currentNick = nicknames.value[selectedConversation.value.id] || conversationTitle(selectedConversation.value)
  const nick = await ui.prompt({
    title: t.value.changeNickname,
    label: t.value.newNicknameLabel,
    initialValue: currentNick,
    confirmLabel: t.value.nicknameSave,
  })
  if (nick !== null) {
    nicknames.value[selectedConversation.value.id] = nick.trim()
    ui.showToast(t.value.nicknameUpdated)
  }
}

function sendQuickWave() {
  draft.value = '👋'
  void sendText()
}

const peekOpen = ref(false)
const peekConv = ref<ChatConversation | null>(null)
const peekMessages = ref<ChatMessage[]>([])
const peekLoading = ref(false)

async function openPeekPreview(conv: ChatConversation) {
  convMenuOpen.value = false
  peekConv.value = conv
  peekLoading.value = true
  peekMessages.value = []
  peekOpen.value = true
  try {
    const out = await api<{ messages: ChatMessage[] }>(`/chat/conversations/${conv.id}/messages?limit=6`)
    peekMessages.value = out.messages
  } catch {
    ui.showToast(t.value.peekLoadError, 'error')
  } finally {
    peekLoading.value = false
  }
}

function openChatFromPeek() {
  if (!peekConv.value) return
  const id = peekConv.value.id
  peekOpen.value = false
  void selectConversation(id)
  void router.push(`/chat/${id}`)
}

function isPeerOnline(conv?: ChatConversation | null): boolean {
  if (!conv) return false
  if (auth.user?.activeStatusEnabled === false) return false
  return conv.peerStatus === 'online'
}

function formatLastSeen(conv?: ChatConversation | null): string {
  if (!conv) return ''
  if (auth.user?.activeStatusEnabled === false) return ''
  const iso = conv.peerLastSeenAt || conv.peer?.lastSeenAt
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (isNaN(then)) return ''
  const now = Date.now()
  const diffMs = Math.max(0, now - then)
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDays = Math.floor(diffHour / 24)

  if (diffMin < 1) {
    return t.value.activeJustNow
  }
  if (diffMin < 60) {
    return t.value.activeMinutesAgo.replace('{m}', String(diffMin))
  }
  if (diffHour < 24) {
    return t.value.activeHoursAgo.replace('{h}', String(diffHour))
  }
  if (diffDays === 1 || diffHour < 48) {
    return t.value.activeYesterday
  }
  if (diffDays <= 7) {
    return t.value.activeDaysAgo.replace('{d}', String(diffDays))
  }
  const thenDate = new Date(iso)
  const nowDate = new Date()
  const day = String(thenDate.getDate()).padStart(2, '0')
  const month = String(thenDate.getMonth() + 1).padStart(2, '0')
  if (thenDate.getFullYear() === nowDate.getFullYear()) {
    return t.value.activeOnDate.replace('{date}', `${day}/${month}`)
  }
  return t.value.activeOnDate.replace('{date}', `${day}/${month}/${thenDate.getFullYear()}`)
}

function isConversationUnread(conv: ChatConversation): boolean {
  if (conv.id === selectedId.value) return false
  return (conv.unreadCount ?? 0) > 0
}

function isMessageSeenByPeer(message: ChatMessage, index: number): boolean {
  if (message.senderId !== auth.user?.id || !selectedConversation.value) return false
  const list = visibleMessages.value
  let lastOutgoingIndex = -1
  for (let i = list.length - 1; i >= 0; i--) {
    if (list[i]?.senderId === auth.user?.id) {
      lastOutgoingIndex = i
      break
    }
  }
  if (index !== lastOutgoingIndex) return false
  const peerReadMsgId = selectedConversation.value.peerLastReadMessageId
  const peerReadAt = selectedConversation.value.peerLastReadAt
  if (peerReadMsgId && peerReadMsgId === message.id) return true
  if (peerReadAt && new Date(message.createdAt).getTime() <= new Date(peerReadAt).getTime()) return true
  return false
}

let typingTimer: ReturnType<typeof setTimeout> | null = null

function onComposerInput() {
  autoGrow()
  if (selectedId.value) {
    void chatStore.sendTyping(selectedId.value, true)
    if (typingTimer) clearTimeout(typingTimer)
    typingTimer = setTimeout(() => {
      if (selectedId.value) void chatStore.sendTyping(selectedId.value, false)
    }, 2500)
  }
}

function isPeerTyping(conversationId?: string): boolean {
  if (!conversationId) return false
  const typing = chatStore.typingUsers[conversationId]
  if (!typing) return false
  return Boolean(typing.userId && typing.userId !== auth.user?.id)
}


function updateViewport() {
  isMobile.value =
    window.matchMedia('(max-width: 767px)').matches ||
    window.matchMedia('(pointer: coarse)').matches
  isDesktop.value = window.matchMedia('(min-width: 1024px)').matches
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


function toggleLanguage() {
  const next = locale.value === 'vi' ? 'en' : 'vi'
  setLocale(next)
}

async function toggleActiveStatus() {
  const current = auth.user?.activeStatusEnabled !== false
  const next = !current
  try {
    await auth.updateActiveStatus(next)
    ui.showToast(next ? t.value.activeStatusEnabled : t.value.activeStatusDisabled)
    await loadConversations()
  } catch (e) {
    ui.showToast(formatApiError(e, t.value.activeStatusUpdateError), 'error')
  }
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
    const routeId = (route.params.id as string) || null
    if (routeId) {
      if (selectedId.value !== routeId) {
        await selectConversation(routeId)
      }
    } else if (!selectedId.value && out.conversations[0] && !isMobileViewport()) {
      await selectConversation(out.conversations[0].id)
      void router.replace(`/chat/${out.conversations[0].id}`)
    }
  } catch (e) {
    error.value = formatApiError(e, t.value.chatLoadFailed)
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
    void router.push(`/chat/${conversation.id}`)
  } catch (e) {
    error.value = formatApiError(e, t.value.chatOpenDirectFailed)
  }
}

async function selectConversation(id: string) {
  stickerPickerOpen.value = false
  const selection = activeSelection + 1
  activeSelection = selection
  selectedId.value = id
  void chatStore.openConversation(id)
  searchQuery.value = ''
  searchResults.value = null
  threadSearchOpen.value = false
  infoSearchOpen.value = false
  infoSearchQuery.value = ''
  infoSearchResults.value = []
  infoSearchDone.value = false
  chatInfoCurrentView.value = 'main'
  loadingThread.value = true
  messages.value = []
  hasMore.value = false
  nextBefore.value = null
  await Promise.all([loadMessages(id), loadMedia(id)])
  if (selection !== activeSelection || selectedId.value !== id) return
  await nextTick()
  scrollToLatest({ smooth: false })
  requestAnimationFrame(() => {
    scrollToLatest({ smooth: false })
  })
  focusComposer()
  const latestMsg = messages.value[messages.value.length - 1]
  void chatStore.markAsRead(id, latestMsg?.id)
}

async function loadMessages(id = selectedId.value) {
  if (!id) {
    loadingThread.value = false
    return
  }
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
    if (out.messages.length > 0) {
      void chatStore.markAsRead(id, out.messages[out.messages.length - 1]?.id)
    }
  } catch (e) {
    error.value = formatApiError(e, t.value.chatMessagesLoadFailed)
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
    error.value = formatApiError(e, t.value.chatOlderMessagesLoadFailed)
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

function scrollToLatest(options: { smooth?: boolean } = {}) {
  const el = threadBodyRef.value
  if (!el) return
  if (options.smooth && typeof el.scrollTo === 'function') {
    el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
  } else {
    el.scrollTop = el.scrollHeight
  }
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
  if (date.toDateString() === today.toDateString()) return t.value.today
  if (date.toDateString() === yesterday.toDateString()) return t.value.yesterday
  return date.toLocaleDateString(locale.value === 'vi' ? 'vi-VN' : 'en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function conversationPreview(conv: ChatConversation): string {
  if (conv.preview?.body) {
    if (isStickerMessage(conv.preview.body)) {
      const sym = parseStickerSymbol(conv.preview.body)
      return sym ? `[Sticker ${sym}]` : '[Sticker]'
    }
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
    error.value = formatApiError(e, t.value.chatSearchFailed)
  }
}

function clearSearch() {
  searchResults.value = null
  searchQuery.value = ''
}

const heicThumbs = ref(new Map<string, string>())

function getMediaThumbSrc(item: ChatAttachment): string {
  if (isHeic(item.name, item.mimeType)) {
    const cached = heicThumbs.value.get(item.id)
    if (cached) return cached
  }
  return item.thumbnailUrl || ''
}

async function resolveMediaThumbs(items: ChatAttachment[]) {
  for (const item of items) {
    if (isHeic(item.name, item.mimeType) && item.thumbnailUrl && !heicThumbs.value.has(item.id)) {
      getHeicDisplayUrl(item.thumbnailUrl, 0.4)
        .then((url) => {
          heicThumbs.value.set(item.id, url)
        })
        .catch(() => {})
    }
  }
}

async function loadMedia(id = selectedId.value) {
  if (!id) return
  const selection = activeSelection
  try {
    const out = await api<{ media: ChatAttachment[] }>(`/chat/conversations/${id}/media?type=all`)
    if (selection !== activeSelection || selectedId.value !== id) return
    media.value = out.media
    void resolveMediaThumbs(out.media)
  } catch (e) {
    error.value = formatApiError(e, t.value.chatMediaLoadFailed)
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
    title: t.value.chatEditMessageTitle,
    label: t.value.messageLabel,
    initialValue: message.body,
    confirmLabel: t.value.save,
  })
  if (!body?.trim() || body.trim() === message.body || !selectedId.value) return
  try {
    const updated = await api<ChatMessage>(
      `/chat/conversations/${selectedId.value}/messages/${message.id}`,
      { method: 'PATCH', body: JSON.stringify({ body: body.trim() }) },
    )
    messages.value = messages.value.map((item) => item.id === updated.id ? updated : item)
  } catch (e) {
    error.value = formatApiError(e, t.value.chatMessageEditFailed)
  }
}

async function removeMessage(message: ChatMessage) {
  if (!selectedId.value) return
  const confirmed = await ui.confirm({
    title: t.value.chatRemoveMessageTitle,
    message: t.value.chatRemoveMessageDescription,
    confirmLabel: t.value.remove,
    danger: true,
  })
  if (!confirmed) return
  try {
    await api(`/chat/conversations/${selectedId.value}/messages/${message.id}`, { method: 'DELETE' })
    messages.value = messages.value.map((item) =>
      item.id === message.id ? { ...item, body: '', removedAt: new Date().toISOString() } : item,
    )
  } catch (e) {
    error.value = formatApiError(e, t.value.chatMessageRemoveFailed)
  }
}

function attachmentLabel(attachment: ChatAttachment): string {
  if (attachment.availability === 'trashed') return t.value.fileMovedToTrash
  if (attachment.availability === 'purged') return t.value.filePermanentlyDeleted
  return ''
}

async function sendText() {
  if (sendingText.value) return
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
  sendingText.value = true
  if (typingTimer) clearTimeout(typingTimer)
  void chatStore.sendTyping(conversationId, false)
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
    pendingMessageError.value = formatApiError(e, t.value.chatMessageSendFailed)
    error.value = formatApiError(e, t.value.chatMessageSendFailed)
  } finally {
    sendingText.value = false
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

  const isImg = file.type.startsWith('image/') || isHeic(file.name, file.type)
  const convId = selectedId.value
  const itemId = generateUUID()
  let previewUrl = ''

  if (isImg) {
    if (isHeic(file.name, file.type)) {
      void (async () => {
        try {
          const converted = await convertHeicBlobToJpeg(file, 0.75)
          const objUrl = URL.createObjectURL(converted)
          const target = attachmentQueue.value.find((e) => e.id === itemId)
          if (target) {
            target.previewUrl = objUrl
          }
        } catch {
          // ignore HEIC preview error
        }
      })()
    } else {
      try {
        previewUrl = URL.createObjectURL(file)
      } catch {
        // ignore
      }
    }
  }

  const item: AttachmentUpload = {
    id: itemId,
    file,
    conversationId: convId,
    body: draft.value.trim(),
    status: 'queued',
    controller: new AbortController(),
    previewUrl,
    progress: 0,
    isImage: isImg,
  }
  attachmentQueue.value = [...attachmentQueue.value, item]
  draft.value = ''
  autoGrow()
  scrollToLatest({ smooth: true })
  processNextInAttachmentQueue()
}

async function processAttachment(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item || item.status !== 'queued') return
  if (hasActiveAttachmentUpload.value) return
  item.status = 'uploading'
  item.error = undefined
  item.progress = 0
  try {
    const contentType = resolveContentType(item.file) || item.file.type || 'application/octet-stream'
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
      item.progress = ratio
    }, item.controller.signal)
    const message = await api<ChatMessage>(`/chat/attachments/${session.fileId}/complete`, {
      method: 'POST',
      body: JSON.stringify({ conversationId: item.conversationId, body: item.body }),
    })
    if (selectedId.value === item.conversationId) messages.value = [...messages.value, message]
    searchResults.value = null
    await loadMedia(item.conversationId)
    if (item.previewUrl && item.previewUrl.startsWith('blob:')) {
      URL.revokeObjectURL(item.previewUrl)
    }
    attachmentQueue.value = attachmentQueue.value.filter((entry) => entry.id !== id)
    scrollToLatest({ smooth: true })
    ui.showToast(t.value.chatAttachmentSent)
  } catch (e) {
    if (item.status !== 'uploading') return
    item.status = 'failed'
    item.error = formatApiError(e, t.value.chatAttachmentSendFailed)
  } finally {
    processNextInAttachmentQueue()
  }
}

function retryUpload(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item || item.status === 'uploading') return
  item.status = 'queued'
  item.error = undefined
  item.progress = 0
  item.controller = new AbortController()
  processNextInAttachmentQueue()
}

function cancelUpload(id: string) {
  const item = attachmentQueue.value.find((entry) => entry.id === id)
  if (!item) return
  const wasUploading = item.status === 'uploading'
  item.status = 'canceled'
  item.controller.abort()
  if (item.previewUrl && item.previewUrl.startsWith('blob:')) {
    URL.revokeObjectURL(item.previewUrl)
  }
  attachmentQueue.value = attachmentQueue.value.filter((entry) => entry.id !== id)
  if (wasUploading) processNextInAttachmentQueue()
}

async function openAttachment(attachmentId: string) {
  if (!selectedId.value) return
  try {
    const out = await api<DownloadURL>(`/chat/conversations/${selectedId.value}/attachments/${attachmentId}/download`)
    const targetUrl = normalizePresignedUrl(out.downloadUrl)
    window.open(targetUrl, '_blank', 'noopener')
  } catch (e) {
    error.value = formatApiError(e, t.value.chatDownloadFailed)
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
    lightboxUrl.value = normalizePresignedUrl(out.downloadUrl)
    lightboxOpen.value = true
  } catch (e) {
    error.value = formatApiError(e, t.value.chatViewFailed)
  }
}

async function downloadLightbox() {
  if (!lightboxFileId.value) return
  await openAttachment(lightboxFileId.value)
}

function focusComposer(force = false) {
  if (!force) {
    const isTouchOrMobile =
      isMobile.value ||
      (typeof window !== 'undefined' &&
        (window.matchMedia('(max-width: 767px)').matches ||
          window.matchMedia('(pointer: coarse)').matches))
    if (isTouchOrMobile) return
  }
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

function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (reactionPickerOpen.value) {
      reactionPickerOpen.value = false
    } else if (messageMenuOpen.value) {
      messageMenuOpen.value = false
    } else if (stickerPickerOpen.value) {
      stickerPickerOpen.value = false
    } else if (infoSearchOpen.value) {
      infoSearchOpen.value = false
    } else if (threadSearchOpen.value) {
      threadSearchOpen.value = false
    }
  }
}

function onVisibilityChange(): void {
  if (typeof document !== 'undefined' && document.visibilityState === 'visible') {
    chatStore.handleWakeup()
  }
}

function onWindowFocus(): void {
  chatStore.handleWakeup()
}

onMounted(() => {
  updateViewport()
  window.addEventListener('resize', updateViewport)
  window.addEventListener('keydown', handleGlobalKeydown)
  window.addEventListener('offline', chatStore.markOffline)
  window.addEventListener('online', chatStore.markOnline)
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', onVisibilityChange)
    document.addEventListener('pointerdown', handleClickOutsideStickerPicker)
  }
  window.addEventListener('focus', onWindowFocus)
  if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', onVisualViewportResize)
    window.visualViewport.addEventListener('scroll', onVisualViewportResize)
  }
  const routeId = (route.params.id as string) || null
  if (routeId) {
    void selectConversation(routeId)
  }
  void loadConversations()
  chatStore.connectEvents()
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewport)
  window.removeEventListener('keydown', handleGlobalKeydown)
  window.removeEventListener('offline', chatStore.markOffline)
  window.removeEventListener('online', chatStore.markOnline)
  if (typeof document !== 'undefined') {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    document.removeEventListener('pointerdown', handleClickOutsideStickerPicker)
  }
  window.removeEventListener('focus', onWindowFocus)
  if (window.visualViewport) {
    window.visualViewport.removeEventListener('resize', onVisualViewportResize)
    window.visualViewport.removeEventListener('scroll', onVisualViewportResize)
  }
})

function initiateCall(isVideo: boolean) {
  if (!selectedConversation.value) return
  if (selectedConversation.value.type !== 'direct') return
  if (callStore.state !== 'idle') {
    ui.showToast(t.value.callAlreadyActive, 'info')
    return
  }
  callStore.startCall(selectedConversation.value.id, {
    isVideo,
    peerName: conversationTitle(selectedConversation.value),
    peerAvatar: selectedConversation.value.peer?.avatarUrl,
  })
}

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
    if (!nextMessages || !selectedId.value) return
    if (messages.value.length === 0) {
      messages.value = [...nextMessages]
      return
    }
    const newOnes = nextMessages.filter((m) => !messages.value.some((cur) => cur.id === m.id))
    if (newOnes.length > 0) {
      messages.value = [...messages.value, ...newOnes]
    }
  },
)

watch(
  () => chatStore.lastReactionUpdate,
  (update) => {
    if (!update || update.conversationId !== selectedId.value) return
    const msg = messages.value.find((m) => m.id === update.messageId)
    if (msg) {
      msg.reactions = update.reactions
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

watch(
  () => route.params.id,
  async (newId) => {
    const id = (newId as string) || null
    if (id) {
      if (selectedId.value !== id) {
        await selectConversation(id)
      }
    } else {
      selectedId.value = null
      messages.value = []
      searchQuery.value = ''
      searchResults.value = null
      threadSearchOpen.value = false
      loadingThread.value = false
    }
  },
)
</script>

<template>
  <div class="chat-app" :class="{ 'in-thread': inThread, 'resizing-col': isResizingRail || isResizingInfo }" :style="{ '--rail-w': `${railWidth}px`, '--info-w': `${infoWidth}px` }">
    <aside class="chat-rail" :aria-label="t.chats">
      <!-- Left resizer splitter (Desktop only) -->
      <div
        v-if="isDesktop"
        class="chat-resizer right-edge"
        :class="{ active: isResizingRail }"
        role="separator"
        tabindex="0"
        aria-orientation="vertical"
        :aria-valuenow="railWidth"
        :aria-valuemin="MIN_RAIL_WIDTH"
        :aria-valuemax="MAX_RAIL_WIDTH"
        :aria-label="t.resizeConversationListAria"
        :title="t.resizeConversationListTitle"
        @pointerdown.stop.prevent="startRailResize"
        @dblclick="resetRailWidth"
        @keydown="onRailResizerKeydown"
      >
        <div class="resizer-handle-line" />
      </div>
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
        <button type="button" class="icon-btn" :aria-label="t.backToFilvault" @click="backToVault">
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
        <LoadingSkeletonChatRail v-if="loading" />
        <div v-else-if="!filteredConversations.length" class="rail-empty">
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
          :class="{
            active: conv.id === selectedId,
            unread: isConversationUnread(conv),
          }"
          :aria-current="conv.id === selectedId ? 'true' : undefined"
          @click="handleConvRowClick(conv.id)"
          @touchstart.passive="onConvTouchStart($event, conv)"
          @touchmove.passive="onConvTouchMove($event)"
          @touchend="onConvTouchEnd"
          @touchcancel="onConvTouchCancel"
          @contextmenu.prevent="openConvMenu(conv)"
        >
          <span class="avatar-wrap">
            <img v-if="conv.peer?.avatarUrl" :src="conv.peer.avatarUrl" class="avatar avatar-img" alt="" />
            <span v-else class="avatar" :class="avatarClass(conversationTitle(conv))" aria-hidden="true">{{ conversationTitle(conv).slice(0, 1).toUpperCase() }}</span>
            <span v-if="isPeerOnline(conv)" class="online-indicator" aria-hidden="true" />
          </span>
          <span class="conversation-meta">
            <span class="conversation-top">
              <span class="conversation-title">{{ conversationTitle(conv) }}</span>
              <span class="conversation-date" :class="{ 'unread-date': isConversationUnread(conv) }">{{ formatRelativeDay(conv.preview?.createdAt ?? conv.updatedAt) }}</span>
            </span>
            <span class="conversation-bottom">
              <span v-if="isPeerTyping(conv.id)" class="conversation-typing">
                <span class="typing-pulse-dot" />
                <span>{{ t.isTyping }}</span>
              </span>
              <span v-else class="conversation-preview">{{ conversationPreview(conv) }}</span>
              <span v-if="isConversationUnread(conv)" class="unread-badge">
                {{ (conv.unreadCount ?? 0) > 1 ? conv.unreadCount : '' }}
              </span>
            </span>
          </span>
        </button>
      </nav>
    </aside>

    <section
      class="message-thread"
      :class="{ 'has-wallpaper': Boolean(threadWallpaperBackground), 'is-dark': activeWallpaperTheme?.isDark }"
      aria-live="polite"
      :style="threadThemeStyle"
    >
      <template v-if="selectedConversation">
        <header class="thread-header">
          <!-- aria-label="Back" -->
          <button type="button" class="icon-btn back-btn" :aria-label="t.back" @click="backToRail">
            <Icon name="arrow-left" :size="20" />
          </button>
          <div
            class="thread-peer-info clickable"
            role="button"
            tabindex="0"
            :title="t.chatInfo"
            @click="openChatInfo"
            @keydown.enter="openChatInfo"
          >
            <img
              v-if="selectedConversation?.peer?.avatarUrl"
              :src="selectedConversation.peer.avatarUrl"
              class="thread-avatar avatar-img"
              alt=""
            />
            <span
              v-else
              class="thread-avatar"
              :class="avatarClass(conversationTitle(selectedConversation))"
              aria-hidden="true"
            >{{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}</span>
            <div class="thread-peer-meta">
              <h2>{{ conversationTitle(selectedConversation) }}</h2>
              <span
                v-if="isPeerOnline(selectedConversation) || formatLastSeen(selectedConversation)"
                class="thread-status"
                :class="{ online: isPeerOnline(selectedConversation) }"
              >
                {{ isPeerOnline(selectedConversation) ? t.activeNow : formatLastSeen(selectedConversation) }}
              </span>
            </div>
          </div>
          <div class="thread-actions">
            <template v-if="selectedConversation?.type === 'direct'">
              <button
                class="icon-btn"
                type="button"
                :aria-label="t.callVoice"
                :title="t.callVoice"
                @click="initiateCall(false)"
              >
                <Icon name="phone" :size="18" />
              </button>
              <button
                class="icon-btn"
                type="button"
                :aria-label="t.callVideo"
                :title="t.callVideo"
                @click="initiateCall(true)"
              >
                <Icon name="camera" :size="18" />
              </button>
            </template>
            <button
              class="icon-btn thread-info-btn"
              :class="{ active: isDesktop ? desktopInfoOpen : threadInfoOpen }"
              type="button"
              :aria-label="t.chatInfo"
              :title="t.chatInfo"
              @click="toggleChatInfo"
            >
              <Icon name="info" :size="18" />
            </button>
          </div>
        </header>

        <div class="thread-canvas-wrap">
          <!-- Fixed wallpaper backdrop: scoped strictly inside canvas viewport, never bleeds into header or composer -->
          <div
            v-if="threadWallpaperBackground"
            class="thread-wallpaper-backdrop"
            :style="{ backgroundImage: threadWallpaperBackground }"
            aria-hidden="true"
          >
            <div
              class="wallpaper-scrim"
              :class="{ 'is-dark': activeWallpaperTheme?.isDark }"
            />
          </div>

          <p
            v-if="chatStore.connectionState !== 'connected'"
            class="connection-status"
            role="status"
            :title="t.retry"
            @click="chatStore.connectEvents()"
          >
            {{ chatStore.connectionState === 'offline' ? t.offlineStatus : t.reconnecting }}
          </p>

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
          <p v-else-if="loadingOlder" class="loading-older">{{ t.loadingOlder }}</p>
          <div v-else-if="visibleMessages.length" class="message-list">
            <TransitionGroup name="msg">
              <template v-for="(message, index) in visibleMessages" :key="message.id">
                <div v-if="shouldShowDate(index)" class="day-separator">
                  <span>{{ formatDayLabel(message.createdAt) }}</span>
                </div>
                <div
                  :id="'msg-' + message.id"
                  class="message-row"
                  :class="{
                    outgoing: message.senderId === auth.user?.id,
                    'msg-highlighted': highlightedMessageId === message.id,
                    'cluster-start': isFirstInCluster(index),
                    'cluster-last': isLastInCluster(index),
                    'cluster-middle': !isFirstInCluster(index) && !isLastInCluster(index),
                    'cluster-single': isFirstInCluster(index) && isLastInCluster(index),
                    'has-bubble-decoration':
                      message.senderId === auth.user?.id &&
                      isFirstInCluster(index) &&
                      currentBubbleStyle.id !== 'default' &&
                      currentBubbleStyle.decorations.length > 0,
                  }"
                >
                  <img
                    v-if="message.senderId !== auth.user?.id && isLastInCluster(index) && selectedConversation?.peer?.avatarUrl"
                    :src="selectedConversation.peer.avatarUrl"
                    class="row-avatar avatar-img"
                    alt=""
                  />
                  <span
                    v-else-if="message.senderId !== auth.user?.id && isLastInCluster(index)"
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
                  <article
                    class="message-bubble"
                    :class="{
                      outgoing: message.senderId === auth.user?.id,
                      'has-like': message.body === '👍',
                      'has-sticker': isStickerMessage(message.body),
                      'has-reactions': message.reactions && message.reactions.length > 0,
                      'has-dec-left': message.senderId === auth.user?.id && currentBubbleStyle.id !== 'default' && isFirstInCluster(index) && currentBubbleStyle.decorations.some(d => d.position === 'left'),
                      'has-dec-right': message.senderId === auth.user?.id && currentBubbleStyle.id !== 'default' && isFirstInCluster(index) && currentBubbleStyle.decorations.some(d => d.position === 'right'),
                      'has-dec-top': message.senderId === auth.user?.id && currentBubbleStyle.id !== 'default' && isFirstInCluster(index) && currentBubbleStyle.decorations.some(d => d.position.startsWith('top')),
                      'has-dec-bottom': message.senderId === auth.user?.id && currentBubbleStyle.id !== 'default' && isFirstInCluster(index) && currentBubbleStyle.decorations.some(d => d.position.startsWith('bottom')),
                    }"
                    :style="outgoingBubbleStyle(message)"
                    @touchstart.passive="onMessageTouch($event, message)"
                    @touchmove.passive="onMessageTouchMove($event)"
                    @touchend="onMessageTouchEnd"
                    @touchcancel="onMessageTouchCancel"
                    @contextmenu.prevent="openMessageMenu(message, $event.currentTarget as HTMLElement)"
                    @click="handleMessageBubbleClick(message)"
                  >
                    <!-- Decorations for custom bubble style (only on first/single in cluster) -->
                    <template v-if="message.senderId === auth.user?.id && currentBubbleStyle.id !== 'default' && !isStickerMessage(message.body) && message.body !== '👍' && isFirstInCluster(index)">
                      <div
                        v-for="(dec, dIdx) in currentBubbleStyle.decorations"
                        :key="dIdx"
                        class="bubble-decoration"
                        :class="`dec-${dec.position}`"
                      >
                        <img v-if="dec.imgUrl" :src="dec.imgUrl" class="bubble-decoration-img" alt="" />
                        <div v-else-if="dec.svg" v-html="dec.svg" />
                      </div>
                    </template>
                    <!-- Heart burst pop animation on double-tap -->
                    <div v-if="activeBurstMessageId === message.id" class="heart-burst" aria-hidden="true">
                      ❤️
                    </div>
                    <p v-if="message.body && isStickerMessage(message.body)" class="sticker-bubble">{{ parseStickerSymbol(message.body) }}</p>
                    <p v-else-if="message.body" :class="{ 'like-bubble': message.body === '👍' }">{{ message.body }}</p>
                    <template v-for="attachment in message.attachments" :key="attachment.id">
                      <ChatInlineImage
                        v-if="attachment.availability === 'available' && attachment.thumbnailUrl && attachment.mimeType.startsWith('image/')"
                        class="inline-image"
                        :attachment="attachment"
                        @click="openInlineImage"
                      />
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

                    <!-- Reaction badge pill on message bubble -->
                    <button
                      v-if="message.reactions && message.reactions.length > 0"
                      type="button"
                      class="reaction-badge-group"
                      :class="{
                        'reacted-by-me': hasUserReacted(message),
                        'outgoing-badge': message.senderId === auth.user?.id
                      }"
                      :title="reactionTooltip(message)"
                      @click.stop="openMessageMenu(message, $event.currentTarget as HTMLElement)"
                    >
                      <span
                        v-for="r in message.reactions.slice(0, 3)"
                        :key="r.reaction"
                        class="rx-emoji"
                        :class="{ 'my-rx': r.reacted }"
                      >
                        {{ r.reaction }}
                      </span>
                      <span v-if="totalReactionCount(message) > 1" class="rx-count">
                        {{ totalReactionCount(message) }}
                      </span>
                    </button>
                  </article>
                </div>
                <div v-if="isMessageSeenByPeer(message, index)" class="seen-indicator">
                  <img
                    v-if="selectedConversation?.peer?.avatarUrl"
                    :src="selectedConversation.peer.avatarUrl"
                    class="seen-avatar avatar-img"
                    alt=""
                    :title="`${t.seenAt} ${formatTime(selectedConversation?.peerLastReadAt || message.createdAt)}`"
                  />
                  <span
                    v-else
                    class="seen-avatar"
                    :class="avatarClass(conversationTitle(selectedConversation))"
                    :title="`${t.seenAt} ${formatTime(selectedConversation?.peerLastReadAt || message.createdAt)}`"
                  >
                    {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
                  </span>
                  <span class="seen-text">{{ t.seen }}</span>
                </div>
              </template>
            </TransitionGroup>
            <!-- Typing indicator in thread -->
            <div
              v-if="selectedConversation && isPeerTyping(selectedConversation.id)"
              class="message-row typing-row"
            >
              <img
                v-if="selectedConversation?.peer?.avatarUrl"
                :src="selectedConversation.peer.avatarUrl"
                class="row-avatar avatar-img"
                alt=""
              />
              <span
                v-else
                class="row-avatar"
                :class="avatarClass(conversationTitle(selectedConversation))"
                aria-hidden="true"
              >
                {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
              </span>
              <div class="typing-bubble" :aria-label="`${conversationTitle(selectedConversation)} ${t.isTyping}`">
                <span class="typing-dot" />
                <span class="typing-dot" />
                <span class="typing-dot" />
              </div>
            </div>
            <Transition name="msg">
              <div v-if="pendingMessage" class="message-row outgoing cluster-single">
                <article
                  class="message-bubble pending outgoing"
                  :class="{
                    'has-like': pendingMessage.body === '👍',
                    'has-sticker': isStickerMessage(pendingMessage.body)
                  }"
                  :style="outgoingBubbleStyle({ senderId: auth.user?.id, body: pendingMessage.body } as any)"
                  aria-live="polite"
                >
                  <!-- Decorations for custom bubble style -->
                  <template v-if="currentBubbleStyle.id !== 'default' && !isStickerMessage(pendingMessage.body) && pendingMessage.body !== '👍'">
                    <div
                      v-for="(dec, dIdx) in currentBubbleStyle.decorations"
                      :key="dIdx"
                      class="bubble-decoration"
                      :class="`dec-${dec.position}`"
                    >
                      <img v-if="dec.imgUrl" :src="dec.imgUrl" class="bubble-decoration-img" alt="" />
                      <div v-else-if="dec.svg" v-html="dec.svg" />
                    </div>
                  </template>
                  <p v-if="isStickerMessage(pendingMessage.body)" class="sticker-bubble">{{ parseStickerSymbol(pendingMessage.body) }}</p>
                  <p v-else :class="{ 'like-bubble': pendingMessage.body === '👍' }">{{ pendingMessage.body }}</p>
                  <span v-if="pendingMessageError" class="message-status">{{ pendingMessageError }}</span>
                  <div v-if="pendingMessageError" class="message-actions">
                    <button type="button" class="message-action" @click="retryPendingMessage">{{ t.retry }}<!-- Retry --></button>
                    <button type="button" class="message-action" @click="discardPendingMessage">{{ t.closeSelection }}<!-- Discard --></button>
                  </div>
                </article>
              </div>
            </Transition>

            <!-- Messenger-style optimistic pending attachment cards -->
            <TransitionGroup name="msg">
              <div
                v-for="item in activeAttachmentQueue"
                :key="item.id"
                class="message-row outgoing cluster-single pending-attachment-row"
              >
                <article
                  class="message-bubble pending outgoing has-pending-attachment"
                  :style="outgoingBubbleStyle({ senderId: auth.user?.id, body: item.body } as any)"
                >
                  <p v-if="item.body" class="pending-attachment-body">{{ item.body }}</p>

                  <!-- Image attachment: Messenger-style placeholder card -->
                  <div v-if="item.isImage" class="pending-image-card">
                    <img
                      v-if="item.previewUrl"
                      :src="item.previewUrl"
                      class="pending-image-thumb"
                      alt="Đang gửi ảnh..."
                    />
                    <div v-else class="pending-image-placeholder">
                      <div class="pending-skeleton-pulse" />
                      <span class="pending-heic-pill">HEIC</span>
                    </div>

                    <!-- Messenger-style central progress ring & blur overlay -->
                    <div class="pending-image-overlay" :class="{ 'has-error': item.status === 'failed' }">
                      <div v-if="item.status === 'uploading' || item.status === 'queued'" class="messenger-progress-ring">
                        <svg class="ring-svg" viewBox="0 0 44 44">
                          <circle class="ring-track" cx="22" cy="22" r="18" fill="none" stroke-width="3.5" />
                          <circle
                            class="ring-fill"
                            cx="22"
                            cy="22"
                            r="18"
                            fill="none"
                            stroke-width="3.5"
                            :stroke-dasharray="113.1"
                            :stroke-dashoffset="113.1 * (1 - (item.progress || 0))"
                          />
                        </svg>
                        <span class="ring-pct">{{ Math.round((item.progress || 0) * 100) }}%</span>
                      </div>

                      <div v-else-if="item.status === 'failed'" class="pending-failed-content">
                        <span class="failed-badge-icon">⚠️</span>
                        <span class="failed-msg">{{ item.error || t.uploadImageFailed }}</span>
                        <div class="failed-actions">
                          <button type="button" class="bubble-action-btn retry" @click="retryUpload(item.id)">
                            {{ t.retry }}
                          </button>
                          <button type="button" class="bubble-action-btn cancel" @click="cancelUpload(item.id)">
                            {{ t.cancelUpload }}
                          </button>
                        </div>
                      </div>
                    </div>

                    <!-- Top right cancel button while uploading -->
                    <button
                      v-if="item.status === 'uploading' || item.status === 'queued'"
                      type="button"
                      class="pending-bubble-cancel"
                      :title="t.cancelUpload"
                      :aria-label="t.cancelUpload"
                      @click="cancelUpload(item.id)"
                    >
                      <Icon name="x" :size="13" />
                    </button>
                  </div>

                  <!-- Non-image generic file placeholder -->
                  <div v-else class="pending-file-card">
                    <div class="pending-file-icon">
                      <Icon name="file" :size="20" />
                    </div>
                    <div class="pending-file-meta">
                      <span class="pending-file-name">{{ item.file.name }}</span>
                      <span class="pending-file-size">{{ formatBytes(item.file.size) }}</span>
                    </div>
                    <div v-if="item.status === 'uploading' || item.status === 'queued'" class="pending-file-actions">
                      <span class="pending-file-pct">{{ Math.round((item.progress || 0) * 100) }}%</span>
                      <button type="button" class="pending-file-cancel-btn" :title="t.cancelUpload" :aria-label="t.cancelUpload" @click="cancelUpload(item.id)">
                        <Icon name="x" :size="14" />
                      </button>
                    </div>
                    <div v-else-if="item.status === 'failed'" class="pending-file-actions">
                      <button type="button" class="message-action" @click="retryUpload(item.id)">
                        {{ t.retry }}
                      </button>
                      <button type="button" class="message-action" @click="cancelUpload(item.id)">
                        {{ t.cancelUpload }}
                      </button>
                    </div>
                  </div>
                </article>
              </div>
            </TransitionGroup>
          </div>
          <div v-else class="thread-empty-state">
            <img
              v-if="selectedConversation?.peer?.avatarUrl"
              :src="selectedConversation.peer.avatarUrl"
              class="thread-empty-avatar avatar-img"
              alt=""
            />
            <span v-else class="thread-empty-avatar" :class="avatarClass(conversationTitle(selectedConversation))" aria-hidden="true">
              {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
            </span>
            <h3 class="thread-empty-title">{{ conversationTitle(selectedConversation) }}</h3>
            <p class="thread-empty-subtitle">{{ t.connectedOnFilvault }}</p>
            <button type="button" class="btn ink thread-empty-wave-btn" @click="sendQuickWave">
              <span>👋</span>
              <span>{{ t.sayHello }}</span>
            </button>
          </div>
          <Transition name="msg">
            <button v-if="showJump && !searchResults" type="button" class="jump-latest" @click="scrollToLatest({ smooth: true })">
              <Icon name="download" :size="16" />
              <span>{{ t.latest }}</span>
            </button>
          </Transition>
        </div>
      </div>

        <UploadProgress
          v-if="attachmentUploadProgressItems.length"
          class="sr-only"
          :aggregate-progress="attachmentAggregateProgress"
          :items="attachmentUploadProgressItems"
          :label="t.attachment"
          @retry="retryUpload"
          @cancel="cancelUpload"
        />
        <!-- Sticker Picker Drawer -->
        <div v-if="stickerPickerOpen" class="sticker-picker-drawer">
          <EmojiPicker
            :show-stickers="true"
            :show-close="true"
            @select-emoji="handleInsertEmoji"
            @select-sticker="handleSendSticker"
            @close="stickerPickerOpen = false"
          />
        </div>
        <form class="chat-composer" @submit.prevent="sendText">
          <button type="button" class="icon-btn attach-btn" :aria-label="t.attachFile" @click="triggerAttachment">
            <Icon name="plus" :size="18" />
          </button>
          <button
            type="button"
            class="icon-btn sticker-toggle-btn"
            :class="{ active: stickerPickerOpen }"
            :aria-label="t.chooseSticker"
            :title="t.stickers"
            @click="toggleStickerPicker"
          >
            <Icon name="sticker" :size="20" />
          </button>
          <label class="sr-only" for="chat-message">{{ t.messageLabel }}</label>
          <textarea
            id="chat-message"
            ref="composerRef"
            v-model="draft"
            rows="1"
            placeholder="Aa"
            :disabled="sendingText"
            @input="onComposerInput"
            @keydown.enter="onComposerKeydown"
            @focus="stickerPickerOpen = false"
          ></textarea>
          <button
            v-if="!draft.trim()"
            type="button"
            class="like-btn"
            :aria-label="t.sendLike"
            :disabled="sendingText"
            @click="sendQuickLike"
          >
            <Icon name="thumb-up" :size="20" />
          </button>
          <button
            v-show="Boolean(draft.trim())"
            class="send-btn"
            type="submit"
            :aria-label="t.send"
            :class="{ active: Boolean(draft.trim()) }"
            :disabled="!draft.trim() || sendingText"
            @click.prevent="sendText"
          >
            <Icon name="send" :size="18" />
          </button>
          <label class="sr-only" for="chat-attachment">{{ t.attachment }}</label>
          <input id="chat-attachment" ref="fileInputRef" type="file" class="sr-only" @change="onAttachmentChange" />
        </form>
      </template>
      <template v-else>
        <p
          v-if="chatStore.connectionState !== 'connected'"
          class="connection-status"
          role="status"
          :title="t.retry"
          @click="chatStore.connectEvents()"
        >
          {{ chatStore.connectionState === 'offline' ? t.offlineStatus : t.reconnecting }}
        </p>
        <EmptyState
          :title="t.selectChat"
          :description="t.selectChatDesc"
          icon="chat"
        />
      </template>
    </section>

    <aside class="media-panel" :aria-label="t.sharedMedia">
      <div class="media-panel-header">
        <h2>{{ t.sharedMedia }}</h2>
      </div>
      <div class="media-panel-tabs" role="tablist" :aria-label="t.sharedMedia">
        <button
          id="media-panel-tab-media"
          type="button"
          class="panel-tab-btn"
          :class="{ active: activeMediaTab === 'media' }"
          role="tab"
          :aria-selected="activeMediaTab === 'media'"
          aria-controls="media-panel-panel-media"
          @click="selectMediaTab('media')"
          @keydown="onMediaTabKeydown($event, 'media', 'media-panel')"
        >
          <span>{{ t.photosAndVideos }}</span>
          <span v-if="sharedPhotos.length" class="panel-tab-count">{{ sharedPhotos.length }}</span>
        </button>
        <button
          id="media-panel-tab-file"
          type="button"
          class="panel-tab-btn"
          :class="{ active: activeMediaTab === 'file' }"
          role="tab"
          :aria-selected="activeMediaTab === 'file'"
          aria-controls="media-panel-panel-file"
          @click="selectMediaTab('file')"
          @keydown="onMediaTabKeydown($event, 'file', 'media-panel')"
        >
          <span>{{ t.files }}</span>
          <span v-if="sharedFiles.length" class="panel-tab-count">{{ sharedFiles.length }}</span>
        </button>
        <button
          id="media-panel-tab-link"
          type="button"
          class="panel-tab-btn"
          :class="{ active: activeMediaTab === 'link' }"
          role="tab"
          :aria-selected="activeMediaTab === 'link'"
          aria-controls="media-panel-panel-link"
          @click="selectMediaTab('link')"
          @keydown="onMediaTabKeydown($event, 'link', 'media-panel')"
        >
          <span>{{ t.links }}</span>
          <span v-if="sharedLinks.length" class="panel-tab-count">{{ sharedLinks.length }}</span>
        </button>
      </div>

      <!-- Photos & Videos -->
      <div
        v-if="activeMediaTab === 'media'"
        id="media-panel-panel-media"
        class="panel-tab-body"
        role="tabpanel"
        aria-labelledby="media-panel-tab-media"
        tabindex="0"
      >
        <div v-if="sharedPhotos.length" class="panel-media-grid">
          <button
            v-for="item in sharedPhotos"
            :key="item.id"
            type="button"
            class="media-card"
            :disabled="!item.fileId"
            @click="openInlineImage(item)"
          >
            <img
              v-if="item.thumbnailUrl"
              :src="getMediaThumbSrc(item)"
              :alt="item.name"
              class="media-thumb"
              loading="lazy"
            />
            <Icon v-else :name="item.mimeType.startsWith('video/') ? 'video' : 'image'" :size="18" />
            <span class="media-name">{{ item.name }}</span>
            <small>{{ new Date(item.createdAt).toLocaleDateString() }}</small>
          </button>
        </div>
        <p v-else class="muted-hint">{{ t.noSharedMedia }}</p>
      </div>

      <!-- Files -->
      <div
        v-else-if="activeMediaTab === 'file'"
        id="media-panel-panel-file"
        class="panel-tab-body"
        role="tabpanel"
        aria-labelledby="media-panel-tab-file"
        tabindex="0"
      >
        <div v-if="sharedFiles.length" class="panel-files-list">
          <button
            v-for="item in sharedFiles"
            :key="item.id"
            type="button"
            class="panel-file-row"
            @click="openAttachment(item.id)"
          >
            <span class="file-row-icon"><Icon name="file" :size="18" /></span>
            <span class="file-row-info">
              <span class="file-row-name" :title="item.name">{{ item.name }}</span>
              <span class="file-row-meta">{{ formatBytes(item.sizeBytes) }} · {{ new Date(item.createdAt).toLocaleDateString() }}</span>
            </span>
            <Icon name="download" :size="16" class="file-row-dl" />
          </button>
        </div>
        <p v-else class="muted-hint">{{ t.noSharedFiles }}</p>
      </div>

      <!-- Links -->
      <div
        v-else-if="activeMediaTab === 'link'"
        id="media-panel-panel-link"
        class="panel-tab-body"
        role="tabpanel"
        aria-labelledby="media-panel-tab-link"
        tabindex="0"
      >
        <div v-if="sharedLinks.length" class="panel-links-list">
          <div
            v-for="item in sharedLinks"
            :key="item.messageId + '-' + item.url"
            class="panel-link-card"
          >
            <a :href="item.url" target="_blank" rel="noopener noreferrer" class="panel-link-anchor">
              <span class="panel-link-icon"><Icon name="link" :size="16" /></span>
              <span class="panel-link-text">
                <span class="panel-link-domain">{{ item.domain }}</span>
                <span class="panel-link-raw">{{ item.url }}</span>
              </span>
              <Icon name="external-link" :size="13" class="panel-link-ext" />
            </a>
            <div class="panel-link-bottom">
              <span class="panel-link-sender">{{ item.senderName }} · {{ formatRelativeDay(item.createdAt) }}</span>
              <button type="button" class="panel-link-jump" @click="jumpToMessageById(item.messageId)">
                {{ t.jumpToMessage }}
              </button>
            </div>
          </div>
        </div>
        <p v-else class="muted-hint">{{ t.noSharedLinks }}</p>
      </div>
    </aside>

    <!-- Desktop 3rd Column: Chat Info Sidebar (Messenger style) -->
    <aside
      v-if="selectedConversation && isDesktop && desktopInfoOpen"
      class="chat-info-sidebar"
      :aria-label="t.chatInfo"
    >
      <!-- Right resizer splitter (Desktop only) -->
      <div
        class="chat-resizer left-edge"
        :class="{ active: isResizingInfo }"
        role="separator"
        tabindex="0"
        aria-orientation="vertical"
        :aria-valuenow="infoWidth"
        :aria-valuemin="MIN_INFO_WIDTH"
        :aria-valuemax="MAX_INFO_WIDTH"
        :aria-label="t.resizeChatDetailsAria"
        :title="t.resizeChatDetailsTitle"
        @pointerdown.stop.prevent="startInfoResize"
        @dblclick="resetInfoWidth"
        @keydown="onInfoResizerKeydown"
      >
        <div class="resizer-handle-line" />
      </div>
      <div class="desktop-info-header">
        <span class="desktop-info-header-title">{{ chatInfoTitle }}</span>
        <button
          type="button"
          class="icon-btn desktop-info-close-btn"
          :aria-label="t.close"
          :title="t.close"
          @click="closeDesktopInfo"
        >
          <Icon name="close" :size="18" />
        </button>
      </div>

      <div class="chat-info-content">
        <!-- Sub-page Navigation Header for Search, Media and Wallpaper -->
        <div v-if="chatInfoCurrentView !== 'main'" class="info-subpage-nav">
          <button type="button" class="subpage-back-btn" @click="returnToMainInfo">
            <Icon name="arrow-left" :size="18" />
            <span>{{ t.back }}</span>
          </button>
          <span class="subpage-title">{{ chatInfoTitle }}</span>
          <button type="button" class="subpage-close-btn" :aria-label="t.close" @click="closeDesktopInfo">
            <Icon name="close" :size="16" />
          </button>
        </div>

        <!-- 1. MAIN INFO VIEW -->
        <div v-if="chatInfoCurrentView === 'main'" class="info-main-view">
          <div class="chat-info-header">
            <img
              v-if="selectedConversation?.peer?.avatarUrl"
              :src="selectedConversation.peer.avatarUrl"
              class="chat-info-avatar avatar-img"
              alt=""
            />
            <span v-else class="chat-info-avatar" :class="avatarClass(conversationTitle(selectedConversation))" aria-hidden="true">
              {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
            </span>
            <h3 class="chat-info-name">{{ conversationTitle(selectedConversation) }}</h3>
            <span
              v-if="isPeerOnline(selectedConversation) || formatLastSeen(selectedConversation)"
              class="chat-info-status"
              :class="{ online: isPeerOnline(selectedConversation) }"
            >
              {{ isPeerOnline(selectedConversation) ? t.activeNow : formatLastSeen(selectedConversation) }}
            </span>
          </div>

          <div class="chat-info-actions">
            <template v-if="selectedConversation?.type === 'direct'">
              <button
                type="button"
                class="chat-info-action-btn"
                @click="initiateCall(false)"
              >
                <span class="action-icon-circle"><Icon name="phone" :size="18" /></span>
                <span>{{ t.callVoice }}</span>
              </button>
              <button
                type="button"
                class="chat-info-action-btn"
                @click="initiateCall(true)"
              >
                <span class="action-icon-circle"><Icon name="camera" :size="18" /></span>
                <span>{{ t.callVideo }}</span>
              </button>
            </template>
            <button
              type="button"
              class="chat-info-action-btn"
              @click="openInfoSearchView"
            >
              <span class="action-icon-circle"><Icon name="search" :size="18" /></span>
              <span>{{ t.search }}</span>
            </button>
          </div>

          <!-- Section 1: Chat options -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('customization')"
            >
              <span class="menu-section-label">{{ t.thisChatAppearance }}</span>
              <Icon
                :name="infoSectionsOpen.customization ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.customization" class="menu-items-group">
              <div class="theme-picker-row">
                <div class="theme-picker-header">
                  <span class="menu-item-icon badge-theme"><Icon name="palette" :size="18" /></span>
                  <span class="menu-item-text">{{ t.themeColor }}</span>
                </div>
                <div class="theme-dots">
                  <button
                    v-for="th in chatThemes"
                    :key="th.id"
                    type="button"
                    class="theme-dot"
                    :style="{ backgroundColor: th.color }"
                    :class="{ active: activeTheme === th.id }"
                    :title="th.name"
                    :aria-label="th.name"
                    @click="activeTheme = th.id"
                  />
                </div>
              </div>

              <button type="button" class="menu-row-item" @click="openWallpaperSubPage">
                <span class="menu-item-icon badge-theme"><Icon name="image" :size="18" /></span>
                <span class="menu-item-text">{{ t.chatWallpaper }}</span>
                <span class="menu-badge">{{ currentWallpaperPresetName }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="bubblePickerOpen = true">
                <span class="menu-item-icon badge-theme"><Icon name="chat" :size="18" /></span>
                <span class="menu-item-text">{{ t.chatBubbleStyle }}</span>
                <span class="menu-badge">{{ currentBubbleStyle.name }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="changeNickname">
                <span class="menu-item-icon badge-folder"><Icon name="pencil" :size="18" /></span>
                <span class="menu-item-text">{{ t.changeNickname }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="openInfoSearchView">
                <span class="menu-item-icon badge-folder"><Icon name="search" :size="18" /></span>
                <span class="menu-item-text">{{ t.searchInChat }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
            </div>
          </div>

          <!-- Section 2: Media, Files & Links -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('media')"
            >
              <span class="menu-section-label">{{ t.sharedMedia }}</span>
              <Icon
                :name="infoSectionsOpen.media ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.media" class="menu-items-group">
              <button type="button" class="menu-row-item" @click="openMediaSubPage('media')">
                <span class="menu-item-icon badge-theme"><Icon name="photos" :size="18" /></span>
                <span class="menu-item-text">{{ t.photosAndVideos }}</span>
                <span v-if="sharedPhotos.length" class="menu-badge">{{ sharedPhotos.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
              <button type="button" class="menu-row-item" @click="openMediaSubPage('file')">
                <span class="menu-item-icon badge-folder"><Icon name="file" :size="18" /></span>
                <span class="menu-item-text">{{ t.files }}</span>
                <span v-if="sharedFiles.length" class="menu-badge">{{ sharedFiles.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
              <button type="button" class="menu-row-item" @click="openMediaSubPage('link')">
                <span class="menu-item-icon badge-theme"><Icon name="link" :size="18" /></span>
                <span class="menu-item-text">{{ t.links }}</span>
                <span v-if="sharedLinks.length" class="menu-badge">{{ sharedLinks.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
            </div>
          </div>

          <!-- Section 3: Privacy & Actions -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('privacy')"
            >
              <span class="menu-section-label">{{ t.privacyAndSupport }}</span>
              <Icon
                :name="infoSectionsOpen.privacy ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.privacy" class="menu-items-group">
              <button type="button" class="menu-row-item" @click="toggleMuteConversation(selectedConversation.id)">
                <span class="menu-item-icon badge-trash"><Icon :name="mutedConversations[selectedConversation.id] ? 'bell' : 'bell-off'" :size="18" /></span>
                <span class="menu-item-text">{{ mutedConversations[selectedConversation.id] ? t.unmuteChat : t.muteChat }}</span>
              </button>
              <button type="button" class="menu-row-item danger-item" @click="deleteConversation(selectedConversation.id)">
                <span class="menu-item-icon badge-trash"><Icon name="trash" :size="18" /></span>
                <span class="menu-item-text danger-text">{{ t.deleteChat }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 2. SEARCH SUB-PAGE -->
        <div v-else-if="chatInfoCurrentView === 'search'" class="info-search-page">
          <div class="info-search-page-bar">
            <Icon name="search" :size="16" class="info-search-icon" />
            <input
              ref="infoSearchInputRef"
              v-model="infoSearchQuery"
              type="search"
              class="info-search-input"
              :placeholder="t.searchInChat"
              @input="onInfoSearchInput"
              @keydown.enter.prevent="performInfoSearch"
            />
            <span
              v-if="infoSearchDone && infoSearchResults.length"
              class="info-search-results-tag"
            >
              {{ infoSearchResults.length }} kết quả
            </span>
            <button
              v-if="infoSearchQuery"
              type="button"
              class="info-search-clear"
              :aria-label="t.clearSearch"
              @click="clearInfoSearch"
            >
              <Icon name="close" :size="14" />
            </button>
          </div>

          <div v-if="infoSearching" class="subpage-status">
            {{ t.searchingMessages }}
          </div>
          <div v-else-if="!infoSearchQuery.trim()" class="subpage-empty-hint">
            <p class="search-enter-hint">{{ t.pressEnterToSearch }}</p>
          </div>
          <div v-else-if="infoSearchDone && !infoSearchResults.length" class="subpage-empty-hint">
            <p>{{ t.noMessagesFound }}</p>
          </div>
          <div v-else-if="infoSearchResults.length" class="subpage-results-container">
            <div class="info-search-list">
              <button
                v-for="msg in infoSearchResults"
                :key="msg.id"
                type="button"
                class="info-search-item"
                @click="jumpToMessageFromInfo(msg)"
              >
                <div class="search-item-avatar">
                  <img
                    v-if="messageSenderAvatar(msg)"
                    :src="messageSenderAvatar(msg)!"
                    class="avatar-img"
                    alt=""
                  />
                  <span
                    v-else
                    class="avatar-fallback"
                    :class="avatarClass(messageSenderName(msg))"
                  >
                    {{ messageSenderName(msg).slice(0, 1).toUpperCase() }}
                  </span>
                </div>
                <div class="info-search-item-content">
                  <div class="info-search-sender-row">
                    <span class="info-search-sender">{{ messageSenderName(msg) }}</span>
                  </div>
                  <div class="info-search-body info-search-body-row">
                    <span class="info-search-snippet" v-html="formatSearchSnippet(msg.body, infoSearchQuery)" />
                    <span class="info-search-dot">·</span>
                    <span class="info-search-time">{{ formatRelativeDay(msg.createdAt) }}</span>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- 3. MEDIA SUB-PAGE -->
        <div v-else-if="chatInfoCurrentView === 'media'" class="info-media-page">
          <div class="info-tabs-header" role="tablist" :aria-label="t.sharedMedia">
            <button
              id="shared-media-tab-media"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'media' }"
              role="tab"
              :aria-selected="activeMediaTab === 'media'"
              aria-controls="shared-media-panel-media"
              @click="selectMediaTab('media')"
              @keydown="onMediaTabKeydown($event, 'media')"
            >
              <span>{{ t.photosAndVideos }}</span>
              <span v-if="sharedPhotos.length" class="tab-badge">{{ sharedPhotos.length }}</span>
            </button>
            <button
              id="shared-media-tab-file"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'file' }"
              role="tab"
              :aria-selected="activeMediaTab === 'file'"
              aria-controls="shared-media-panel-file"
              @click="selectMediaTab('file')"
              @keydown="onMediaTabKeydown($event, 'file')"
            >
              <span>{{ t.files }}</span>
              <span v-if="sharedFiles.length" class="tab-badge">{{ sharedFiles.length }}</span>
            </button>
            <button
              id="shared-media-tab-link"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'link' }"
              role="tab"
              :aria-selected="activeMediaTab === 'link'"
              aria-controls="shared-media-panel-link"
              @click="selectMediaTab('link')"
              @keydown="onMediaTabKeydown($event, 'link')"
            >
              <span>{{ t.links }}</span>
              <span v-if="sharedLinks.length" class="tab-badge">{{ sharedLinks.length }}</span>
            </button>
          </div>

          <!-- Media Tab -->
          <div
            v-if="activeMediaTab === 'media'"
            id="shared-media-panel-media"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-media"
            tabindex="0"
          >
            <div v-if="sharedPhotos.length" class="chat-info-media-grid">
              <button
                v-for="item in sharedPhotos"
                :key="item.id"
                type="button"
                class="chat-info-media-thumb"
                :title="item.name"
                @click="openInlineImage(item)"
              >
                <img v-if="item.thumbnailUrl" :src="getMediaThumbSrc(item)" :alt="item.name" loading="lazy" />
                <Icon v-else :name="item.mimeType.startsWith('video/') ? 'video' : 'image'" :size="20" />
              </button>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedMedia }}</p>
          </div>

          <!-- Files Tab -->
          <div
            v-else-if="activeMediaTab === 'file'"
            id="shared-media-panel-file"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-file"
            tabindex="0"
          >
            <div v-if="sharedFiles.length" class="chat-info-files-list">
              <button
                v-for="item in sharedFiles"
                :key="item.id"
                type="button"
                class="shared-file-item"
                @click="openAttachment(item.id)"
              >
                <span class="file-icon-wrap"><Icon name="file" :size="20" /></span>
                <span class="file-details">
                  <span class="file-title" :title="item.name">{{ item.name }}</span>
                  <span class="file-sub">{{ formatBytes(item.sizeBytes) }} · {{ formatRelativeDay(item.createdAt) }}</span>
                </span>
                <span class="file-action-icon"><Icon name="download" :size="16" /></span>
              </button>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedFiles }}</p>
          </div>

          <!-- Links Tab -->
          <div
            v-else-if="activeMediaTab === 'link'"
            id="shared-media-panel-link"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-link"
            tabindex="0"
          >
            <div v-if="sharedLinks.length" class="chat-info-links-list">
              <div
                v-for="item in sharedLinks"
                :key="item.messageId + '-' + item.url"
                class="shared-link-card"
              >
                <a :href="item.url" target="_blank" rel="noopener noreferrer" class="shared-link-main">
                  <span class="link-icon-wrap"><Icon name="link" :size="18" /></span>
                  <span class="link-details">
                    <span class="link-domain">{{ item.domain }}</span>
                    <span class="link-url">{{ item.url }}</span>
                  </span>
                  <span class="link-external"><Icon name="external-link" :size="14" /></span>
                </a>
                <div class="link-footer">
                  <span class="link-meta">{{ item.senderName }} · {{ formatRelativeDay(item.createdAt) }}</span>
                  <button type="button" class="link-jump-btn" @click="jumpToMessageById(item.messageId)">
                    {{ t.jumpToMessage }}
                  </button>
                </div>
              </div>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedLinks }}</p>
          </div>
        </div>

        <!-- 4. WALLPAPER SUB-PAGE -->
        <div v-else-if="chatInfoCurrentView === 'wallpaper'" class="info-wallpaper-page">
          <!-- Live Preview Mockup -->
          <div class="wallpaper-preview-card" :style="activePreviewStyle">
            <div class="wallpaper-preview-overlay" />
            <div class="wallpaper-preview-bubbles">
              <div class="preview-bubble incoming">
                <span>{{ conversationTitle(selectedConversation) }}</span>
                <p>{{ t.wallpaperPreviewIncoming }}</p>
              </div>
              <div
                class="preview-bubble outgoing"
                :style="{ background: activeWallpaperTheme ? activeWallpaperTheme.gradient : currentTheme.gradient }"
              >
                <p>{{ t.wallpaperPreviewOutgoing }}</p>
              </div>
            </div>
          </div>

          <!-- Custom Upload, Storage & Reset Actions -->
          <div class="wallpaper-actions-row">
            <button type="button" class="wallpaper-action-btn primary" @click="triggerWallpaperFileInput">
              <Icon name="upload" :size="16" />
              <span>{{ t.uploadCustomWallpaper }}</span>
            </button>
            <button type="button" class="wallpaper-action-btn storage" @click="openMediaStoragePicker">
              <Icon name="folder" :size="16" />
              <span>{{ t.chooseFromStorage }}</span>
            </button>
            <button
              v-if="threadWallpaper"
              type="button"
              class="wallpaper-action-btn secondary"
              @click="removeConversationWallpaper"
            >
              <Icon name="trash" :size="16" />
              <span>{{ t.resetWallpaper }}</span>
            </button>
          </div>
          <input
            ref="wallpaperFileInputRef"
            type="file"
            accept="image/*"
            class="sr-only"
            @change="handleWallpaperUpload"
          />

          <!-- Custom Wallpapers of this Conversation -->
          <div v-if="currentConversationCustomWallpapers.length" class="wallpaper-presets-section wallpaper-custom-section">
            <div class="wallpaper-section-header">
              <span class="wallpaper-section-title">{{ t.customWallpapers }}</span>
              <span class="custom-wallpaper-count">({{ currentConversationCustomWallpapers.length }})</span>
            </div>
            <div class="wallpaper-presets-grid">
              <div
                v-for="wp in currentConversationCustomWallpapers"
                :key="wp.id"
                class="wallpaper-preset-item custom-wallpaper-item"
                :class="{ active: threadWallpaper === wp.url }"
                role="button"
                tabindex="0"
                :title="wp.name"
                @click="setConversationWallpaper(selectedConversation.id, wp.url)"
                @keydown.enter="setConversationWallpaper(selectedConversation.id, wp.url)"
              >
                <div class="wallpaper-preset-thumb custom-thumb" :style="{ backgroundImage: `url('${wp.url}')` }">
                  <span v-if="threadWallpaper === wp.url" class="preset-check-badge">
                    <Icon name="check" :size="14" />
                  </span>
                  <button
                    type="button"
                    class="delete-custom-wp-btn"
                    :title="t.removeWallpaperFromChat"
                    :aria-label="t.removeWallpaperFromChat"
                    @click.stop="removeCustomWallpaperItem(selectedConversation.id, wp.id)"
                  >
                    <Icon name="trash" :size="12" />
                  </button>
                </div>
                <span class="wallpaper-preset-name">{{ wp.name }}</span>
              </div>
            </div>
          </div>

          <!-- Presets Grid -->
          <div class="wallpaper-presets-section">
            <div class="wallpaper-section-header">
              <span class="wallpaper-section-title">{{ t.chooseWallpaper }}</span>
            </div>

            <!-- Mode Filter Tabs -->
            <div class="wallpaper-mode-tabs">
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'all' }"
                @click="wallpaperCategoryFilter = 'all'"
              >
                {{ t.allWallpapers }}
              </button>
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'light' }"
                @click="wallpaperCategoryFilter = 'light'"
              >
                {{ t.lightModeWallpapers }}
              </button>
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'dark' }"
                @click="wallpaperCategoryFilter = 'dark'"
              >
                {{ t.darkModeWallpapers }}
              </button>
            </div>

            <div class="wallpaper-presets-grid">
              <button
                v-for="wp in filteredWallpaperPresets"
                :key="wp.id"
                type="button"
                class="wallpaper-preset-item"
                :class="{ active: (threadWallpaper === wp.id) || (!threadWallpaper && wp.id === 'none') }"
                @click="selectWallpaperPreset(wp.id)"
              >
                <div class="wallpaper-preset-thumb" :style="{ background: wp.preview }">
                  <span v-if="(threadWallpaper === wp.id) || (!threadWallpaper && wp.id === 'none')" class="preset-check-badge">
                    <Icon name="check" :size="14" />
                  </span>
                </div>
                <span class="wallpaper-preset-name">{{ wp.name }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </aside>

    <MediaLightbox
      :open="lightboxOpen"
      :name="lightboxName"
      :mime-type="lightboxMime"
      :url="lightboxUrl"
      @download="downloadLightbox"
      @close="lightboxOpen = false"
    />

    <!-- Media & Personal Storage Picker BottomSheet for Wallpaper -->
    <BottomSheet
      :open="storagePickerOpen"
      :title="t.pickWallpaperFromStorage"
      @close="storagePickerOpen = false"
    >
      <div class="storage-picker-content">
        <div class="storage-picker-tabs" role="tablist">
          <button
            type="button"
            class="storage-tab-btn"
            :class="{ active: storagePickerTab === 'chat' }"
            role="tab"
            :aria-selected="storagePickerTab === 'chat'"
            @click="switchStoragePickerTab('chat')"
          >
            <Icon name="chat" :size="15" />
            <span>{{ t.fromThisChat }} ({{ sharedPhotos.length }})</span>
          </button>
          <button
            type="button"
            class="storage-tab-btn"
            :class="{ active: storagePickerTab === 'personal' }"
            role="tab"
            :aria-selected="storagePickerTab === 'personal'"
            @click="switchStoragePickerTab('personal')"
          >
            <Icon name="folder" :size="15" />
            <span>{{ t.fromPersonalStorage }}</span>
          </button>
        </div>

        <!-- Tab 1: Chat shared photos -->
        <div v-if="storagePickerTab === 'chat'" class="storage-picker-body">
          <div v-if="sharedPhotos.length" class="storage-picker-grid">
            <button
              v-for="photo in sharedPhotos"
              :key="photo.id"
              type="button"
              class="storage-photo-card"
              :title="photo.name || photo.originalName"
              @click="selectPhotoFromChatMedia(photo)"
            >
              <img
                v-if="photo.thumbnailUrl"
                :src="photo.thumbnailUrl"
                :alt="photo.name"
                loading="lazy"
              />
              <div v-else class="storage-photo-fallback">
                <Icon name="image" :size="24" />
              </div>
              <span class="storage-photo-name">{{ photo.name || photo.originalName }}</span>
            </button>
          </div>
          <div v-else class="storage-picker-empty">
            <Icon name="image" :size="36" />
            <p>{{ t.noChatPhotos }}</p>
          </div>
        </div>

        <!-- Tab 2: Personal Storage Photos -->
        <div v-else-if="storagePickerTab === 'personal'" class="storage-picker-body">
          <div v-if="loadingPersonalPhotos" class="storage-picker-loading">
            <span class="spinner" />
            <p>{{ t.loadingPhotos }}</p>
          </div>
          <p v-else-if="personalPhotosError" class="alert">{{ personalPhotosError }}</p>
          <div v-else-if="personalPhotos.length" class="storage-picker-grid">
            <button
              v-for="item in personalPhotos"
              :key="item.id"
              type="button"
              class="storage-photo-card"
              :title="item.name"
              @click="selectPhotoFromPersonalStorage(item)"
            >
              <img
                v-if="item.thumbnailUrl"
                :src="item.thumbnailUrl"
                :alt="item.name"
                loading="lazy"
              />
              <div v-else class="storage-photo-fallback">
                <Icon name="image" :size="24" />
              </div>
              <span class="storage-photo-name">{{ item.name }}</span>
            </button>
          </div>
          <div v-else class="storage-picker-empty">
            <Icon name="folder" :size="36" />
            <p>{{ t.noPersonalPhotos }}</p>
          </div>
        </div>
      </div>
    </BottomSheet>

    <BottomSheet :open="appMenuOpen" :title="t.filvaultMenu" @close="appMenuOpen = false">
      <div class="app-menu-content">
        <!-- User Profile Card -->
        <button type="button" class="menu-profile-card" @click="navigateTo('/profile')">
          <img v-if="auth.user?.avatarUrl" :src="auth.user.avatarUrl" class="menu-avatar avatar-img" alt="" />
          <span v-else class="menu-avatar" aria-hidden="true">{{ userInitials(auth.user?.displayName ?? '', auth.user?.email ?? '') }}</span>
          <span class="menu-profile-info">
            <span class="menu-profile-name">{{ auth.user?.displayName || t.myAccount }}</span>
            <span class="menu-profile-email">{{ auth.user?.email }}</span>
            <span class="menu-status-badge" :class="{ offline: auth.user?.activeStatusEnabled === false }">
              <span class="status-dot" :class="{ offline: auth.user?.activeStatusEnabled === false }" aria-hidden="true" />
              <span>{{ t.activeStatus }}: {{ auth.user?.activeStatusEnabled !== false ? t.activeStatusOn : t.activeStatusOff }}</span>
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
            <button type="button" class="menu-row-item" @click="navigateTo('/settings#appearance')">
              <span class="menu-item-icon badge-theme">
                <Icon :name="resolvedIsDark ? 'moon' : 'sun'" :size="18" />
              </span>
              <span class="menu-item-text">{{ t.appAppearance }}</span>
              <span class="menu-toggle-state">{{ resolvedIsDark ? t.appearanceModeDark : t.appearanceModeLight }}</span>
              <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
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
            <button
              type="button"
              class="menu-row-item"
              @click="toggleActiveStatus"
            >
              <span class="menu-item-icon badge-status" :class="{ offline: auth.user?.activeStatusEnabled === false }">
                <span class="status-dot" :class="{ offline: auth.user?.activeStatusEnabled === false }" />
              </span>
              <span class="menu-item-text">{{ t.activeStatus }}</span>
              <span :class="auth.user?.activeStatusEnabled !== false ? 'menu-badge-green' : 'menu-badge-muted'">
                {{ auth.user?.activeStatusEnabled !== false ? t.activeStatusOn : t.activeStatusOff }}
              </span>
            </button>
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

    <!-- Messenger Message Reaction & Quick Action Overlay -->
    <Teleport to="body">
      <div
        v-if="messageMenuOpen && activeMessage"
        class="reaction-modal-overlay"
        @click.self="messageMenuOpen = false"
      >
        <!-- Floating Reaction Capsule -->
        <div
          class="reaction-capsule-wrap"
          :style="capsulePos ? { top: `${capsulePos.top}px`, left: `${capsulePos.left}px` } : undefined"
          @click.stop
        >
          <div class="reaction-capsule">
            <button
              v-for="rx in QUICK_REACTIONS"
              :key="rx"
              type="button"
              class="rx-btn"
              :class="{ 'rx-active': isReactedWith(activeMessage, rx) }"
              :aria-label="rx"
              @click="selectReaction(rx)"
            >
              <span class="rx-char">{{ rx }}</span>
            </button>
            <button
              type="button"
              class="rx-btn rx-plus-btn"
              :title="t.addReaction"
              :aria-label="t.addReaction"
              @click="openReactionPicker"
            >
              <Icon name="plus" :size="18" />
            </button>
          </div>
        </div>

        <!-- Active Message Bubble (elevated in focus, precisely aligned with original bubble) -->
        <div
          class="reaction-active-bubble-wrap"
          :style="activeBubbleRect ? {
            top: `${activeBubbleRect.top}px`,
            left: `${activeBubbleRect.left}px`,
            width: `${activeBubbleRect.width}px`,
          } : undefined"
          @click="messageMenuOpen = false"
        >
          <div
            class="message-bubble active-elevated"
            :class="{
              outgoing: activeMessage.senderId === auth.user?.id,
              'has-like': activeMessage.body === '👍',
              'has-sticker': isStickerMessage(activeMessage.body)
            }"
            :style="activeBubbleRect ? { width: '100%', maxWidth: '100%' } : undefined"
          >
            <p v-if="activeMessage.body && isStickerMessage(activeMessage.body)" class="sticker-bubble">{{ parseStickerSymbol(activeMessage.body) }}</p>
            <p v-else-if="activeMessage.body" :class="{ 'like-bubble': activeMessage.body === '👍' }">{{ activeMessage.body }}</p>
            <span class="bubble-time">{{ formatTime(activeMessage.createdAt) }}</span>
          </div>
        </div>

        <!-- Bottom Quick Actions Bar (Messenger style: Trả lời, Sao chép, Sửa, Xóa, Đóng) -->
        <div class="reaction-bottom-bar" @click.stop>
          <button type="button" class="bottom-action-btn" @click="triggerReplyMessage">
            <div class="action-icon-circle">
              <Icon name="reply" :size="20" />
            </div>
            <span>{{ t.reply }}</span>
          </button>
          <button type="button" class="bottom-action-btn" @click="copyMessageText">
            <div class="action-icon-circle">
              <Icon name="copy" :size="20" />
            </div>
            <span>{{ t.copyMessage }}</span>
          </button>
          <button
            v-if="canMutateMessage(activeMessage)"
            type="button"
            class="bottom-action-btn"
            @click="triggerEditMessage"
          >
            <div class="action-icon-circle">
              <Icon name="pencil" :size="20" />
            </div>
            <span>{{ t.edit }}</span>
          </button>
          <button
            v-if="canMutateMessage(activeMessage)"
            type="button"
            class="bottom-action-btn danger-action"
            @click="triggerRemoveMessage"
          >
            <div class="action-icon-circle danger-circle">
              <Icon name="trash" :size="20" />
            </div>
            <span>{{ t.remove }}</span>
          </button>
          <button type="button" class="bottom-action-btn" @click="messageMenuOpen = false">
            <div class="action-icon-circle">
              <Icon name="close" :size="20" />
            </div>
            <span>{{ t.close }}</span>
          </button>
        </div>
      </div>
    </Teleport>

    <!-- Custom Emoji Reaction Picker BottomSheet -->
    <BottomSheet :open="reactionPickerOpen" :title="t.chooseReaction" @close="reactionPickerOpen = false">
      <EmojiPicker
        :active-reactions="targetMessageReactionList"
        @select-emoji="handlePickReactionEmoji"
        @close="reactionPickerOpen = false"
      />
    </BottomSheet>

    <!-- Conversation Long-press Menu -->
    <BottomSheet :open="convMenuOpen" :title="t.chatOptions" @close="convMenuOpen = false">
      <div v-if="activeConv" class="sheet-action-list">
        <div class="sheet-message-preview">
          <strong>{{ conversationTitle(activeConv) }}</strong>
        </div>
        <button type="button" class="sheet-action-item" @click="openPeekPreview(activeConv)">
          <Icon name="eye" :size="18" />
          <span>{{ t.peekPreview }}</span>
        </button>
        <button type="button" class="sheet-action-item" @click="selectConversation(activeConv.id); convMenuOpen = false">
          <Icon name="chat" :size="18" />
          <span>{{ t.openChat }}</span>
        </button>
        <button type="button" class="sheet-action-item" @click="toggleMuteConversation(activeConv.id)">
          <Icon :name="mutedConversations[activeConv.id] ? 'bell' : 'bell-off'" :size="18" />
          <span>{{ mutedConversations[activeConv.id] ? t.unmuteChat : t.muteChat }}</span>
        </button>
        <button type="button" class="sheet-action-item danger-text" @click="deleteConversation(activeConv.id)">
          <Icon name="trash" :size="18" />
          <span>{{ t.deleteChat }}</span>
        </button>
      </div>
    </BottomSheet>

    <!-- Peek Preview BottomSheet (Xem trước không dính đã xem) -->
    <BottomSheet :open="peekOpen" :title="t.peekPreview" @close="peekOpen = false">
      <div v-if="peekConv" class="peek-preview-content">
        <div class="peek-preview-header">
          <img
            v-if="peekConv?.peer?.avatarUrl"
            :src="peekConv.peer.avatarUrl"
            class="peek-avatar avatar-img"
            alt=""
          />
          <span v-else class="peek-avatar" :class="avatarClass(conversationTitle(peekConv))" aria-hidden="true">
            {{ conversationTitle(peekConv).slice(0, 1).toUpperCase() }}
          </span>
          <div class="peek-meta">
            <h4>{{ conversationTitle(peekConv) }}</h4>
            <span class="peek-badge">{{ t.peekUnreadBadge }}</span>
          </div>
        </div>

        <div v-if="peekLoading" class="peek-loading">
          <p>{{ t.peekLoading }}</p>
        </div>
        <div v-else-if="!peekMessages.length" class="peek-empty">
          <p>{{ t.noMessagesYet }}</p>
        </div>
        <div v-else class="peek-messages-list">
          <div
            v-for="msg in peekMessages"
            :key="msg.id"
            class="peek-message-row"
            :class="{ outgoing: msg.senderId === auth.user?.id }"
          >
            <div
              class="peek-bubble"
              :class="{
                outgoing: msg.senderId === auth.user?.id,
                'has-like': msg.body === '👍',
                'has-sticker': isStickerMessage(msg.body)
              }"
            >
              <p v-if="msg.body && isStickerMessage(msg.body)" class="sticker-bubble">{{ parseStickerSymbol(msg.body) }}</p>
              <p v-else-if="msg.body" :class="{ 'like-bubble': msg.body === '👍' }">{{ msg.body }}</p>
              <span class="peek-time">{{ formatTime(msg.createdAt) }}</span>
            </div>
          </div>
        </div>

        <div class="peek-actions">
          <button type="button" class="btn ink peek-open-btn" @click="openChatFromPeek">
            <Icon name="chat" :size="16" />
            <span>{{ t.openChat }}</span>
          </button>
        </div>
      </div>
    </BottomSheet>

    <!-- Thread Info / Settings (Mobile BottomSheet) -->
    <BottomSheet v-if="!isDesktop" :open="threadInfoOpen" :title="chatInfoCurrentView === 'main' ? chatInfoTitle : undefined" @close="closeChatInfo">
      <div v-if="selectedConversation" class="chat-info-content">
        <!-- Sub-page Navigation Header for Search, Media and Wallpaper -->
        <div v-if="chatInfoCurrentView !== 'main'" class="info-subpage-nav">
          <button type="button" class="subpage-back-btn" @click="returnToMainInfo">
            <Icon name="arrow-left" :size="18" />
            <span>{{ t.back }}</span>
          </button>
          <span class="subpage-title">{{ chatInfoTitle }}</span>
          <button type="button" class="subpage-close-btn" :aria-label="t.close" @click="closeChatInfo">
            <Icon name="close" :size="16" />
          </button>
        </div>

        <!-- 1. MAIN INFO VIEW -->
        <div v-if="chatInfoCurrentView === 'main'" class="info-main-view">
          <div class="chat-info-header">
            <img
              v-if="selectedConversation?.peer?.avatarUrl"
              :src="selectedConversation.peer.avatarUrl"
              class="chat-info-avatar avatar-img"
              alt=""
            />
            <span v-else class="chat-info-avatar" :class="avatarClass(conversationTitle(selectedConversation))" aria-hidden="true">
              {{ conversationTitle(selectedConversation).slice(0, 1).toUpperCase() }}
            </span>
            <h3 class="chat-info-name">{{ conversationTitle(selectedConversation) }}</h3>
            <span
              v-if="isPeerOnline(selectedConversation) || formatLastSeen(selectedConversation)"
              class="chat-info-status"
              :class="{ online: isPeerOnline(selectedConversation) }"
            >
              {{ isPeerOnline(selectedConversation) ? t.activeNow : formatLastSeen(selectedConversation) }}
            </span>
          </div>

          <div class="chat-info-actions">
            <template v-if="selectedConversation?.type === 'direct'">
              <button
                type="button"
                class="chat-info-action-btn"
                @click="initiateCall(false); threadInfoOpen = false"
              >
                <span class="action-icon-circle"><Icon name="phone" :size="18" /></span>
                <span>{{ t.callVoice }}</span>
              </button>
              <button
                type="button"
                class="chat-info-action-btn"
                @click="initiateCall(true); threadInfoOpen = false"
              >
                <span class="action-icon-circle"><Icon name="camera" :size="18" /></span>
                <span>{{ t.callVideo }}</span>
              </button>
            </template>
            <button
              type="button"
              class="chat-info-action-btn"
              @click="openInfoSearchView"
            >
              <span class="action-icon-circle"><Icon name="search" :size="18" /></span>
              <span>{{ t.search }}</span>
            </button>
          </div>

          <!-- Section 1: Chat options -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('customization')"
            >
              <span class="menu-section-label">{{ t.thisChatAppearance }}</span>
              <Icon
                :name="infoSectionsOpen.customization ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.customization" class="menu-items-group">
              <div class="theme-picker-row">
                <div class="theme-picker-header">
                  <span class="menu-item-icon badge-theme"><Icon name="palette" :size="18" /></span>
                  <span class="menu-item-text">{{ t.themeColor }}</span>
                </div>
                <div class="theme-dots">
                  <button
                    v-for="th in chatThemes"
                    :key="th.id"
                    type="button"
                    class="theme-dot"
                    :style="{ backgroundColor: th.color }"
                    :class="{ active: activeTheme === th.id }"
                    :title="th.name"
                    :aria-label="th.name"
                    @click="activeTheme = th.id"
                  />
                </div>
              </div>

              <button type="button" class="menu-row-item" @click="openWallpaperSubPage">
                <span class="menu-item-icon badge-theme"><Icon name="image" :size="18" /></span>
                <span class="menu-item-text">{{ t.chatWallpaper }}</span>
                <span class="menu-badge">{{ currentWallpaperPresetName }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="bubblePickerOpen = true">
                <span class="menu-item-icon badge-theme"><Icon name="chat" :size="18" /></span>
                <span class="menu-item-text">{{ t.chatBubbleStyle }}</span>
                <span class="menu-badge">{{ currentBubbleStyle.name }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="changeNickname">
                <span class="menu-item-icon badge-folder"><Icon name="pencil" :size="18" /></span>
                <span class="menu-item-text">{{ t.changeNickname }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>

              <button type="button" class="menu-row-item" @click="openInfoSearchView">
                <span class="menu-item-icon badge-folder"><Icon name="search" :size="18" /></span>
                <span class="menu-item-text">{{ t.searchInChat }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
            </div>
          </div>

          <!-- Section 2: Media, Files & Links (Compact Messenger-style rows) -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('media')"
            >
              <span class="menu-section-label">{{ t.sharedMedia }}</span>
              <Icon
                :name="infoSectionsOpen.media ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.media" class="menu-items-group">
              <button type="button" class="menu-row-item" @click="openMediaSubPage('media')">
                <span class="menu-item-icon badge-theme"><Icon name="photos" :size="18" /></span>
                <span class="menu-item-text">{{ t.photosAndVideos }}</span>
                <span v-if="sharedPhotos.length" class="menu-badge">{{ sharedPhotos.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
              <button type="button" class="menu-row-item" @click="openMediaSubPage('file')">
                <span class="menu-item-icon badge-folder"><Icon name="file" :size="18" /></span>
                <span class="menu-item-text">{{ t.files }}</span>
                <span v-if="sharedFiles.length" class="menu-badge">{{ sharedFiles.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
              <button type="button" class="menu-row-item" @click="openMediaSubPage('link')">
                <span class="menu-item-icon badge-theme"><Icon name="link" :size="18" /></span>
                <span class="menu-item-text">{{ t.links }}</span>
                <span v-if="sharedLinks.length" class="menu-badge">{{ sharedLinks.length }}</span>
                <Icon name="chevron-right" :size="16" class="menu-item-arrow" />
              </button>
            </div>
          </div>

          <!-- Section 3: Privacy & Actions -->
          <div class="menu-section">
            <button
              type="button"
              class="menu-section-header-btn"
              @click="toggleInfoSection('privacy')"
            >
              <span class="menu-section-label">{{ t.privacyAndSupport }}</span>
              <Icon
                :name="infoSectionsOpen.privacy ? 'chevron-up' : 'chevron-down'"
                :size="16"
                class="section-chevron"
              />
            </button>
            <div v-show="infoSectionsOpen.privacy" class="menu-items-group">
              <button type="button" class="menu-row-item" @click="toggleMuteConversation(selectedConversation.id)">
                <span class="menu-item-icon badge-trash"><Icon :name="mutedConversations[selectedConversation.id] ? 'bell' : 'bell-off'" :size="18" /></span>
                <span class="menu-item-text">{{ mutedConversations[selectedConversation.id] ? t.unmuteChat : t.muteChat }}</span>
              </button>
              <button type="button" class="menu-row-item danger-item" @click="deleteConversation(selectedConversation.id)">
                <span class="menu-item-icon badge-trash"><Icon name="trash" :size="18" /></span>
                <span class="menu-item-text danger-text">{{ t.deleteChat }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 2. DEDICATED SEARCH SUB-PAGE (Messenger style) -->
        <div v-else-if="chatInfoCurrentView === 'search'" class="info-search-page">
          <div class="info-search-page-bar">
            <Icon name="search" :size="16" class="info-search-icon" />
            <input
              ref="infoSearchInputRef"
              v-model="infoSearchQuery"
              type="search"
              class="info-search-input"
              :placeholder="t.searchInChat"
              @input="onInfoSearchInput"
              @keydown.enter.prevent="performInfoSearch"
            />
            <span
              v-if="infoSearchDone && infoSearchResults.length"
              class="info-search-results-tag"
            >
              {{ infoSearchResults.length }} kết quả
            </span>
            <button
              v-if="infoSearchQuery"
              type="button"
              class="info-search-clear"
              :aria-label="t.clearSearch"
              @click="clearInfoSearch"
            >
              <Icon name="close" :size="14" />
            </button>
          </div>

          <div v-if="infoSearching" class="subpage-status">
            {{ t.searchingMessages }}
          </div>
          <div v-else-if="!infoSearchQuery.trim()" class="subpage-empty-hint">
            <p class="search-enter-hint">{{ t.pressEnterToSearch }}</p>
          </div>
          <div v-else-if="infoSearchDone && !infoSearchResults.length" class="subpage-empty-hint">
            <p>{{ t.noMessagesFound }}</p>
          </div>
          <div v-else-if="infoSearchResults.length" class="subpage-results-container">
            <div class="info-search-list">
              <button
                v-for="msg in infoSearchResults"
                :key="msg.id"
                type="button"
                class="info-search-item"
                @click="jumpToMessageFromInfo(msg)"
              >
                <div class="search-item-avatar">
                  <img
                    v-if="messageSenderAvatar(msg)"
                    :src="messageSenderAvatar(msg)!"
                    class="avatar-img"
                    alt=""
                  />
                  <span
                    v-else
                    class="avatar-fallback"
                    :class="avatarClass(messageSenderName(msg))"
                  >
                    {{ messageSenderName(msg).slice(0, 1).toUpperCase() }}
                  </span>
                </div>
                <div class="info-search-item-content">
                  <div class="info-search-sender-row">
                    <span class="info-search-sender">{{ messageSenderName(msg) }}</span>
                  </div>
                  <div class="info-search-body info-search-body-row">
                    <span class="info-search-snippet" v-html="formatSearchSnippet(msg.body, infoSearchQuery)" />
                    <span class="info-search-dot">·</span>
                    <span class="info-search-time">{{ formatRelativeDay(msg.createdAt) }}</span>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- 3. DEDICATED MEDIA SUB-PAGE (Messenger style) -->
        <div v-else-if="chatInfoCurrentView === 'media'" class="info-media-page">
          <div class="info-tabs-header" role="tablist" :aria-label="t.sharedMedia">
            <button
              id="shared-media-tab-media"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'media' }"
              role="tab"
              :aria-selected="activeMediaTab === 'media'"
              aria-controls="shared-media-panel-media"
              @click="selectMediaTab('media')"
              @keydown="onMediaTabKeydown($event, 'media')"
            >
              <span>{{ t.photosAndVideos }}</span>
              <span v-if="sharedPhotos.length" class="tab-badge">{{ sharedPhotos.length }}</span>
            </button>
            <button
              id="shared-media-tab-file"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'file' }"
              role="tab"
              :aria-selected="activeMediaTab === 'file'"
              aria-controls="shared-media-panel-file"
              @click="selectMediaTab('file')"
              @keydown="onMediaTabKeydown($event, 'file')"
            >
              <span>{{ t.files }}</span>
              <span v-if="sharedFiles.length" class="tab-badge">{{ sharedFiles.length }}</span>
            </button>
            <button
              id="shared-media-tab-link"
              type="button"
              class="info-tab-btn"
              :class="{ active: activeMediaTab === 'link' }"
              role="tab"
              :aria-selected="activeMediaTab === 'link'"
              aria-controls="shared-media-panel-link"
              @click="selectMediaTab('link')"
              @keydown="onMediaTabKeydown($event, 'link')"
            >
              <span>{{ t.links }}</span>
              <span v-if="sharedLinks.length" class="tab-badge">{{ sharedLinks.length }}</span>
            </button>
          </div>

          <!-- Media Tab (Photos & Videos) -->
          <div
            v-if="activeMediaTab === 'media'"
            id="shared-media-panel-media"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-media"
            tabindex="0"
          >
            <div v-if="sharedPhotos.length" class="chat-info-media-grid">
              <button
                v-for="item in sharedPhotos"
                :key="item.id"
                type="button"
                class="chat-info-media-thumb"
                :title="item.name"
                @click="openInlineImage(item)"
              >
                <img v-if="item.thumbnailUrl" :src="getMediaThumbSrc(item)" :alt="item.name" loading="lazy" />
                <Icon v-else :name="item.mimeType.startsWith('video/') ? 'video' : 'image'" :size="20" />
              </button>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedMedia }}</p>
          </div>

          <!-- Files Tab -->
          <div
            v-else-if="activeMediaTab === 'file'"
            id="shared-media-panel-file"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-file"
            tabindex="0"
          >
            <div v-if="sharedFiles.length" class="chat-info-files-list">
              <button
                v-for="item in sharedFiles"
                :key="item.id"
                type="button"
                class="shared-file-item"
                @click="openAttachment(item.id)"
              >
                <span class="file-icon-wrap"><Icon name="file" :size="20" /></span>
                <span class="file-details">
                  <span class="file-title" :title="item.name">{{ item.name }}</span>
                  <span class="file-sub">{{ formatBytes(item.sizeBytes) }} · {{ formatRelativeDay(item.createdAt) }}</span>
                </span>
                <span class="file-action-icon"><Icon name="download" :size="16" /></span>
              </button>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedFiles }}</p>
          </div>

          <!-- Links Tab -->
          <div
            v-else-if="activeMediaTab === 'link'"
            id="shared-media-panel-link"
            class="tab-content"
            role="tabpanel"
            aria-labelledby="shared-media-tab-link"
            tabindex="0"
          >
            <div v-if="sharedLinks.length" class="chat-info-links-list">
              <div
                v-for="item in sharedLinks"
                :key="item.messageId + '-' + item.url"
                class="shared-link-card"
              >
                <a :href="item.url" target="_blank" rel="noopener noreferrer" class="shared-link-main">
                  <span class="link-icon-wrap"><Icon name="link" :size="18" /></span>
                  <span class="link-details">
                    <span class="link-domain">{{ item.domain }}</span>
                    <span class="link-url">{{ item.url }}</span>
                  </span>
                  <span class="link-external"><Icon name="external-link" :size="14" /></span>
                </a>
                <div class="link-footer">
                  <span class="link-meta">{{ item.senderName }} · {{ formatRelativeDay(item.createdAt) }}</span>
                  <button type="button" class="link-jump-btn" @click="jumpToMessageById(item.messageId)">
                    {{ t.jumpToMessage }}
                  </button>
                </div>
              </div>
            </div>
            <p v-else class="tab-empty-hint">{{ t.noSharedLinks }}</p>
          </div>
        </div>

        <!-- 4. DEDICATED WALLPAPER SUB-PAGE -->
        <div v-else-if="chatInfoCurrentView === 'wallpaper'" class="info-wallpaper-page">
          <!-- Live Preview Mockup -->
          <div class="wallpaper-preview-card" :style="activePreviewStyle">
            <div class="wallpaper-preview-overlay" />
            <div class="wallpaper-preview-bubbles">
              <div class="preview-bubble incoming">
                <span>{{ conversationTitle(selectedConversation) }}</span>
                <p>{{ t.wallpaperPreviewIncoming }}</p>
              </div>
              <div
                class="preview-bubble outgoing"
                :style="{ background: activeWallpaperTheme ? activeWallpaperTheme.gradient : currentTheme.gradient }"
              >
                <p>{{ t.wallpaperPreviewOutgoing }}</p>
              </div>
            </div>
          </div>

          <!-- Custom Upload, Storage & Reset Actions -->
          <div class="wallpaper-actions-row">
            <button type="button" class="wallpaper-action-btn primary" @click="triggerWallpaperFileInput">
              <Icon name="upload" :size="16" />
              <span>{{ t.uploadCustomWallpaper }}</span>
            </button>
            <button type="button" class="wallpaper-action-btn storage" @click="openMediaStoragePicker">
              <Icon name="folder" :size="16" />
              <span>{{ t.chooseFromStorage }}</span>
            </button>
            <button
              v-if="threadWallpaper"
              type="button"
              class="wallpaper-action-btn secondary"
              @click="removeConversationWallpaper"
            >
              <Icon name="trash" :size="16" />
              <span>{{ t.resetWallpaper }}</span>
            </button>
          </div>
          <input
            ref="wallpaperFileInputRef"
            type="file"
            accept="image/*"
            class="sr-only"
            @change="handleWallpaperUpload"
          />

          <!-- Custom Wallpapers of this Conversation -->
          <div v-if="currentConversationCustomWallpapers.length" class="wallpaper-presets-section wallpaper-custom-section">
            <div class="wallpaper-section-header">
              <span class="wallpaper-section-title">{{ t.customWallpapers }}</span>
              <span class="custom-wallpaper-count">({{ currentConversationCustomWallpapers.length }})</span>
            </div>
            <div class="wallpaper-presets-grid">
              <div
                v-for="wp in currentConversationCustomWallpapers"
                :key="wp.id"
                class="wallpaper-preset-item custom-wallpaper-item"
                :class="{ active: threadWallpaper === wp.url }"
                role="button"
                tabindex="0"
                :title="wp.name"
                @click="setConversationWallpaper(selectedConversation.id, wp.url)"
                @keydown.enter="setConversationWallpaper(selectedConversation.id, wp.url)"
              >
                <div class="wallpaper-preset-thumb custom-thumb" :style="{ backgroundImage: `url('${wp.url}')` }">
                  <span v-if="threadWallpaper === wp.url" class="preset-check-badge">
                    <Icon name="check" :size="14" />
                  </span>
                  <button
                    type="button"
                    class="delete-custom-wp-btn"
                    :title="t.removeWallpaperFromChat"
                    :aria-label="t.removeWallpaperFromChat"
                    @click.stop="removeCustomWallpaperItem(selectedConversation.id, wp.id)"
                  >
                    <Icon name="trash" :size="12" />
                  </button>
                </div>
                <span class="wallpaper-preset-name">{{ wp.name }}</span>
              </div>
            </div>
          </div>

          <!-- Presets Grid -->
          <div class="wallpaper-presets-section">
            <div class="wallpaper-section-header">
              <span class="wallpaper-section-title">{{ t.chooseWallpaper }}</span>
            </div>

            <!-- Mode Filter Tabs -->
            <div class="wallpaper-mode-tabs">
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'all' }"
                @click="wallpaperCategoryFilter = 'all'"
              >
                {{ t.allWallpapers }}
              </button>
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'light' }"
                @click="wallpaperCategoryFilter = 'light'"
              >
                {{ t.lightModeWallpapers }}
              </button>
              <button
                type="button"
                class="wallpaper-mode-tab"
                :class="{ active: wallpaperCategoryFilter === 'dark' }"
                @click="wallpaperCategoryFilter = 'dark'"
              >
                {{ t.darkModeWallpapers }}
              </button>
            </div>

            <div class="wallpaper-presets-grid">
              <button
                v-for="wp in filteredWallpaperPresets"
                :key="wp.id"
                type="button"
                class="wallpaper-preset-item"
                :class="{ active: (threadWallpaper === wp.id) || (!threadWallpaper && wp.id === 'none') }"
                @click="selectWallpaperPreset(wp.id)"
              >
                <div class="wallpaper-preset-thumb" :style="{ background: wp.preview }">
                  <span v-if="(threadWallpaper === wp.id) || (!threadWallpaper && wp.id === 'none')" class="preset-check-badge">
                    <Icon name="check" :size="14" />
                  </span>
                </div>
                <span class="wallpaper-preset-label">{{ wp.name }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </BottomSheet>

    <!-- Modal Chọn kiểu bong bóng -->
    <ChatBubblePickerModal
      :open="bubblePickerOpen"
      @saved="bubblePickerOpen = false"
      @close="bubblePickerOpen = false"
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
  --chat-accent: var(--accent);
  --chat-accent-secondary: var(--accent);
  --chat-bubble-outgoing: linear-gradient(135deg, var(--chat-accent) 0%, var(--chat-accent-secondary) 100%);
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
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--sidebar-bg, var(--canvas));
  border-right: 1px solid var(--hairline);
}

.rail-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  height: 60px;
  min-height: 60px;
  box-sizing: border-box;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
}

.rail-header h1 {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
  background: var(--chat-bubble-outgoing);
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

.menu-status-badge.offline {
  color: var(--muted);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-pill);
  background: #22c55e;
  flex-shrink: 0;
}

.status-dot.offline {
  background: var(--muted);
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

.menu-section-header-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  border-radius: var(--radius-sm);
  color: var(--ink);
  transition: background var(--duration-short) var(--ease-standard);
  text-align: left;
}

.menu-section-header-btn:hover {
  background: var(--surface-soft);
}

.menu-section-header-btn .menu-section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
  text-transform: none;
  letter-spacing: normal;
  padding: 0;
  cursor: pointer;
}

.menu-section-header-btn .section-chevron {
  color: var(--muted);
  transition: transform var(--duration-short) var(--ease-standard);
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
.badge-status.offline { background: var(--surface-soft); color: var(--muted); }
.badge-archive { background: var(--surface-soft); color: var(--muted); }

:global([data-theme='dark']) .badge-folder { background: rgba(2, 132, 199, 0.22); color: #38bdf8; }
:global([data-theme='dark']) .badge-photos { background: rgba(162, 28, 175, 0.22); color: #e879f9; }
:global([data-theme='dark']) .badge-trash { background: rgba(220, 38, 38, 0.22); color: #f87171; }
:global([data-theme='dark']) .badge-theme { background: rgba(217, 119, 6, 0.22); color: #fbbf24; }
:global([data-theme='dark']) .badge-lang { background: rgba(67, 56, 202, 0.22); color: #818cf8; }
:global([data-theme='dark']) .badge-status { background: rgba(22, 163, 74, 0.22); color: #4ade80; }
:global([data-theme='dark']) .badge-status.offline { background: rgba(255, 255, 255, 0.08); color: var(--muted); }

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
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.conversation-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-xs);
  min-width: 0;
}

.conversation-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ink);
  font-weight: 600;
  min-width: 0;
  flex: 1;
}

.conversation-date {
  color: var(--muted);
  font-size: 12px;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: auto;
}

.conversation-row.unread .conversation-title {
  font-weight: 700;
  color: var(--ink);
}

.conversation-row.unread .conversation-preview {
  font-weight: 600;
  color: var(--ink);
}

.unread-date {
  color: var(--accent) !important;
  font-weight: 600;
}

.conversation-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  min-width: 0;
}

.conversation-preview {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
  font-size: 13px;
  min-width: 0;
  flex: 1;
}

.unread-badge {
  min-width: 8px;
  height: 8px;
  border-radius: var(--radius-pill);
  background: var(--accent);
  flex-shrink: 0;
}

.unread-badge:not(:empty) {
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  font-size: 11px;
  font-weight: 700;
  color: #ffffff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.conversation-typing {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-style: italic;
  font-size: 13px;
  color: var(--accent);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.typing-pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-pill);
  background: var(--accent);
  animation: pulseDot 1s infinite alternate;
  flex-shrink: 0;
}

@keyframes pulseDot {
  0% { transform: scale(0.8); opacity: 0.5; }
  100% { transform: scale(1.2); opacity: 1; }
}

.message-thread {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--canvas);
  overflow: hidden;
}

.thread-canvas-wrap {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  background: var(--canvas);
}

/* Fixed Wallpaper Backdrop: stays locked to canvas viewport between header and composer */
.thread-wallpaper-backdrop {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  z-index: 0;
}

.wallpaper-scrim {
  position: absolute;
  inset: 0;
  background: var(--chat-scrim-overlay, rgba(255, 255, 255, 0.12));
  transition: background var(--duration-short) var(--ease-standard);
}

:global([data-theme='dark']) .wallpaper-scrim,
[data-theme='dark'] .wallpaper-scrim {
  background: var(--chat-scrim-overlay, rgba(0, 0, 0, 0.42));
}

.wallpaper-scrim.is-dark {
  background: var(--chat-scrim-overlay, rgba(0, 0, 0, 0.28));
}

:global([data-theme='dark']) .wallpaper-scrim.is-dark,
[data-theme='dark'] .wallpaper-scrim.is-dark {
  background: var(--chat-scrim-overlay, rgba(0, 0, 0, 0.50));
}

.thread-header {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  height: 60px;
  min-height: 60px;
  box-sizing: border-box;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
  transition: background var(--duration-short) var(--ease-standard), border-color var(--duration-short) var(--ease-standard);
}

.thread-peer-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  padding: 4px 8px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background var(--duration-short) var(--ease-standard);
}

.thread-peer-info:hover {
  background: var(--surface-soft);
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
  color: var(--muted);
  font-weight: 500;
  line-height: 1.25;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.thread-status.online {
  color: #22c55e;
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
  min-width: 36px;
  min-height: 36px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  color: var(--chat-accent);
  transition: background var(--duration-short) var(--ease-standard), transform var(--duration-short) var(--ease-standard);
}

.thread-actions .icon-btn:hover {
  background: var(--surface-soft);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  border-radius: var(--radius-pill);
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
  z-index: 1;
  background: transparent;
}

.message-thread.has-wallpaper .message-bubble:not(.outgoing) {
  background: #ffffff;
  color: #111827;
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.06);
}

:global([data-theme='dark']) .message-thread.has-wallpaper .message-bubble:not(.outgoing),
[data-theme='dark'] .message-thread.has-wallpaper .message-bubble:not(.outgoing) {
  background: var(--surface-card, #1f2937);
  color: var(--ink, #f9fafb);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
}

.message-thread.has-wallpaper .message-bubble.outgoing {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.22);
}

.message-thread.has-wallpaper .day-separator {
  background: rgba(255, 255, 255, 0.88);
  color: var(--muted);
  border: 1px solid rgba(0, 0, 0, 0.08);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.10);
}

:global([data-theme='dark']) .message-thread.has-wallpaper .day-separator,
[data-theme='dark'] .message-thread.has-wallpaper .day-separator {
  background: rgba(17, 24, 39, 0.85);
  color: var(--muted);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.35);
}

.message-thread.has-wallpaper .jump-latest {
  background: var(--canvas);
  border: 1px solid var(--hairline);
  color: var(--ink);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
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

.message-bubble.pending.has-pending-attachment {
  opacity: 1;
  padding: 4px;
  background: transparent;
  box-shadow: none;
}

.pending-attachment-body {
  padding: var(--space-xs) var(--space-md);
  margin: 0 0 6px 0;
  border-radius: var(--radius-lg);
  background: var(--ink, #3b82f6);
  color: var(--on-ink, #ffffff);
  font-size: 14px;
}

.pending-image-card {
  position: relative;
  width: min(260px, 100%);
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--surface-card, rgba(0, 0, 0, 0.08));
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.12));
  display: block;
}

.pending-image-thumb {
  display: block;
  width: 100%;
  max-height: 320px;
  object-fit: cover;
  border-radius: var(--radius-lg);
}

.pending-image-placeholder {
  width: 220px;
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: var(--surface-soft, rgba(0, 0, 0, 0.06));
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.pending-skeleton-pulse {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  animation: skeletonShimmer 1.5s infinite;
}

@keyframes skeletonShimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.pending-heic-pill {
  position: relative;
  z-index: 1;
  padding: 3px 9px;
  border-radius: var(--radius-pill, 12px);
  background: rgba(0, 0, 0, 0.6);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.pending-image-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(2px);
  -webkit-backdrop-filter: blur(2px);
  color: #ffffff;
  border-radius: var(--radius-lg);
  padding: var(--space-sm);
  text-align: center;
  transition: opacity 0.2s ease;
}

.pending-image-overlay.has-error {
  background: rgba(185, 28, 28, 0.82);
  backdrop-filter: blur(3px);
  -webkit-backdrop-filter: blur(3px);
}

.messenger-progress-ring {
  position: relative;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ring-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-track {
  stroke: rgba(255, 255, 255, 0.28);
}

.ring-fill {
  stroke: #ffffff;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.25s linear;
}

.ring-pct {
  position: absolute;
  font-size: 11px;
  font-weight: 700;
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.6);
}

.pending-bubble-cancel {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.65);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  transition: transform 0.15s ease, background 0.15s ease;
  z-index: 3;
}

.pending-bubble-cancel:hover {
  transform: scale(1.08);
  background: rgba(0, 0, 0, 0.88);
}

.pending-failed-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.failed-badge-icon {
  font-size: 22px;
}

.failed-msg {
  font-size: 12px;
  font-weight: 600;
  color: #ffffff;
  max-width: 200px;
  word-break: break-word;
}

.failed-actions {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}

.bubble-action-btn {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: opacity 0.15s;
}

.bubble-action-btn.retry {
  background: #ffffff;
  color: #111827;
}

.bubble-action-btn.cancel {
  background: rgba(255, 255, 255, 0.25);
  color: #ffffff;
}

.pending-file-card {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) var(--space-sm);
  background: var(--surface-card);
  border-radius: var(--radius-md);
  border: 1px solid var(--hairline);
  min-width: 200px;
}

.pending-file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--ink);
}

.pending-file-meta {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.pending-file-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pending-file-size {
  font-size: 11px;
  color: var(--muted);
}

.pending-file-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.pending-file-pct {
  font-size: 11px;
  font-weight: 700;
  color: var(--ink);
}

.pending-file-cancel-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted);
  padding: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pending-file-cancel-btn:hover {
  color: var(--ink);
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
  min-height: 100%;
  gap: 2px;
  padding: var(--space-md);
}

.message-list > :first-child {
  margin-top: auto;
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

/* Extra space so top decorations don't overlap the message above */
.message-row.has-bubble-decoration {
  margin-top: 26px;
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
  position: relative;
  max-width: min(76%, 520px);
  padding: 9px 15px;
  border-radius: 18px;
  background: var(--surface-card);
  color: var(--ink);
  font-size: 15px;
  line-height: 1.4;
  overflow-wrap: break-word;
  word-break: normal;
  white-space: pre-wrap;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  cursor: pointer;
  -webkit-touch-callout: none;
}

.message-bubble.outgoing {
  background: var(--chat-bubble-outgoing);
  color: #ffffff;
}

/* Bubble decorations */
.bubble-decoration {
  position: absolute;
  pointer-events: none;
  z-index: 3;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bubble-decoration-img {
  display: block;
  max-height: 40px;
  width: auto;
  object-fit: contain;
  pointer-events: none;
  user-select: none;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.dec-top-left {
  top: -20px;
  left: 6px;
}

.dec-top-right {
  top: -20px;
  right: 6px;
}

.dec-bottom-left {
  bottom: -14px;
  left: 6px;
}

.dec-bottom-right {
  bottom: -14px;
  right: 6px;
}

.dec-left {
  left: -18px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-right {
  right: -18px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-top {
  top: -18px;
  left: 50%;
  transform: translateX(-50%);
}

.dec-bottom {
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
}

.message-bubble.has-dec-left {
  padding-left: 36px;
}

.message-bubble.has-dec-right {
  padding-right: 36px;
}

.message-bubble.has-dec-top {
  padding-top: 12px;
}

.message-bubble.has-dec-bottom {
  padding-bottom: 12px;
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

.message-bubble.has-sticker {
  background: transparent !important;
  box-shadow: none !important;
  padding: 4px 0 !important;
}

.message-bubble.has-sticker .bubble-time {
  display: none;
}

.sticker-bubble {
  font-size: 72px !important;
  line-height: 1.1 !important;
  display: inline-block;
  user-select: none;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.12));
  animation: messenger-pop 0.25s cubic-bezier(0.175, 0.885, 0.32, 1.275);
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
  background: var(--surface-card);
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
  text-align: center;
  border-bottom: 1px solid var(--hairline);
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;
  transition: background var(--duration-short) var(--ease-standard), color var(--duration-short) var(--ease-standard);
  z-index: 5;
}

.connection-status:hover {
  background: var(--surface-soft);
  color: var(--ink);
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
  flex-shrink: 0;
  z-index: 20;
  transition: background var(--duration-short) var(--ease-standard), border-color var(--duration-short) var(--ease-standard);
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
  color: var(--muted);
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition:
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard);
}

.attach-btn:hover {
  color: var(--ink);
  background: var(--surface-soft);
}

.attach-btn:focus-visible {
  outline: 2px solid var(--chat-accent);
  outline-offset: 2px;
}

.sticker-toggle-btn {
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
  color: var(--muted);
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition:
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard),
    transform var(--duration-short) var(--ease-standard);
}

.sticker-toggle-btn:hover {
  color: var(--ink);
  background: var(--surface-soft);
}

.sticker-toggle-btn.active {
  background: color-mix(in srgb, var(--chat-accent) 18%, transparent);
  color: var(--chat-accent);
}

.sticker-picker-drawer {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border-top: 1px solid var(--hairline);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.08);
  height: 280px;
  max-height: 40vh;
  max-height: 40dvh;
  overflow: hidden;
  flex-shrink: 0;
  animation: slideUpSticker 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 10;
}

.sticker-picker-drawer :deep(.emoji-picker) {
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}

.sticker-picker-drawer :deep(.picker-grid) {
  flex: 1;
  min-height: 0;
  max-height: none;
}

@keyframes slideUpSticker {
  from {
    transform: translateY(16px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

img.avatar-img {
  object-fit: cover;
  border-radius: 50%;
  display: block;
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
  outline: 2px solid var(--chat-accent);
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
  color: var(--chat-accent);
  cursor: pointer;
  flex-shrink: 0;
  margin-bottom: 2px;
  transition:
    transform var(--duration-short) var(--ease-standard),
    background var(--duration-short) var(--ease-standard);
}

.like-btn:hover {
  transform: scale(1.15);
  background: color-mix(in srgb, var(--chat-accent) 10%, transparent);
}

.like-btn:active {
  transform: scale(0.92);
}

.like-btn:focus-visible {
  outline: 2px solid var(--chat-accent);
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
  background: var(--chat-bubble-outgoing);
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
    grid-template-columns: var(--rail-w, 320px) minmax(0, 1fr);
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
    grid-template-columns: var(--rail-w, 340px) minmax(0, 1fr);
  }

  .chat-app:has(.chat-info-sidebar) {
    grid-template-columns: var(--rail-w, 340px) minmax(0, 1fr) var(--info-w, 340px);
  }

  .media-panel {
    display: none !important;
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
    grid-template-columns: var(--rail-w, 320px) minmax(0, 1fr);
  }

  .chat-app:has(.chat-info-sidebar) {
    grid-template-columns: var(--rail-w, 300px) minmax(0, 1fr) var(--info-w, 320px);
  }

  .media-panel {
    display: none;
  }
}

/* Draggable resizer splitter for desktop */
.chat-resizer {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 12px;
  cursor: col-resize;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: center;
  touch-action: none;
  user-select: none;
}

.chat-resizer.right-edge {
  right: -6px;
}

.chat-resizer.left-edge {
  left: -6px;
}

.chat-resizer .resizer-handle-line {
  width: 2px;
  height: 100%;
  background: transparent;
  border-radius: 1px;
  transition: background 0.15s ease, width 0.15s ease;
}

.chat-resizer:hover .resizer-handle-line,
.chat-resizer.active .resizer-handle-line,
.chat-resizer:focus-visible .resizer-handle-line {
  width: 4px;
  background: var(--accent);
}

.chat-resizer:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

.chat-app.resizing-col {
  user-select: none !important;
  cursor: col-resize !important;
}

.chat-app.resizing-col * {
  user-select: none !important;
  cursor: col-resize !important;
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

  .sticker-picker-drawer {
    height: 250px;
    max-height: 38dvh;
  }
}

.clickable {
  cursor: pointer;
  border-radius: var(--radius-md);
  padding: 2px 6px;
  margin: -2px -6px;
  transition: background var(--duration-short) var(--ease-standard);
}

.clickable:hover {
  background: var(--surface-soft);
}

/* Messenger-grade Thread Empty State */
.thread-empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl) var(--space-md);
  text-align: center;
  gap: var(--space-xs);
}

.thread-empty-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: var(--radius-pill);
  font-size: 28px;
  font-weight: 700;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  margin-bottom: var(--space-xs);
}

.thread-empty-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--ink);
}

.thread-empty-subtitle {
  margin: 0 0 var(--space-md);
  font-size: 13px;
  color: var(--muted);
}

.thread-empty-wave-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 600;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  color: var(--accent);
  border: 1px solid transparent;
  cursor: pointer;
  transition: transform var(--duration-short) var(--ease-standard), background var(--duration-short) var(--ease-standard);
}

.thread-empty-wave-btn:hover {
  transform: scale(1.04);
  background: var(--accent);
  color: var(--on-accent);
}

/* Action Sheet Styles */
.sheet-action-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  padding: var(--space-xs) 0;
}

.sheet-message-preview {
  padding: var(--space-xs) var(--space-sm) var(--space-sm);
  border-bottom: 1px solid var(--hairline);
  margin-bottom: var(--space-xs);
  color: var(--muted);
  font-size: 13px;
  word-break: break-word;
  max-height: 80px;
  overflow-y: auto;
}

.sheet-action-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: 12px var(--space-sm);
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  color: var(--ink);
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  text-align: left;
  transition: background var(--duration-short) var(--ease-standard);
}

.sheet-action-item:hover {
  background: var(--surface-soft);
}

.sheet-action-item.danger-text {
  color: #ef4444;
}

/* Chat Info Sheet */
.chat-info-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding-bottom: var(--space-md);
}

.chat-info-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 4px;
  padding: var(--space-sm) 0;
}

.chat-info-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: var(--radius-pill);
  font-size: 28px;
  font-weight: 700;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  margin-bottom: var(--space-xxs);
}

.chat-info-name {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--ink);
}

.chat-info-status {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
}

.chat-info-status.online {
  color: #22c55e;
}

.chat-info-actions {
  display: flex;
  justify-content: center;
  gap: var(--space-md);
  padding: var(--space-xs) 0;
}

.chat-info-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 12px;
  color: var(--ink);
  font-weight: 500;
}

.action-icon-circle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-pill);
  background: var(--surface-soft);
  color: var(--ink);
  transition: background var(--duration-short) var(--ease-standard), transform var(--duration-short) var(--ease-standard);
}

.chat-info-action-btn:hover .action-icon-circle {
  background: var(--accent-soft);
  color: var(--accent);
  transform: scale(1.05);
}

.theme-picker-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--hairline);
}

.theme-picker-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.theme-picker-header .menu-item-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink);
  white-space: nowrap;
}

.theme-dots {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding-left: 36px;
  max-width: 100%;
}

.theme-dot {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-pill);
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  transition: transform var(--duration-short) var(--ease-standard);
}

.theme-dot:hover {
  transform: scale(1.15);
}

.theme-dot.active {
  border-color: var(--ink);
  transform: scale(1.15);
  box-shadow: 0 0 0 2px var(--canvas);
}

.chat-info-media-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 3px;
  padding: 4px var(--space-sm);
}

.chat-info-media-thumb {
  aspect-ratio: 1;
  border-radius: 4px;
  overflow: hidden;
  border: none;
  padding: 0;
  background: var(--surface-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.chat-info-media-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity 0.15s ease;
}

.chat-info-media-thumb:hover img {
  opacity: 0.88;
}

/* Seen indicator */
.seen-indicator {
  display: inline-flex;
  align-self: flex-end;
  align-items: center;
  gap: 5px;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
  border-radius: var(--radius-pill);
  margin-top: 2px;
  margin-bottom: 2px;
  user-select: none;
  transition: all var(--duration-short) var(--ease-standard);
}

.seen-avatar {
  width: 14px;
  height: 14px;
  border-radius: var(--radius-pill);
  font-size: 8px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 0 0 1px var(--canvas);
}

.seen-text {
  font-size: 11px;
  font-weight: 500;
  color: inherit;
  white-space: nowrap;
}

/* Adaptive Seen Indicator when wallpaper is active */
.message-thread.has-wallpaper .seen-indicator {
  background: color-mix(in srgb, var(--chat-accent) 8%, rgba(255, 255, 255, 0.92));
  border: 1px solid color-mix(in srgb, var(--chat-accent) 20%, rgba(0, 0, 0, 0.08));
  color: color-mix(in srgb, var(--chat-accent) 85%, #000000);
  font-weight: 600;
  padding: 2px 8px 2px 5px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

:global([data-theme='dark']) .message-thread.has-wallpaper .seen-indicator,
[data-theme='dark'] .message-thread.has-wallpaper .seen-indicator {
  background: color-mix(in srgb, var(--chat-accent, #38bdf8) 12%, rgba(17, 24, 39, 0.90));
  border: 1px solid color-mix(in srgb, var(--chat-accent, #38bdf8) 25%, rgba(255, 255, 255, 0.12));
  color: color-mix(in srgb, var(--chat-accent, #38bdf8) 60%, #ffffff);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.40);
}

.message-thread.has-wallpaper .seen-avatar {
  box-shadow: 0 0 0 1.5px rgba(255, 255, 255, 0.45);
}

:global([data-theme='dark']) .message-thread.has-wallpaper .seen-avatar,
[data-theme='dark'] .message-thread.has-wallpaper .seen-avatar {
  box-shadow: 0 0 0 1.5px rgba(255, 255, 255, 0.25);
}

/* Typing indicator in thread */
.typing-row {
  margin-top: 4px;
}

.typing-bubble {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 10px 14px;
  border-radius: 18px 18px 18px 4px;
  background: var(--surface-card);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  width: fit-content;
}

.typing-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--muted);
  animation: typingBounce 1.4s infinite ease-in-out both;
}

.typing-dot:nth-child(1) { animation-delay: -0.32s; }
.typing-dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes typingBounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

/* Peek preview BottomSheet */
.peek-preview-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding-bottom: var(--space-md);
}

.peek-preview-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) 0;
  border-bottom: 1px solid var(--hairline);
}

.peek-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: var(--radius-pill);
  font-weight: 600;
  font-size: 15px;
}

.peek-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.peek-meta h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
}

.peek-badge {
  font-size: 11px;
  color: #10b981;
  font-weight: 500;
}

.peek-loading,
.peek-empty {
  padding: var(--space-lg) var(--space-sm);
  text-align: center;
  color: var(--muted);
  font-size: 13px;
}

.peek-messages-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 260px;
  overflow-y: auto;
  padding: var(--space-xs) 0;
}

.peek-message-row {
  display: flex;
  width: 100%;
}

.peek-message-row.outgoing {
  justify-content: flex-end;
}

.peek-bubble {
  max-width: 80%;
  padding: 8px 12px;
  border-radius: 14px;
  background: var(--surface-card);
  font-size: 14px;
  line-height: 1.4;
  color: var(--ink);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.peek-bubble.outgoing {
  background: var(--accent-soft);
  color: var(--accent);
}

.peek-time {
  display: block;
  font-size: 10px;
  opacity: 0.6;
  text-align: right;
  margin-top: 2px;
}

.peek-actions {
  padding-top: var(--space-xs);
}

.peek-open-btn {
  width: 100%;
  justify-content: center;
  border-radius: var(--radius-pill);
}

/* --- Messenger Reactions & Double-tap Heart System --- */
.heart-burst {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 48px;
  pointer-events: none;
  z-index: 10;
  animation: heart-burst-pop 0.9s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
}

@keyframes heart-burst-pop {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.2) rotate(-15deg);
  }
  30% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.35) rotate(0deg);
  }
  70% {
    opacity: 1;
    transform: translate(-50%, -70%) scale(1.15);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -100%) scale(0.9);
  }
}

.message-bubble.has-reactions {
  margin-bottom: 12px;
}

.reaction-badge-group {
  position: absolute;
  bottom: -11px;
  right: 8px;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
  background: var(--surface-card);
  border: 1px solid var(--hairline);
  border-radius: 999px;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.12);
  font-size: 13px;
  line-height: 1;
  color: var(--ink);
  cursor: pointer;
  z-index: 2;
  transition: transform 0.15s cubic-bezier(0.34, 1.56, 0.64, 1);
  user-select: none;
}

.reaction-badge-group:hover,
.reaction-badge-group:active {
  transform: scale(1.12);
}

.reaction-badge-group.outgoing-badge {
  right: 8px;
}

.reaction-badge-group.reacted-by-me {
  background: var(--surface);
  border-color: var(--accent);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--accent) 20%, transparent);
}

.rx-emoji {
  display: inline-block;
  line-height: 1;
}

.rx-emoji.my-rx {
  transform: scale(1.05);
}

.rx-count {
  font-size: 11px;
  font-weight: 700;
  margin-left: 1px;
  color: var(--muted);
}

.reacted-by-me .rx-count {
  color: var(--accent);
}

</style>

<!-- Non-scoped: Teleport renders outside component DOM, scoped attrs don't apply -->
<style>
/* Modal Overlay on Long Press */
.reaction-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  animation: rx-fade-in 0.2s ease-out;
}

@keyframes rx-fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.reaction-capsule-wrap {
  position: fixed;
  z-index: 10001;
  animation: rx-capsule-bounce 0.28s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes rx-capsule-bounce {
  from {
    opacity: 0;
    transform: scale(0.6) translateY(16px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.reaction-capsule {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #242526;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.45), 0 2px 8px rgba(0, 0, 0, 0.25);
}

[data-theme='light'] .reaction-capsule,
:root:not([data-theme='dark']) .reaction-capsule {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.rx-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 0;
  transition: transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1), background 0.15s ease;
}

.rx-btn:hover,
.rx-btn:active {
  transform: scale(1.35) translateY(-5px);
  background: rgba(255, 255, 255, 0.12);
}

[data-theme='light'] .rx-btn:hover,
:root:not([data-theme='dark']) .rx-btn:hover {
  background: rgba(0, 0, 0, 0.06);
}

.rx-btn.rx-active {
  background: color-mix(in srgb, var(--accent) 18%, transparent);
  transform: scale(1.18);
}

.rx-char {
  font-size: 26px;
  line-height: 1;
  user-select: none;
}

.rx-plus-btn {
  width: 36px;
  height: 36px;
  background: rgba(255, 255, 255, 0.1);
  color: #e4e6eb;
}

[data-theme='light'] .rx-plus-btn,
:root:not([data-theme='dark']) .rx-plus-btn {
  background: #f0f2f5;
  color: #050505;
}

.reaction-active-bubble-wrap {
  position: fixed;
  z-index: 10000;
  pointer-events: none;
  animation: rx-bubble-pop 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.reaction-modal-overlay > .reaction-active-bubble-wrap:not([style*="top"]) {
  top: calc(50% - 40px);
  left: 50%;
  transform: translate(-50%, -50%);
  width: auto;
  max-width: min(80%, 480px);
}

@keyframes rx-bubble-pop {
  from {
    transform: scale(0.96);
  }
  to {
    transform: scale(1);
  }
}

.message-bubble.active-elevated {
  box-sizing: border-box;
  width: 100%;
  padding: 8px 14px;
  border-radius: 18px 18px 18px 4px;
  background: var(--surface-card);
  color: var(--ink);
  font-size: 15px;
  line-height: 1.36;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.45);
  pointer-events: auto;
  margin: 0;
}

.message-bubble.active-elevated p {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.message-bubble.active-elevated.outgoing {
  border-radius: 18px 18px 4px 18px;
  background: var(--chat-bubble-outgoing);
  color: #ffffff;
}

.message-bubble.active-elevated .bubble-time {
  display: block;
  font-size: 10px;
  color: var(--muted);
  margin-top: 4px;
  opacity: 0.7;
  white-space: nowrap;
}

.message-bubble.active-elevated.outgoing .bubble-time {
  color: rgba(255, 255, 255, 0.75);
}

.message-bubble.active-elevated .sticker-bubble {
  font-size: 56px;
  line-height: 1;
  text-align: center;
}

.message-bubble.active-elevated.has-sticker {
  background: transparent !important;
  box-shadow: none !important;
}

.message-bubble.active-elevated .like-bubble {
  font-size: 56px;
  line-height: 1;
  text-align: center;
}

.message-bubble.active-elevated.has-like {
  background: transparent !important;
  box-shadow: none !important;
}

/* Bottom Actions Bar matching Screenshot */
.reaction-bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  background: #18191a;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 20px 20px 0 0;
  box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.4);
  z-index: 10001;
  animation: rx-slide-up 0.22s cubic-bezier(0.2, 0.9, 0.3, 1);
}

[data-theme='light'] .reaction-bottom-bar,
:root:not([data-theme='dark']) .reaction-bottom-bar {
  background: #ffffff;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.12);
}

@keyframes rx-slide-up {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

.bottom-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: #e4e6eb;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 12px;
  transition: transform 0.15s ease, background 0.15s ease;
  min-width: 60px;
}

[data-theme='light'] .bottom-action-btn,
:root:not([data-theme='dark']) .bottom-action-btn {
  color: #050505;
}

.bottom-action-btn:hover,
.bottom-action-btn:active {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
}

[data-theme='light'] .bottom-action-btn:hover,
:root:not([data-theme='dark']) .bottom-action-btn:hover {
  background: rgba(0, 0, 0, 0.05);
}

.action-icon-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  color: #e4e6eb;
  transition: background 0.15s ease;
}

[data-theme='light'] .action-icon-circle,
:root:not([data-theme='dark']) .action-icon-circle {
  background: #f0f2f5;
  color: #050505;
}

.bottom-action-btn.danger-action {
  color: #f87171;
}

.danger-circle {
  background: rgba(248, 113, 113, 0.15);
  color: #f87171;
}

/* Highlighted message pulse */
.message-row.msg-highlighted .message-bubble {
  animation: pulse-highlight 2.4s ease;
}

@keyframes pulse-highlight {
  0%, 100% {
    box-shadow: 0 0 0 0 transparent;
  }
  20%, 50% {
    box-shadow: 0 0 0 4px color-mix(in srgb, var(--accent) 40%, transparent), 0 4px 16px color-mix(in srgb, var(--accent) 25%, transparent);
    transform: scale(1.02);
  }
}

/* Sub-page Navigation inside Chat Info BottomSheet */
.info-subpage-nav {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 var(--space-sm) var(--space-sm);
  border-bottom: 1px solid var(--hairline);
  margin-bottom: var(--space-xs);
}

.subpage-back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  color: var(--accent);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 8px 4px 2px;
  border-radius: var(--radius-sm);
  transition: background 0.15s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.subpage-back-btn:hover {
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}

.subpage-title {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: center;
}

.subpage-close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  min-width: 28px;
  min-height: 28px;
  border: none;
  border-radius: var(--radius-pill);
  background: var(--surface-soft);
  color: var(--muted);
  cursor: pointer;
  margin-left: auto;
  transition: background 0.15s ease, color 0.15s ease;
}

.subpage-close-btn:hover {
  background: var(--surface-card);
  color: var(--ink);
}

.menu-badge {
  display: inline-block;
  padding: 1px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--muted);
  margin-left: auto;
  margin-right: 4px;
}

.menu-item-arrow {
  color: var(--muted);
  margin-left: auto;
  opacity: 0.6;
}

.menu-badge + .menu-item-arrow {
  margin-left: 0;
}

/* Dedicated Search Sub-page */
.info-search-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  padding: 0 var(--space-sm);
  animation: fadeInDown 0.2s ease-out;
}

.info-search-page-bar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 8px 14px;
  background: var(--surface-soft);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
}

.subpage-status {
  padding: 12px;
  color: var(--muted);
  font-size: 13px;
  text-align: center;
}

.subpage-empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl) var(--space-md);
  text-align: center;
  color: var(--muted);
  gap: 12px;
}

.empty-hint-icon {
  opacity: 0.35;
  color: var(--muted);
}

.subpage-empty-hint p {
  margin: 0;
  font-size: 13px;
  max-width: 240px;
  line-height: 1.4;
}

.subpage-results-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.subpage-results-count {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  padding: 0 4px;
}

/* Dedicated Media Sub-page */
.info-media-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  animation: fadeInDown 0.2s ease-out;
}

/* Dedicated Wallpaper Sub-page */
.info-wallpaper-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: 0 var(--space-xs);
  animation: fadeInDown 0.2s ease-out;
}

.wallpaper-preview-card {
  position: relative;
  height: 150px;
  border-radius: var(--radius-lg, 16px);
  overflow: hidden;
  border: 1px solid var(--hairline);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  background-size: cover;
  background-position: center;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 12px;
}

.wallpaper-preview-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.08);
  pointer-events: none;
}

.wallpaper-preview-bubbles {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.preview-bubble {
  max-width: 80%;
  padding: 6px 12px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.35;
}

.preview-bubble p {
  margin: 0;
}

.preview-bubble.incoming {
  align-self: flex-start;
  background: var(--surface-card);
  color: var(--ink);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
}

.preview-bubble.incoming span {
  display: block;
  font-size: 10px;
  font-weight: 600;
  color: var(--muted);
  margin-bottom: 2px;
}

.preview-bubble.outgoing {
  align-self: flex-end;
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
}

.wallpaper-actions-row {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.wallpaper-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex: 1;
  padding: 8px 12px;
  border-radius: var(--radius-pill);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.wallpaper-action-btn.primary {
  background: var(--accent);
  color: #ffffff;
  border: none;
}

.wallpaper-action-btn.primary:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.wallpaper-action-btn.secondary {
  background: var(--surface-soft);
  color: var(--muted);
  border: 1px solid var(--hairline);
}

.wallpaper-action-btn.secondary:hover {
  color: var(--danger, #ef4444);
  background: rgba(239, 68, 68, 0.08);
}

.wallpaper-presets-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wallpaper-section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--muted);
  padding: 0 2px;
}

.wallpaper-mode-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--surface-soft);
  padding: 3px;
  border-radius: var(--radius-md, 8px);
  margin-bottom: 4px;
}

.wallpaper-mode-tab {
  flex: 1;
  border: none;
  background: transparent;
  padding: 6px 4px;
  border-radius: var(--radius-sm, 6px);
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
  cursor: pointer;
  transition: all var(--duration-short) var(--ease-standard);
  white-space: nowrap;
  text-align: center;
}

.wallpaper-mode-tab:hover {
  color: var(--ink);
}

.wallpaper-mode-tab.active {
  background: var(--canvas);
  color: var(--ink);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.wallpaper-presets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(88px, 1fr));
  gap: 12px;
}

.wallpaper-preset-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  outline: none;
}

.wallpaper-preset-thumb {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  border-radius: 12px;
  border: 2px solid transparent;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  transition: transform 0.15s ease, border-color 0.15s ease;
  overflow: hidden;
}

.wallpaper-preset-item:hover .wallpaper-preset-thumb {
  transform: scale(1.04);
}

.wallpaper-preset-item.active .wallpaper-preset-thumb {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent) 30%, transparent);
}

.preset-check-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--accent);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.wallpaper-preset-name,
.wallpaper-preset-label {
  font-size: 11px;
  color: var(--ink);
  font-weight: 500;
  text-align: center;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.wallpaper-action-btn.storage {
  background: var(--surface-soft);
  color: var(--ink);
  border: 1px solid var(--hairline);
}

.wallpaper-action-btn.storage:hover {
  background: var(--surface-card);
  border-color: var(--accent);
}

.wallpaper-section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 2px;
}

.custom-wallpaper-count {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
}

.custom-thumb {
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.delete-custom-wp-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.65);
  color: #ffffff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  z-index: 2;
}

.delete-custom-wp-btn:hover {
  background: var(--danger, #ef4444);
  transform: scale(1.15);
}

/* Storage & Media Picker BottomSheet */
.storage-picker-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: 0 var(--space-xs) var(--space-md);
  min-height: 320px;
}

.storage-picker-tabs {
  display: flex;
  gap: 6px;
  padding: 4px;
  background: var(--surface-soft);
  border-radius: var(--radius-lg);
  border: 1px solid var(--hairline);
}

.storage-tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--duration-short) var(--ease-standard);
}

.storage-tab-btn:hover {
  color: var(--ink);
}

.storage-tab-btn.active {
  background: var(--canvas);
  color: var(--ink);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.storage-picker-body {
  max-height: 55vh;
  overflow-y: auto;
  padding: 4px 2px;
}

.storage-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
  gap: 8px;
}

.storage-photo-card {
  position: relative;
  aspect-ratio: 1;
  width: 100%;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--hairline);
  padding: 0;
  background: var(--surface-soft);
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
  outline: none;
}

.storage-photo-card:hover {
  transform: scale(1.03);
  border-color: var(--accent);
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.12);
}

.storage-photo-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.storage-photo-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
}

.storage-photo-name {
  position: absolute;
  bottom: 0;
  inset-inline: 0;
  padding: 3px 6px;
  font-size: 10px;
  color: #ffffff;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.75));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
}

.storage-picker-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl) var(--space-md);
  color: var(--muted);
  text-align: center;
  gap: 8px;
}

.storage-picker-empty p {
  margin: 0;
  font-size: 13px;
}

.storage-picker-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl) var(--space-md);
  color: var(--muted);
  gap: 12px;
}

.storage-picker-loading p {
  margin: 0;
  font-size: 13px;
}

/* In-details Message Search */
.info-search-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: 0 var(--space-sm) var(--space-xs);
  animation: fadeInDown 0.2s ease-out;
}

@keyframes fadeInDown {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

.info-search-bar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 6px 12px;
  background: var(--surface-soft);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
}

.info-search-icon {
  color: var(--muted);
  flex-shrink: 0;
}

.info-search-input {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--ink);
  font-size: 13px;
  outline: none;
}

.info-search-clear {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border: none;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  border-radius: 50%;
}

.info-search-status {
  padding: 8px 12px;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.info-search-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 480px;
  overflow-y: auto;
}

.info-search-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
  width: 100%;
}

.info-search-item:hover {
  background: var(--surface-soft);
}

.search-item-avatar {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-soft);
  flex-shrink: 0;
}

.search-item-avatar .avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.search-item-avatar .avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
}

.info-search-item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-search-sender-row {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-search-body-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--muted);
  overflow: hidden;
  white-space: nowrap;
}

.info-search-snippet {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--muted);
}

.info-search-snippet mark.search-match {
  background: transparent;
  color: var(--ink);
  font-weight: 700;
}

.info-search-dot {
  opacity: 0.6;
  flex-shrink: 0;
}

.info-search-time {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--muted);
}

.info-search-body {
  font-size: 12px;
  color: var(--muted);
}

.info-search-results-tag {
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
  background: var(--surface-card);
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  white-space: nowrap;
  flex-shrink: 0;
}

.search-enter-hint {
  font-size: 13px;
  color: var(--muted);
  text-align: center;
  margin: 32px 0;
}

/* Media Tabs Header & Content */
.info-tabs-header {
  display: flex;
  gap: 8px;
  padding: 0 var(--space-sm);
  border-bottom: 1px solid var(--hairline);
  margin-bottom: var(--space-sm);
}

.info-tab-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 8px;
  border: none;
  border-radius: 0;
  background: transparent;
  color: var(--muted);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  position: relative;
  transition: color 0.15s ease;
}

.info-tab-btn:hover {
  background: transparent;
  color: var(--ink);
}

.info-tab-btn.active {
  background: transparent;
  color: var(--accent);
  font-weight: 600;
  box-shadow: inset 0 -2px 0 var(--accent);
}

.tab-badge {
  display: inline-block;
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 600;
  border-radius: var(--radius-pill);
  background: var(--surface-card);
  color: var(--muted);
}

.tab-empty-hint {
  margin: 0;
  padding: var(--space-md);
  text-align: center;
  font-size: 13px;
  color: var(--muted);
}

/* Shared file items in chat info */
.chat-info-files-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 var(--space-sm);
}

.shared-file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
  text-align: left;
  transition: background 0.15s ease;
}

.shared-file-item:hover {
  background: var(--surface-soft);
}

.file-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 8px;
  background: var(--surface-soft);
  color: var(--ink);
  flex-shrink: 0;
}

.file-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-sub {
  font-size: 11px;
  color: var(--muted);
}

.file-action-icon {
  color: var(--muted);
  flex-shrink: 0;
}

/* Shared link cards in chat info */
.chat-info-links-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 var(--space-sm);
}

.shared-link-card {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  overflow: hidden;
}

.shared-link-main {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s ease;
}

.shared-link-main:hover {
  background: color-mix(in srgb, var(--accent) 6%, transparent);
}

.link-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--surface-card);
  color: var(--accent);
  flex-shrink: 0;
}

.link-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.link-domain {
  font-size: 12px;
  font-weight: 600;
  color: var(--ink);
}

.link-url {
  font-size: 11px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-external {
  color: var(--muted);
  flex-shrink: 0;
}

.link-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 10px 6px;
  font-size: 11px;
  border-top: 1px solid var(--hairline);
}

.link-meta {
  color: var(--muted);
}

.link-jump-btn {
  border: none;
  background: transparent;
  color: var(--accent);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  padding: 2px 4px;
}

.link-jump-btn:hover {
  text-decoration: underline;
}

/* Desktop Media Panel tabs and lists */
.media-panel-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: var(--space-xs);
  border-bottom: 1px solid var(--hairline);
  padding-bottom: 6px;
}

.panel-tab-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 6px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--muted);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.panel-tab-btn:hover {
  background: var(--surface-soft);
  color: var(--ink);
}

.panel-tab-btn.active {
  background: var(--surface-soft);
  color: var(--accent);
  font-weight: 600;
}

.panel-tab-count {
  font-size: 10px;
  opacity: 0.8;
}

.panel-tab-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.panel-media-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.panel-files-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.panel-file-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--ink);
  cursor: pointer;
  text-align: left;
  transition: background 0.15s ease;
}

.panel-file-row:hover {
  background: var(--surface-soft);
}

.file-row-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.file-row-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.file-row-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-row-meta {
  font-size: 10px;
  color: var(--muted);
}

.file-row-dl {
  color: var(--muted);
  flex-shrink: 0;
}

.panel-links-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.panel-link-card {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  overflow: hidden;
}

.panel-link-anchor {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  text-decoration: none;
  color: inherit;
  font-size: 12px;
}

.panel-link-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.panel-link-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.panel-link-domain {
  font-weight: 600;
  color: var(--ink);
  font-size: 11px;
}

.panel-link-raw {
  color: var(--muted);
  font-size: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-link-ext {
  color: var(--muted);
  flex-shrink: 0;
}

.panel-link-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 3px 8px 4px;
  font-size: 10px;
  border-top: 1px solid var(--hairline);
}

.panel-link-sender {
  color: var(--muted);
}

.panel-link-jump {
  border: none;
  background: transparent;
  color: var(--accent);
  font-size: 10px;
  cursor: pointer;
  padding: 0;
}

.panel-link-jump:hover {
  text-decoration: underline;
}

/* Active state for info icon button on thread header */
.thread-actions .icon-btn.active {
  background: var(--accent-soft, color-mix(in srgb, var(--accent) 12%, transparent));
  color: var(--chat-accent);
}

/* Desktop Chat Info Sidebar */
.chat-info-sidebar {
  display: none;
}

@media (min-width: 1024px) {
  .chat-info-sidebar {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    min-width: 0;
    background: var(--canvas);
    border-left: 1px solid var(--hairline);
    overflow-y: auto;
    overflow-x: hidden;
  }
}

.desktop-info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  min-height: 60px;
  box-sizing: border-box;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
  flex-shrink: 0;
}

.desktop-info-header-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--ink);
  letter-spacing: -0.01em;
}

.desktop-info-close-btn {
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  border-radius: var(--radius-pill);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  transition: background var(--duration-short) var(--ease-standard), color var(--duration-short) var(--ease-standard);
}

.desktop-info-close-btn:hover {
  background: var(--surface-soft);
  color: var(--ink);
}

.chat-info-sidebar .chat-info-content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  padding: var(--space-sm) var(--space-md) var(--space-lg);
}

.chat-info-sidebar .info-subpage-nav {
  height: 60px;
  min-height: 60px;
  box-sizing: border-box;
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  border-bottom: 1px solid var(--hairline);
  margin-bottom: var(--space-sm);
  flex-shrink: 0;
}
</style>
