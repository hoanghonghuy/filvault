import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/chat_model.dart';
import '../controllers/chat_list_controller.dart';
import 'chat_room_screen.dart';

class ChatListScreen extends ConsumerStatefulWidget {
  const ChatListScreen({super.key});

  @override
  ConsumerState<ChatListScreen> createState() => _ChatListScreenState();
}

class _ChatListScreenState extends ConsumerState<ChatListScreen> {
  @override
  Widget build(BuildContext context) {
    final state = ref.watch(chatListControllerProvider);
    final ctrl = ref.read(chatListControllerProvider.notifier);
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Đoạn chat', style: TextStyle(fontWeight: FontWeight.w700)),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit_square),
            tooltip: 'Tin nhắn mới',
            onPressed: () => _showNewChatDialog(context),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () => ctrl.loadConversations(),
        child: state.isLoading && state.conversations.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : state.conversations.isEmpty
                ? _buildEmptyState(context)
                : ListView.separated(
                    itemCount: state.conversations.length,
                    separatorBuilder: (context, index) => const Divider(height: 1, indent: 76),
                    itemBuilder: (context, index) {
                      final convo = state.conversations[index];
                      return _buildConversationTile(context, convo);
                    },
                  ),
      ),
      floatingActionButton: FloatingActionButton(
        backgroundColor: AppColors.accent,
        foregroundColor: Colors.white,
        onPressed: () => _showNewChatDialog(context),
        child: const Icon(Icons.chat_rounded),
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
                child: const Icon(Icons.chat_bubble_outline_rounded, size: 38, color: AppColors.accent),
              ),
              const SizedBox(height: 16),
              const Text(
                'Chưa có cuộc trò chuyện nào',
                style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 6),
              Text(
                'Nhấn nút tin nhắn mới để bắt đầu trò chuyện',
                style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildConversationTile(BuildContext context, ChatConversationModel convo) {
    final theme = Theme.of(context);
    final hasUnread = convo.unreadCount > 0;

    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      leading: Stack(
        children: [
          CircleAvatar(
            radius: 26,
            backgroundColor: AppColors.accent.withOpacity(0.15),
            child: Text(
              convo.title.isNotEmpty ? convo.title[0].toUpperCase() : 'U',
              style: const TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 18,
                color: AppColors.accent,
              ),
            ),
          ),
          if (convo.isOnline)
            Positioned(
              right: 0,
              bottom: 0,
              child: Container(
                width: 14,
                height: 14,
                decoration: BoxDecoration(
                  color: Colors.green,
                  shape: BoxShape.circle,
                  border: Border.all(color: theme.scaffoldBackgroundColor, width: 2),
                ),
              ),
            ),
        ],
      ),
      title: Text(
        convo.title,
        style: TextStyle(
          fontWeight: hasUnread ? FontWeight.bold : FontWeight.w600,
          fontSize: 15,
        ),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        convo.preview?.body.isNotEmpty == true
            ? convo.preview!.body
            : (convo.preview?.attachments.isNotEmpty == true
                ? '[Tệp đính kèm]'
                : 'Bắt đầu cuộc trò chuyện'),
        style: TextStyle(
          color: hasUnread ? theme.colorScheme.onSurface : theme.colorScheme.outline,
          fontWeight: hasUnread ? FontWeight.w500 : FontWeight.normal,
          fontSize: 13,
        ),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      trailing: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (convo.lastMessageAt != null)
            Text(
              _formatTime(convo.lastMessageAt!),
              style: TextStyle(
                fontSize: 11,
                color: hasUnread ? AppColors.accent : theme.colorScheme.outline,
                fontWeight: hasUnread ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          const SizedBox(height: 4),
          if (hasUnread)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
              decoration: BoxDecoration(
                color: AppColors.accent,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                convo.unreadCount > 99 ? '99+' : '${convo.unreadCount}',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
        ],
      ),
      onTap: () {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => ChatRoomScreen(conversation: convo),
          ),
        ).then((_) {
          ref.read(chatListControllerProvider.notifier).loadConversations();
        });
      },
    );
  }

  String _formatTime(DateTime dt) {
    final now = DateTime.now();
    if (now.year == dt.year && now.month == dt.month && now.day == dt.day) {
      return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
    }
    return '${dt.day}/${dt.month}';
  }

  void _showNewChatDialog(BuildContext context) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Bắt đầu trò chuyện'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(
            labelText: 'Mã người dùng (User ID)',
            hintText: 'Nhập ULID của người dùng',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () async {
              final targetId = controller.text.trim();
              if (targetId.isNotEmpty) {
                Navigator.pop(ctx);
                final convo = await ref
                    .read(chatListControllerProvider.notifier)
                    .startDirectChat(targetId);
                if (convo != null && context.mounted) {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => ChatRoomScreen(conversation: convo),
                    ),
                  );
                }
              }
            },
            child: const Text('Mở chat'),
          ),
        ],
      ),
    );
  }
}
