import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../files/presentation/widgets/file_icon_helper.dart';
import '../controllers/trash_controller.dart';
import '../../data/models/trash_model.dart';

class TrashScreen extends ConsumerStatefulWidget {
  const TrashScreen({super.key});

  @override
  ConsumerState<TrashScreen> createState() => _TrashScreenState();
}

class _TrashScreenState extends ConsumerState<TrashScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(trashControllerProvider);
    final ctrl = ref.read(trashControllerProvider.notifier);
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Thùng rác', style: TextStyle(fontWeight: FontWeight.w700)),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppColors.accent,
          labelColor: AppColors.accent,
          unselectedLabelColor: theme.colorScheme.onSurfaceVariant,
          tabs: [
            Tab(text: 'Tệp tin (${state.files.length})'),
            Tab(text: 'Thư mục (${state.folders.length})'),
          ],
        ),
      ),
      body: RefreshIndicator(
        onRefresh: () => ctrl.loadTrash(),
        child: state.isLoading && state.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : TabBarView(
                controller: _tabController,
                children: [
                  // Tab 1: Files
                  state.files.isEmpty
                      ? _buildEmptyState('Không có tệp tin nào trong thùng rác')
                      : ListView.separated(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          itemCount: state.files.length,
                          separatorBuilder: (context, index) =>
                              const Divider(height: 1, indent: 68),
                          itemBuilder: (context, index) {
                            final file = state.files[index];
                            return _buildFileItem(context, file, ctrl);
                          },
                        ),

                  // Tab 2: Folders
                  state.folders.isEmpty
                      ? _buildEmptyState('Không có thư mục nào trong thùng rác')
                      : ListView.separated(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          itemCount: state.folders.length,
                          separatorBuilder: (context, index) =>
                              const Divider(height: 1, indent: 68),
                          itemBuilder: (context, index) {
                            final folder = state.folders[index];
                            return _buildFolderItem(context, folder, ctrl);
                          },
                        ),
                ],
              ),
      ),
    );
  }

  Widget _buildEmptyState(String message) {
    return ListView(
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.25),
        Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                width: 76,
                height: 76,
                decoration: BoxDecoration(
                  color: AppColors.accent.withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.delete_sweep_rounded, size: 40, color: AppColors.accent),
              ),
              const SizedBox(height: 16),
              Text(
                message,
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 6),
              Text(
                'Các mục đã xóa sẽ được lưu trữ an toàn tại đây',
                style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildFileItem(BuildContext context, TrashItemModel file, TrashController ctrl) {
    final color = FileIconHelper.getColor(file.mimeType ?? '', context);
    final icon = FileIconHelper.getIcon(file.mimeType ?? '');

    return ListTile(
      leading: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(icon, color: color, size: 24),
      ),
      title: Text(
        file.name,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
      ),
      subtitle: Text(
        '${FileIconHelper.formatBytes(file.sizeBytes)} • Xóa ngày ${_formatDate(file.deletedAt)}',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.restore_rounded, color: AppColors.accent),
            tooltip: 'Khôi phục',
            onPressed: () async {
              final ok = await ctrl.restoreFile(file.id);
              if (ok && context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Đã khôi phục "${file.name}"')),
                );
              }
            },
          ),
          IconButton(
            icon: Icon(Icons.delete_forever_rounded, color: Theme.of(context).colorScheme.error),
            tooltip: 'Xóa vĩnh viễn',
            onPressed: () => _confirmPermanentDelete(context, ctrl, 'files', file.id, file.name),
          ),
        ],
      ),
    );
  }

  Widget _buildFolderItem(BuildContext context, TrashItemModel folder, TrashController ctrl) {
    return ListTile(
      leading: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          color: AppColors.accent.withOpacity(0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: const Icon(Icons.folder_rounded, color: AppColors.accent, size: 24),
      ),
      title: Text(
        folder.name,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
      ),
      subtitle: Text(
        'Xóa ngày ${_formatDate(folder.deletedAt)}',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.restore_rounded, color: AppColors.accent),
            tooltip: 'Khôi phục',
            onPressed: () async {
              final ok = await ctrl.restoreFolder(folder.id);
              if (ok && context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Đã khôi phục thư mục "${folder.name}"')),
                );
              }
            },
          ),
          IconButton(
            icon: Icon(Icons.delete_forever_rounded, color: Theme.of(context).colorScheme.error),
            tooltip: 'Xóa vĩnh viễn',
            onPressed: () => _confirmPermanentDelete(context, ctrl, 'folders', folder.id, folder.name),
          ),
        ],
      ),
    );
  }

  void _confirmPermanentDelete(
    BuildContext context,
    TrashController ctrl,
    String type,
    String id,
    String name,
  ) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Xóa vĩnh viễn?'),
        content: Text('Mục "$name" sẽ bị xóa vĩnh viễn và không thể khôi phục lại.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(ctx).colorScheme.error,
            ),
            onPressed: () async {
              Navigator.pop(ctx);
              final ok = await ctrl.permanentDelete(type: type, id: id);
              if (ok && context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Đã xóa vĩnh viễn "$name"')),
                );
              }
            },
            child: const Text('Xóa vĩnh viễn'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime dt) {
    return '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year}';
  }
}
