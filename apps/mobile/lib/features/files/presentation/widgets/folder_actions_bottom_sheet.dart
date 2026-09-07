import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/widgets/app_bottom_sheet.dart';
import '../../data/models/file_item_model.dart';
import '../controllers/files_controller.dart';

class FolderActionsBottomSheet extends ConsumerWidget {
  final FolderModel folder;

  const FolderActionsBottomSheet({super.key, required this.folder});

  static void show(BuildContext context, FolderModel folder) {
    AppBottomSheet.show(
      context: context,
      title: folder.name,
      child: FolderActionsBottomSheet(folder: folder),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final filesCtrl = ref.read(filesControllerProvider.notifier);

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        ListTile(
          leading: const Icon(Icons.drive_file_rename_outline_rounded),
          title: const Text('Đổi tên thư mục'),
          onTap: () {
            Navigator.pop(context);
            _showRenameDialog(context, ref);
          },
        ),
        ListTile(
          leading: Icon(Icons.delete_outline_rounded, color: theme.colorScheme.error),
          title: Text(
            'Xóa thư mục',
            style: TextStyle(color: theme.colorScheme.error),
          ),
          onTap: () {
            Navigator.pop(context);
            _confirmDelete(context, ref);
          },
        ),
      ],
    );
  }

  void _showRenameDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController(text: folder.name);
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Đổi tên thư mục'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Tên thư mục mới'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () {
              final newName = controller.text.trim();
              if (newName.isNotEmpty && newName != folder.name) {
                ref
                    .read(filesControllerProvider.notifier)
                    .renameFolder(folder.id, newName);
              }
              Navigator.pop(ctx);
            },
            child: const Text('Lưu'),
          ),
        ],
      ),
    );
  }

  void _confirmDelete(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Xóa thư mục?'),
        content: Text('Thư mục "${folder.name}" sẽ được xóa.'),
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
              ref.read(filesControllerProvider.notifier).deleteFolder(folder.id);
              Navigator.pop(ctx);
            },
            child: const Text('Xóa'),
          ),
        ],
      ),
    );
  }
}
