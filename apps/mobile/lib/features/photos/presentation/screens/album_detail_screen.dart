import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/photo_model.dart';
import '../controllers/photos_controller.dart';
import '../widgets/photo_grid_item.dart';
import 'media_lightbox_screen.dart';

class AlbumDetailScreen extends ConsumerStatefulWidget {
  final String albumId;
  final String albumName;

  const AlbumDetailScreen({
    super.key,
    required this.albumId,
    required this.albumName,
  });

  @override
  ConsumerState<AlbumDetailScreen> createState() => _AlbumDetailScreenState();
}

class _AlbumDetailScreenState extends ConsumerState<AlbumDetailScreen> {
  AlbumDetailModel? _detail;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _fetchDetail();
  }

  Future<void> _fetchDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final repo = ref.read(photosRepositoryProvider);
      final data = await repo.getAlbumDetail(widget.albumId);
      if (mounted) {
        setState(() {
          _detail = data;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString().replaceFirst('Exception: ', '');
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.albumName),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_outline_rounded),
            onPressed: () => _confirmDeleteAlbum(context),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(child: Text(_errorMessage!))
              : _detail == null || _detail!.items.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.photo_album_outlined, size: 60, color: theme.colorScheme.outline),
                          const SizedBox(height: 12),
                          const Text('Album này chưa có ảnh nào', style: TextStyle(fontSize: 16)),
                        ],
                      ),
                    )
                  : GridView.builder(
                      padding: const EdgeInsets.all(4),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 3,
                        crossAxisSpacing: 3,
                        mainAxisSpacing: 3,
                      ),
                      itemCount: _detail!.items.length,
                      itemBuilder: (context, index) {
                        final item = _detail!.items[index];
                        return PhotoGridItem(
                          item: item,
                          onTap: () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => MediaLightboxScreen(item: item),
                              ),
                            );
                          },
                        );
                      },
                    ),
    );
  }

  void _confirmDeleteAlbum(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Xóa album?'),
        content: Text('Album "${widget.albumName}" sẽ bị xóa. Các ảnh gốc trong tệp không bị ảnh hưởng.'),
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
              await ref.read(photosControllerProvider.notifier).deleteAlbum(widget.albumId);
              if (context.mounted) Navigator.pop(context);
            },
            child: const Text('Xóa'),
          ),
        ],
      ),
    );
  }
}
