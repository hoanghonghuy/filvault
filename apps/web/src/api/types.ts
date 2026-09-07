export interface ApiErrorBody {
  error: {
    code: string
    message: string
  }
}

export interface User {
  id: string
  email: string
  displayName: string
  emailVerified: boolean
  storageUsed: number
  storageQuota: number
  imageThumbnailsEnabled: boolean
  videoThumbnailsEnabled: boolean
  trashAutoDeleteEnabled: boolean
  trashRetentionDays: number
  activeStatusEnabled?: boolean
  createdAt: string
}

export interface Session {
  user: User
  accessToken: string
  refreshToken: string
}

export interface Folder {
  id: string
  parentId: string | null
  name: string
  createdAt: string
  updatedAt: string
}

export interface BrowserFile {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
  updatedAt: string
}

export interface FavoriteFile {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
  updatedAt: string
  favoritedAt: string
}

export interface ShareLinkInfo {
  id: string
  fileId: string
  url?: string
  token?: string
  fileName?: string
  mimeType?: string
  sizeBytes?: number
  expiresAt: string | null
  createdAt: string
}

export interface PublicShareMeta {
  name: string
  mimeType: string
  sizeBytes: number
  expiresAt: string | null
}

export type ShareLinkTTL = '1h' | '24h' | '7d'

export interface Browser {
  folder: Folder | null
  breadcrumb: Folder[]
  folders: Folder[]
  files: BrowserFile[]
}

export interface UploadSession {
  fileId: string
  uploadUrl: string
  expiresAt: string
}

export interface DownloadURL {
  downloadUrl: string
  expiresAt: string
}

export interface StorageUsage {
  usedBytes: number
  quotaBytes: number
}

export interface SearchResult {
  folders: Folder[]
  files: BrowserFile[]
}

export interface SearchFilters {
  type: 'all' | 'image' | 'video' | 'document' | 'archive' | 'folder'
  sort: 'relevance' | 'name' | 'date' | 'size'
  order: 'asc' | 'desc'
  folderId?: string
  from?: string
  to?: string
}

export interface TimelineItem {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
  createdAt: string
  thumbnailUrl?: string
}

export interface TimelineGroup {
  date: string
  items: TimelineItem[]
}

export interface Timeline {
  groups: TimelineGroup[]
  nextBefore?: string
}

export type ChatAttachmentAvailability = 'available' | 'trashed' | 'purged'

export interface ChatAttachment {
  id: string
  fileId: string | null
  originalName: string
  name: string
  mimeType: string
  sizeBytes: number
  createdAt: string
  availability: ChatAttachmentAvailability
  thumbnailUrl?: string
}

export interface ChatMessage {
  id: string
  conversationId: string
  body: string
  senderId: string
  clientMessageId?: string
  editedAt?: string
  removedAt?: string
  createdAt: string
  attachments: ChatAttachment[]
}

export interface ChatConversation {
  id: string
  title: string
  type?: 'legacy' | 'direct'
  peer?: {
    id: string
    name: string
    email: string
    lastSeenAt?: string
  }
  peerStatus?: 'online' | 'offline'
  peerLastSeenAt?: string
  createdAt: string
  updatedAt: string
  lastMessageAt?: string
  unreadCount?: number
  lastReadAt?: string
  lastReadMessageId?: string
  peerLastReadAt?: string
  peerLastReadMessageId?: string
  preview?: {
    messageId: string
    body: string
    createdAt: string
    attachments: string[]
  }
}

export interface Album {
  id: string
  name: string
  itemCount: number
  createdAt: string
  updatedAt: string
  /** Effective cover file (pinned or auto-selected newest item). */
  coverFileId?: string
  /** Presigned URL, present only when the cover is an image and prefs allow. */
  coverUrl?: string
}

export interface AlbumDetail extends Album {
  items: TimelineItem[]
}

export interface TrashItem {
  id: string
  name: string
  deletedAt: string
  mimeType?: string
  sizeBytes?: number
}

export interface TrashList {
  folders: TrashItem[]
  files: TrashItem[]
}

export type ActivityEventType =
  | 'file.uploaded'
  | 'file.trashed'
  | 'file.restored'
  | 'file.purged'
  | 'folder.trashed'
  | 'folder.restored'
  | 'share.created'
  | 'share.revoked'
  | 'password.changed'
  | 'settings.changed'

export interface ActivityEvent {
  id: string
  type: ActivityEventType
  targetName: string
  createdAt: string
}

export interface ActivityPage {
  events: ActivityEvent[]
  nextBefore?: string
}

export type ShareResourceType = 'file' | 'folder'

export interface ShareUserRef {
  id: string
  email: string
  displayName: string
}

export interface OutgoingShare {
  id: string
  resourceType: ShareResourceType
  resourceId: string
  resourceName: string
  recipient: ShareUserRef
  createdAt: string
}

export interface IncomingShare {
  id: string
  resourceType: ShareResourceType
  resourceId: string
  resourceName: string
  owner: ShareUserRef
  createdAt: string
}

export interface SharedFolderInfo {
  id: string
  name: string
}

export interface SharedFileInfo {
  id: string
  name: string
  mimeType: string
  sizeBytes: number
}

export interface SharedBrowser {
  folder: SharedFolderInfo | null
  folders: SharedFolderInfo[]
  files: SharedFileInfo[]
}
