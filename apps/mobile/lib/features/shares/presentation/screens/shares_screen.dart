import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../controllers/shares_controller.dart';
import '../../data/models/share_model.dart';

class SharesScreen extends ConsumerStatefulWidget {
  const SharesScreen({super.key});

  @override
  ConsumerState<SharesScreen> createState() => _SharesScreenState();
}

class _SharesScreenState extends ConsumerState<SharesScreen>
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
    final state = ref.watch(sharesControllerProvider);
    final ctrl = ref.read(sharesControllerProvider.notifier);
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Chia sẻ', style: TextStyle(fontWeight: FontWeight.w700)),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppColors.accent,
          labelColor: AppColors.accent,
          unselectedLabelColor: theme.colorScheme.onSurfaceVariant,
          tabs: [
            Tab(text: 'Được chia sẻ (${state.incoming.length})'),
            Tab(text: 'Tôi đã chia sẻ (${state.outgoing.length})'),
          ],
        ),
      ),
      body: RefreshIndicator(
        onRefresh: () => ctrl.loadShares(),
        child: state.isLoading && state.incoming.isEmpty && state.outgoing.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : TabBarView(
                controller: _tabController,
                children: [
                  // Tab 1: Incoming Shares
                  state.incoming.isEmpty
                      ? _buildEmptyState('Chưa có ai chia sẻ mục nào với bạn')
                      : ListView.separated(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          itemCount: state.incoming.length,
                          separatorBuilder: (context, index) =>
                              const Divider(height: 1, indent: 68),
                          itemBuilder: (context, index) {
                            final share = state.incoming[index];
                            return _buildIncomingTile(context, share);
                          },
                        ),

                  // Tab 2: Outgoing Shares
                  state.outgoing.isEmpty
                      ? _buildEmptyState('Bạn chưa chia sẻ mục nào')
                      : ListView.separated(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          itemCount: state.outgoing.length,
                          separatorBuilder: (context, index) =>
                              const Divider(height: 1, indent: 68),
                          itemBuilder: (context, index) {
                            final share = state.outgoing[index];
                            return _buildOutgoingTile(context, share, ctrl);
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
                child: const Icon(Icons.people_alt_outlined, size: 40, color: AppColors.accent),
              ),
              const SizedBox(height: 16),
              Text(
                message,
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 6),
              Text(
                'Các tệp và thư mục được chia sẻ nội bộ sẽ hiển thị tại đây',
                style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildIncomingTile(BuildContext context, IncomingShareModel share) {
    return ListTile(
      leading: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          color: (share.isFolder ? AppColors.accent : Colors.blue).withOpacity(0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(
          share.isFolder ? Icons.folder_shared_rounded : Icons.insert_drive_file_rounded,
          color: share.isFolder ? AppColors.accent : Colors.blue,
          size: 24,
        ),
      ),
      title: Text(
        share.resourceName,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
      ),
      subtitle: Text(
        'Từ: ${share.owner.displayName} (${share.owner.email})',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: const Icon(Icons.chevron_right_rounded),
      onTap: () {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Đang mở "${share.resourceName}"...')),
        );
      },
    );
  }

  Widget _buildOutgoingTile(
    BuildContext context,
    OutgoingShareModel share,
    SharesController ctrl,
  ) {
    return ListTile(
      leading: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          color: (share.isFolder ? AppColors.accent : Colors.blue).withOpacity(0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(
          share.isFolder ? Icons.folder_shared_rounded : Icons.insert_drive_file_rounded,
          color: share.isFolder ? AppColors.accent : Colors.blue,
          size: 24,
        ),
      ),
      title: Text(
        share.resourceName,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
      ),
      subtitle: Text(
        'Chia sẻ với: ${share.recipient.displayName} (${share.recipient.email})',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: IconButton(
        icon: const Icon(Icons.person_remove_rounded, color: Colors.redAccent),
        tooltip: 'Hủy chia sẻ',
        onPressed: () => _confirmRevoke(context, ctrl, share),
      ),
    );
  }

  void _confirmRevoke(BuildContext context, SharesController ctrl, OutgoingShareModel share) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Hủy chia sẻ?'),
        content: Text('Người dùng ${share.recipient.displayName} sẽ không còn quyền truy cập "${share.resourceName}".'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Giữ lại'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Theme.of(ctx).colorScheme.error),
            onPressed: () async {
              Navigator.pop(ctx);
              final ok = await ctrl.revokeShare(share.id);
              if (ok && context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Đã hủy chia sẻ "${share.resourceName}"')),
                );
              }
            },
            child: const Text('Hủy chia sẻ'),
          ),
        ],
      ),
    );
  }
}
