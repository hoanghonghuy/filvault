import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/chat_model.dart';
import '../../files/presentation/widgets/file_icon_helper.dart';

class MessageBubble extends StatelessWidget {
  final ChatMessageModel message;
  final bool isMine;
  final Function(String emoji) onReact;
  final VoidCallback onDelete;
  final VoidCallback onReply;

  const MessageBubble({
    super.key,
    required this.message,
    required this.isMine,
    required this.onReact,
    required this.onDelete,
    required this.onReply,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Column(
        crossAxisAlignment:
            isMine ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          GestureDetector(
            onDoubleTap: () => onReact('❤️'),
            onLongPress: () => _showContextMenu(context),
            child: Container(
              constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.76,
              ),
              decoration: BoxDecoration(
                color: isMine
                    ? AppColors.accent
                    : (theme.brightness == Brightness.dark
                        ? AppColors.surfaceCard
                        : Colors.grey.shade200),
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(18),
                  topRight: const Radius.circular(18),
                  bottomLeft: Radius.circular(isMine ? 18 : 4),
                  bottomRight: Radius.circular(isMine ? 4 : 18),
                ),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Attachments
                  if (message.attachments.isNotEmpty) ...[
                    for (final att in message.attachments)
                      _buildAttachment(context, att),
                    if (message.body.isNotEmpty) const SizedBox(height: 6),
                  ],

                  // Text body
                  if (message.body.isNotEmpty)
                    Text(
                      message.body,
                      style: TextStyle(
                        color: isMine ? Colors.white : theme.colorScheme.onSurface,
                        fontSize: 15,
                        height: 1.3,
                      ),
                    ),

                  const SizedBox(height: 4),
                  // Timestamp
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        '${message.createdAt.hour.toString().padLeft(2, '0')}:${message.createdAt.minute.toString().padLeft(2, '0')}',
                        style: TextStyle(
                          color: (isMine ? Colors.white70 : theme.colorScheme.outline)
                              .withOpacity(0.8),
                          fontSize: 10,
                        ),
                      ),
                      if (isMine) ...[
                        const SizedBox(width: 4),
                        const Icon(
                          Icons.done_all_rounded,
                          size: 13,
                          color: Colors.white70,
                        ),
                      ],
                    ],
                  ),
                ],
              ),
            ),
          ),

          // Reactions row
          if (message.reactions.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Wrap(
                spacing: 4,
                children: message.reactions.map((r) {
                  return GestureDetector(
                    onTap: () => onReact(r.reaction),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                        color: r.reacted
                            ? AppColors.accent.withOpacity(0.2)
                            : theme.colorScheme.surfaceVariant,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: r.reacted ? AppColors.accent : Colors.transparent,
                          width: 1,
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(r.reaction, style: const TextStyle(fontSize: 12)),
                          const SizedBox(width: 4),
                          Text(
                            '${r.count}',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: r.reacted
                                  ? AppColors.accent
                                  : theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildAttachment(BuildContext context, ChatAttachmentModel att) {
    if (att.isImage && att.thumbnailUrl != null) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 6),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: CachedNetworkImage(
            imageUrl: att.thumbnailUrl!,
            height: 160,
            width: double.infinity,
            fit: BoxFit.cover,
            placeholder: (context, url) => Container(
              height: 160,
              color: Colors.black12,
              child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
            ),
          ),
        ),
      );
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.08),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(FileIconHelper.getIcon(att.mimeType), size: 24),
          const SizedBox(width: 8),
          Flexible(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  att.originalName,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 12),
                ),
                Text(
                  FileIconHelper.formatBytes(att.sizeBytes),
                  style: const TextStyle(fontSize: 10),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _showContextMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Quick emoji bar
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: ['❤️', '👍', '😂', '😮', '😢', '🔥'].map((emoji) {
                  return GestureDetector(
                    onTap: () {
                      Navigator.pop(ctx);
                      onReact(emoji);
                    },
                    child: Text(emoji, style: const TextStyle(fontSize: 28)),
                  );
                }).toList(),
              ),
            ),
            const Divider(height: 1),
            ListTile(
              leading: const Icon(Icons.reply_rounded),
              title: const Text('Trả lời'),
              onTap: () {
                Navigator.pop(ctx);
                onReply();
              },
            ),
            if (message.body.isNotEmpty)
              ListTile(
                leading: const Icon(Icons.copy_rounded),
                title: const Text('Sao chép văn bản'),
                onTap: () {
                  Clipboard.setData(ClipboardData(text: message.body));
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Đã sao chép vào bộ nhớ tạm')),
                  );
                },
              ),
            if (isMine)
              ListTile(
                leading: Icon(Icons.delete_outline_rounded,
                    color: Theme.of(context).colorScheme.error),
                title: Text('Xóa tin nhắn',
                    style: TextStyle(color: Theme.of(context).colorScheme.error)),
                onTap: () {
                  Navigator.pop(ctx);
                  onDelete();
                },
              ),
          ],
        ),
      ),
    );
  }
}
