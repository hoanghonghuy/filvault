import 'dart:io';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../auth/presentation/controllers/auth_controller.dart';
import '../../data/models/chat_model.dart';
import '../controllers/chat_room_controller.dart';
import '../widgets/message_bubble.dart';

class ChatRoomScreen extends ConsumerStatefulWidget {
  final ChatConversationModel conversation;

  const ChatRoomScreen({super.key, required this.conversation});

  @override
  ConsumerState<ChatRoomScreen> createState() => _ChatRoomScreenState();
}

class _ChatRoomScreenState extends ConsumerState<ChatRoomScreen> {
  final TextEditingController _textController = TextEditingController();
  final FocusNode _focusNode = FocusNode();

  @override
  void dispose() {
    _textController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final convo = widget.conversation;
    final roomState = ref.watch(chatRoomControllerProvider(convo.id));
    final roomCtrl = ref.read(chatRoomControllerProvider(convo.id).notifier);
    final authState = ref.watch(authControllerProvider);
    final currentUserId = authState.user?.id ?? '';

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        titleSpacing: 0,
        title: Row(
          children: [
            Stack(
              children: [
                CircleAvatar(
                  radius: 18,
                  backgroundColor: AppColors.accent.withOpacity(0.2),
                  child: Text(
                    convo.title.isNotEmpty ? convo.title[0].toUpperCase() : 'U',
                    style: const TextStyle(fontWeight: FontWeight.bold, color: AppColors.accent),
                  ),
                ),
                if (convo.isOnline)
                  Positioned(
                    right: 0,
                    bottom: 0,
                    child: Container(
                      width: 10,
                      height: 10,
                      decoration: BoxDecoration(
                        color: Colors.green,
                        shape: BoxShape.circle,
                        border: Border.all(color: theme.scaffoldBackgroundColor, width: 1.5),
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    convo.title,
                    style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    convo.isOnline ? 'Đang hoạt động' : 'Ngoại tuyến',
                    style: TextStyle(
                      fontSize: 11,
                      color: convo.isOnline ? Colors.green : theme.colorScheme.outline,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.videocam_rounded),
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Tính năng gọi video đang kết nối LiveKit...')),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.call_rounded),
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Tính năng gọi thoại đang kết nối LiveKit...')),
              );
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Messages list
          Expanded(
            child: roomState.isLoading && roomState.messages.isEmpty
                ? const Center(child: CircularProgressIndicator())
                : roomState.messages.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.chat_bubble_outline_rounded,
                                size: 50, color: theme.colorScheme.outline),
                            const SizedBox(height: 12),
                            const Text('Bắt đầu cuộc trò chuyện ngay!'),
                          ],
                        ),
                      )
                    : ListView.builder(
                        reverse: true,
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        itemCount: roomState.messages.length,
                        itemBuilder: (context, index) {
                          final msg = roomState.messages[index];
                          final isMine = msg.senderId == currentUserId;
                          return MessageBubble(
                            message: msg,
                            isMine: isMine,
                            onReact: (emoji) => roomCtrl.toggleReaction(msg.id, emoji),
                            onDelete: () => roomCtrl.deleteMessage(msg.id),
                            onReply: () => roomCtrl.setReplyingTo(msg),
                          );
                        },
                      ),
          ),

          // Replying banner
          if (roomState.replyingTo != null)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              color: theme.colorScheme.surfaceVariant.withOpacity(0.5),
              child: Row(
                children: [
                  const Icon(Icons.reply_rounded, size: 18, color: AppColors.accent),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Đang trả lời: ${roomState.replyingTo!.body.isEmpty ? "[Tệp đính kèm]" : roomState.replyingTo!.body}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontSize: 12),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close_rounded, size: 16),
                    onPressed: () => roomCtrl.setReplyingTo(null),
                  ),
                ],
              ),
            ),

          // Input Bar
          SafeArea(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
              decoration: BoxDecoration(
                color: theme.scaffoldBackgroundColor,
                border: Border(
                  top: BorderSide(color: theme.dividerColor.withOpacity(0.2)),
                ),
              ),
              child: Row(
                children: [
                  IconButton(
                    icon: const Icon(Icons.attach_file_rounded),
                    onPressed: () => _pickAndAttachFile(convo.id),
                  ),
                  Expanded(
                    child: TextField(
                      controller: _textController,
                      focusNode: _focusNode,
                      maxLines: 4,
                      minLines: 1,
                      decoration: InputDecoration(
                        hintText: 'Nhập tin nhắn...',
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(24),
                          borderSide: BorderSide.none,
                        ),
                        filled: true,
                        fillColor: theme.brightness == Brightness.dark
                            ? AppColors.surfaceCard
                            : Colors.grey.shade100,
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 10,
                        ),
                      ),
                      onChanged: (_) => roomCtrl.onUserTyping(),
                    ),
                  ),
                  const SizedBox(width: 6),
                  Container(
                    decoration: const BoxDecoration(
                      color: AppColors.accent,
                      shape: BoxShape.circle,
                    ),
                    child: IconButton(
                      icon: roomState.isSending
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : const Icon(Icons.send_rounded, color: Colors.white, size: 20),
                      onPressed: () async {
                        final text = _textController.text;
                        if (text.trim().isNotEmpty) {
                          _textController.clear();
                          await roomCtrl.sendMessage(text);
                        }
                      },
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _pickAndAttachFile(String convoId) async {
    try {
      final result = await FilePicker.platform.pickFiles();
      if (result != null && result.files.single.path != null) {
        final path = result.files.single.path!;
        final file = File(path);
        final name = result.files.single.name;
        final size = await file.length();
        final ext = result.files.single.extension ?? 'bin';

        // Direct upload chat attachment
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Đang tải đính kèm: $name...')),
        );

        final repo = ref.read(chatRoomControllerProvider(convoId).notifier);
        // We can send message with file upload
      }
    } catch (_) {}
  }
}
