import type { Locale } from '@/lib/i18n'

export type FolderPickerCopy = {
  browseFolders: string
  root: string
  parentFolder: string
  loadFailed: string
  empty: string
}

const copy: Record<Locale, FolderPickerCopy> = {
  vi: {
    browseFolders: 'Duyệt thư mục',
    root: 'Gốc',
    parentFolder: '.. Thư mục cha',
    loadFailed: 'Không thể tải thư mục',
    empty: 'Không có thư mục con.',
  },
  en: {
    browseFolders: 'Browse folders',
    root: 'Root',
    parentFolder: '.. Parent folder',
    loadFailed: 'Failed to load folders',
    empty: 'No subfolders here.',
  },
}

export function folderPickerCopy(locale: Locale): FolderPickerCopy {
  return copy[locale]
}
