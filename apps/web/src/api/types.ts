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

export interface Album {
  id: string
  name: string
  itemCount: number
  createdAt: string
  updatedAt: string
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
