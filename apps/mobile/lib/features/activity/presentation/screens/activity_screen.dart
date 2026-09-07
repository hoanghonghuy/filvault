import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../controllers/activity_controller.dart';
import '../../data/models/activity_model.dart';

class ActivityScreen extends ConsumerWidget {
  const ActivityScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(activityControllerProvider);
    final ctrl = ref.read(activityControllerProvider.notifier);
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Nhật ký hoạt động', style: TextStyle(fontWeight: FontWeight.w700)),
      ),
      body: RefreshIndicator(
        onRefresh: () => ctrl.loadActivity(),
        child: state.isLoading && state.events.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : state.events.isEmpty
                ? _buildEmptyState(context)
                : NotificationListener<ScrollNotification>(
                    onNotification: (ScrollNotification scrollInfo) {
                      if (scrollInfo.metrics.pixels >= scrollInfo.metrics.maxScrollExtent - 200) {
                        ctrl.loadMoreActivity();
                      }
                      return false;
                    },
                    child: ListView.separated(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      itemCount: state.events.length + (state.isLoadingMore ? 1 : 0),
                      separatorBuilder: (context, index) => const Divider(height: 1, indent: 68),
                      itemBuilder: (context, index) {
                        if (index == state.events.length) {
                          return const Padding(
                            padding: EdgeInsets.symmetric(vertical: 16),
                            child: Center(child: CircularProgressIndicator()),
                          );
                        }
                        final event = state.events[index];
                        return _buildEventTile(context, event);
                      },
                    ),
                  ),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
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
                child: const Icon(Icons.history_toggle_off_rounded, size: 40, color: AppColors.accent),
              ),
              const SizedBox(height: 16),
              const Text(
                'Chưa có hoạt động nào',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 6),
              Text(
                'Các thao tác tải lên, xóa và chia sẻ sẽ hiển thị tại đây',
                style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildEventTile(BuildContext context, ActivityEventModel event) {
    final meta = _getEventMeta(event.type);

    return ListTile(
      leading: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          color: meta.color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Icon(meta.icon, color: meta.color, size: 22),
      ),
      title: Text(
        meta.title,
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
      ),
      subtitle: event.targetName.isNotEmpty
          ? Text(
              event.targetName,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 12),
            )
          : null,
      trailing: Text(
        _formatEventTime(event.createdAt),
        style: TextStyle(fontSize: 11, color: Theme.of(context).colorScheme.outline),
      ),
    );
  }

  ({String title, IconData icon, Color color}) _getEventMeta(String type) {
    switch (type) {
      case 'file.uploaded':
        return (title: 'Đã tải lên tệp', icon: Icons.upload_file_rounded, color: Colors.green);
      case 'file.trashed':
        return (title: 'Đã bỏ vào thùng rác', icon: Icons.delete_outline_rounded, color: Colors.orange);
      case 'file.restored':
        return (title: 'Đã khôi phục tệp', icon: Icons.restore_rounded, color: Colors.blue);
      case 'file.purged':
        return (title: 'Đã xóa vĩnh viễn tệp', icon: Icons.delete_forever_rounded, color: Colors.red);
      case 'folder.trashed':
        return (title: 'Đã bỏ thư mục vào rác', icon: Icons.folder_delete_rounded, color: Colors.orange);
      case 'folder.restored':
        return (title: 'Đã khôi phục thư mục', icon: Icons.restore_from_trash_rounded, color: Colors.blue);
      case 'share.created':
        return (title: 'Đã chia sẻ', icon: Icons.share_rounded, color: Colors.purple);
      case 'share.revoked':
        return (title: 'Đã hủy chia sẻ', icon: Icons.link_off_rounded, color: Colors.grey);
      case 'password.changed':
        return (title: 'Đã đổi mật khẩu', icon: Icons.lock_reset_rounded, color: Colors.amber.shade800);
      case 'settings.changed':
        return (title: 'Đã cập nhật cài đặt', icon: Icons.settings_rounded, color: Colors.teal);
      default:
        return (title: 'Hoạt động tài khoản', icon: Icons.info_outline_rounded, color: Colors.blueGrey);
    }
  }

  String _formatEventTime(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return 'Vừa xong';
    if (diff.inMinutes < 60) return '${diff.inMinutes} phút trước';
    if (diff.inHours < 24) return '${diff.inHours} giờ trước';
    if (diff.inDays < 7) return '${diff.inDays} ngày trước';
    return '${dt.day}/${dt.month}/${dt.year}';
  }
}
