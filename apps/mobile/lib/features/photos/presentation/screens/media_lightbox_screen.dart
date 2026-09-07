import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../files/presentation/widgets/file_icon_helper.dart';
import '../../data/models/photo_model.dart';
import '../controllers/photos_controller.dart';

class MediaLightboxScreen extends ConsumerStatefulWidget {
  final TimelineItemModel item;

  const MediaLightboxScreen({super.key, required this.item});

  @override
  ConsumerState<MediaLightboxScreen> createState() => _MediaLightboxScreenState();
}

class _MediaLightboxScreenState extends ConsumerState<MediaLightboxScreen> {
  bool _showControls = true;
  String? _downloadUrl;
  bool _isLoadingUrl = false;

  @override
  void initState() {
    super.initState();
    _fetchFullUrl();
  }

  Future<void> _fetchFullUrl() async {
    setState(() => _isLoadingUrl = true);
    try {
      final repo = ref.read(photosRepositoryProvider);
      final url = await repo.getDownloadUrl(widget.item.id);
      if (mounted) {
        setState(() {
          _downloadUrl = url;
          _isLoadingUrl = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _isLoadingUrl = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final item = widget.item;

    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Zoomable viewer
          GestureDetector(
            onTap: () => setState(() => _showControls = !_showControls),
            child: Center(
              child: Hero(
                tag: 'photo_${item.id}',
                child: InteractiveViewer(
                  minScale: 0.8,
                  maxScale: 4.0,
                  child: _downloadUrl != null
                      ? CachedNetworkImage(
                          imageUrl: _downloadUrl!,
                          fit: BoxFit.contain,
                          placeholder: (context, url) => item.thumbnailUrl != null
                              ? CachedNetworkImage(imageUrl: item.thumbnailUrl!, fit: BoxFit.contain)
                              : const Center(child: CircularProgressIndicator(color: Colors.white)),
                          errorWidget: (context, url, error) => const Center(
                            child: Icon(Icons.broken_image_rounded, color: Colors.white54, size: 60),
                          ),
                        )
                      : item.thumbnailUrl != null
                          ? CachedNetworkImage(imageUrl: item.thumbnailUrl!, fit: BoxFit.contain)
                          : const Center(child: CircularProgressIndicator(color: Colors.white)),
                ),
              ),
            ),
          ),

          // Top App Bar Controls
          if (_showControls)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: Container(
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    colors: [Colors.black87, Colors.transparent],
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                  ),
                ),
                child: SafeArea(
                  child: Row(
                    children: [
                      IconButton(
                        icon: const Icon(Icons.arrow_back_rounded, color: Colors.white),
                        onPressed: () => Navigator.pop(context),
                      ),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item.name,
                              style: const TextStyle(
                                color: Colors.white,
                                fontWeight: FontWeight.w600,
                                fontSize: 15,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            Text(
                              '${item.createdAt.day}/${item.createdAt.month}/${item.createdAt.year}',
                              style: const TextStyle(color: Colors.white70, fontSize: 12),
                            ),
                          ],
                        ),
                      ),
                      if (_isLoadingUrl)
                        const Padding(
                          padding: EdgeInsets.all(12),
                          child: SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          ),
                        ),
                      IconButton(
                        icon: const Icon(Icons.info_outline_rounded, color: Colors.white),
                        onPressed: () => _showDetailsModal(context, item),
                      ),
                    ],
                  ),
                ),
              ),
            ),

          // Bottom Bar Controls
          if (_showControls)
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    colors: [Colors.transparent, Colors.black87],
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                  ),
                ),
                child: SafeArea(
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                    children: [
                      _buildBottomAction(
                        icon: Icons.download_rounded,
                        label: 'Tải xuống',
                        onTap: () {
                          if (_downloadUrl != null) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text('Liên kết tải: $_downloadUrl'),
                                behavior: SnackBarBehavior.floating,
                              ),
                            );
                          }
                        },
                      ),
                      _buildBottomAction(
                        icon: Icons.share_rounded,
                        label: 'Chia sẻ',
                        onTap: () {
                          if (_downloadUrl != null) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Đang tạo liên kết chia sẻ...'),
                                behavior: SnackBarBehavior.floating,
                              ),
                            );
                          }
                        },
                      ),
                    ],
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildBottomAction({
    required IconData icon,
    required String label,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, color: Colors.white, size: 24),
            const SizedBox(height: 4),
            Text(label, style: const TextStyle(color: Colors.white, fontSize: 12)),
          ],
        ),
      ),
    );
  }

  void _showDetailsModal(BuildContext context, TimelineItemModel item) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Chi tiết mục',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18),
              ),
              const SizedBox(height: 16),
              _buildDetailRow('Tên tệp', item.name),
              _buildDetailRow('Kích thước', FileIconHelper.formatBytes(item.sizeBytes)),
              _buildDetailRow('Định dạng', item.mimeType),
              _buildDetailRow(
                'Thời gian chụp / tạo',
                '${item.createdAt.hour.toString().padLeft(2, '0')}:${item.createdAt.minute.toString().padLeft(2, '0')} - ${item.createdAt.day}/${item.createdAt.month}/${item.createdAt.year}',
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 140,
            child: Text(
              label,
              style: TextStyle(color: Theme.of(context).colorScheme.outline),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ),
        ],
      ),
    );
  }
}
