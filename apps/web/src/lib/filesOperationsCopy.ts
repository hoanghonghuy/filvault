import type { Locale } from '@/lib/i18n'

export interface FilesOperationsCopy {
  renameFileTitle: string
  renameFolderTitle: string
  nameLabel: string
  fileRenamed: string
  folderRenamed: string
  renameFailed: string
  moveFileTitle: string
  moveFolderTitle: string
  moveBatchTitle: (count: number) => string
  fileMoved: string
  folderMoved: string
  movedBatch: (count: number) => string
  moveFailed: string
  trashBatchTitle: (count: number) => string
  trashBatchMessage: string
  trashFileTitle: string
  trashFileMessage: string
  trashFolderTitle: string
  trashFolderMessage: string
  movedToTrash: string
  movedBatchToTrash: (count: number) => string
  trashFailed: string
}

export function filesOperationsCopy(locale: Locale): FilesOperationsCopy {
  if (locale === 'vi') {
    return {
      renameFileTitle: 'Đổi tên tệp',
      renameFolderTitle: 'Đổi tên thư mục',
      nameLabel: 'Tên',
      fileRenamed: 'Đã đổi tên tệp',
      folderRenamed: 'Đã đổi tên thư mục',
      renameFailed: 'Không thể đổi tên',
      moveFileTitle: 'Di chuyển tệp',
      moveFolderTitle: 'Di chuyển thư mục',
      moveBatchTitle: (count) => `Di chuyển ${count} mục`,
      fileMoved: 'Đã di chuyển tệp',
      folderMoved: 'Đã di chuyển thư mục',
      movedBatch: (count) => `Đã di chuyển ${count} mục`,
      moveFailed: 'Không thể di chuyển',
      trashBatchTitle: (count) => `Chuyển ${count} mục vào thùng rác?`,
      trashBatchMessage: 'Bạn có thể khôi phục các mục này từ Thùng rác sau.',
      trashFileTitle: 'Chuyển vào thùng rác?',
      trashFileMessage: 'Bạn có thể khôi phục tệp này từ Thùng rác sau.',
      trashFolderTitle: 'Chuyển thư mục vào thùng rác?',
      trashFolderMessage: 'Thư mục phải trống. Bạn có thể khôi phục thư mục từ Thùng rác sau.',
      movedToTrash: 'Đã chuyển vào thùng rác',
      movedBatchToTrash: (count) => `Đã chuyển ${count} mục vào thùng rác`,
      trashFailed: 'Không thể chuyển vào thùng rác',
    }
  }

  return {
    renameFileTitle: 'Rename file',
    renameFolderTitle: 'Rename folder',
    nameLabel: 'Name',
    fileRenamed: 'File renamed',
    folderRenamed: 'Folder renamed',
    renameFailed: 'Rename failed',
    moveFileTitle: 'Move file',
    moveFolderTitle: 'Move folder',
    moveBatchTitle: (count) => `Move ${count} ${count === 1 ? 'item' : 'items'}`,
    fileMoved: 'File moved',
    folderMoved: 'Folder moved',
    movedBatch: (count) => `Moved ${count} ${count === 1 ? 'item' : 'items'}`,
    moveFailed: 'Move failed',
    trashBatchTitle: (count) => `Move ${count} ${count === 1 ? 'item' : 'items'} to trash?`,
    trashBatchMessage: 'You can restore them from Trash later.',
    trashFileTitle: 'Move to trash?',
    trashFileMessage: 'You can restore this file from Trash later.',
    trashFolderTitle: 'Move folder to trash?',
    trashFolderMessage: 'The folder must be empty. You can restore it from Trash later.',
    movedToTrash: 'Moved to trash',
    movedBatchToTrash: (count) => `Moved ${count} ${count === 1 ? 'item' : 'items'} to trash`,
    trashFailed: 'Could not move to trash',
  }
}
