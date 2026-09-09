import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/widgets/app_bottom_sheet.dart';
import '../../data/models/file_item_model.dart';
import '../controllers/files_controller.dart';
import 'file_icon_helper.dart';

class FileActionsBottomSheet extends ConsumerWidget {
  final FileItemModel file;

  const FileActionsBottomSheet({super.key, required this.file});

  static void show(BuildContext context, FileItemModel file) {
    AppBottomSheet.show(
      context: context,
      title: file.name,
      child: FileActionsBottomSheet(file: file),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final filesCtrl = ref.read(filesControllerProvider.notifier);

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // File meta preview
        Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: FileIconHelper.getColor(file.mimeType, context)
                      .withOpacity(0.12),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  FileIconHelper.getIcon(file.mimeType),
                  color: FileIconHelper.getColor(file.mimeType, context),
                  size: 26,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      file.name,
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${FileIconHelper.formatBytes(file.sizeBytes)} • ${file.mimeType}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const Divider(height: 1),

        // Action: Download
        ListTile(
          leading: const Icon(Icons.download_rounded),
          title: const Text('Tải xuống / Lấy liên kết'),
          onTap: () async {
            Navigator.pop(context);
            final url = await filesCtrl.getDownloadUrl(file.id);
            if (url != null && context.mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('Đã tạo liên kết tải: $url'),
                  behavior: SnackBarBehavior.floating,
                ),
              );
            }
          },
        ),

        // Action: Toggle Favorite
        ListTile(
          leading: Icon(
            file.isFavorite ? Icons.star_rounded : Icons.star_outline_rounded,
            color: file.isFavorite ? Colors.amber : null,
          ),
          title: Text(file.isFavorite ? 'Bỏ yêu thích' : 'Thêm vào yêu thích'),
          onTap: () {
            Navigator.pop(context);
            filesCtrl.toggleFavorite(file.id, !file.isFavorite);
          },
        ),

        // Action: Rename
        ListTile(
          leading: const Icon(Icons.edit_rounded),
          title: const Text('Đổi tên tệp'),
          onTap: () {
            Navigator.pop(context);
            _showRenameDialog(context, ref, file);
          },
        ),

        // Action: Delete
        ListTile(
          leading: Icon(Icons.delete_outline_rounded, color: theme.colorScheme.error),
          title: Text(
            'Chuyển vào thùng rác',
            style: TextStyle(color: theme.colorScheme.error),
          ),
          onTap: () {
            Navigator.pop(context);
            _confirmDelete(context, ref, file);
          },
        ),
      ],
    );
  }

  void _showRenameDialog(
      BuildContext context, WidgetRef ref, FileItemModel file) {
    final controller = TextEditingController(text: file.name);
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Đổi tên tệp'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Tên mới'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () {
              final newName = controller.text.trim();
              if (newName.isNotEmpty && newName != file.name) {
                ref
                    .read(filesControllerProvider.notifier)
                    .renameFile(file.id, newName);
              }
              Navigator.pop(ctx);
            },
            child: const Text('Lưu'),
          ),
        ],
      ),
    );
  }

  void _confirmDelete(
      BuildContext context, WidgetRef ref, FileItemModel file) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Chuyển vào thùng rác?'),
        content: Text('Tệp "${file.name}" sẽ được chuyển vào thùng rác.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(ctx).colorScheme.error,
            ),
            onPressed: () {
              ref.read(filesControllerProvider.notifier).deleteFile(file.id);
              Navigator.pop(ctx);
            },
            child: const Text('Xóa'),
          ),
        ],
      ),
    );
  }
}
