import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/photo_model.dart';
import '../controllers/photos_controller.dart';
import '../widgets/album_card.dart';
import '../widgets/photo_grid_item.dart';
import 'album_detail_screen.dart';
import 'media_lightbox_screen.dart';

class PhotosScreen extends ConsumerStatefulWidget {
  const PhotosScreen({super.key});

  @override
  ConsumerState<PhotosScreen> createState() => _PhotosScreenState();
}

class _PhotosScreenState extends ConsumerState<PhotosScreen>
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
    final state = ref.watch(photosControllerProvider);
    final ctrl = ref.read(photosControllerProvider.notifier);
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: const Text('Kho ảnh & Album', style: TextStyle(fontWeight: FontWeight.w700)),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppColors.accent,
          labelColor: AppColors.accent,
          unselectedLabelColor: theme.colorScheme.onSurfaceVariant,
          tabs: const [
            Tab(text: 'Dòng thời gian'),
            Tab(text: 'Bộ sưu tập'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          // Tab 1: Timeline
          RefreshIndicator(
            onRefresh: () => ctrl.loadTimeline(),
            child: _buildTimelineView(state, ctrl),
          ),

          // Tab 2: Albums
          RefreshIndicator(
            onRefresh: () => ctrl.loadAlbums(),
            child: _buildAlbumsView(state, ctrl),
          ),
        ],
      ),
      floatingActionButton: _tabController.index == 1
          ? FloatingActionButton.extended(
              backgroundColor: AppColors.accent,
              foregroundColor: Colors.white,
              onPressed: () => _showCreateAlbumDialog(context),
              icon: const Icon(Icons.add_photo_alternate_rounded),
              label: const Text('Tạo Album'),
            )
          : null,
    );
  }

  Widget _buildTimelineView(PhotosState state, PhotosController ctrl) {
    if (state.isLoading && state.timelineGroups.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.timelineGroups.isEmpty) {
      return Center(
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
              child: const Icon(Icons.photo_library_outlined, size: 38, color: AppColors.accent),
            ),
            const SizedBox(height: 16),
            const Text(
              'Chưa có hình ảnh nào',
              style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 6),
            Text(
              'Tải ảnh lên tab Tệp để xem dòng thời gian tại đây',
              style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
            ),
          ],
        ),
      );
    }

    return NotificationListener<ScrollNotification>(
      onNotification: (ScrollNotification scrollInfo) {
        if (scrollInfo.metrics.pixels >= scrollInfo.metrics.maxScrollExtent - 200) {
          ctrl.loadMoreTimeline();
        }
        return false;
      },
      child: CustomScrollView(
        slivers: [
          for (final group in state.timelineGroups) ...[
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                child: Text(
                  _formatGroupDate(group.date),
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 15,
                  ),
                ),
              ),
            ),
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              sliver: SliverGrid(
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3,
                  crossAxisSpacing: 3,
                  mainAxisSpacing: 3,
                ),
                delegate: SliverChildBuilderDelegate(
                  (context, index) {
                    final item = group.items[index];
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
                  childCount: group.items.length,
                ),
              ),
            ),
          ],
          if (state.isLoadingMore)
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(child: CircularProgressIndicator()),
              ),
            ),
          const SliverToBoxAdapter(child: SizedBox(height: 80)),
        ],
      ),
    );
  }

  Widget _buildAlbumsView(PhotosState state, PhotosController ctrl) {
    if (state.albums.isEmpty) {
      return Center(
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
              child: const Icon(Icons.collections_bookmark_outlined, size: 38, color: AppColors.accent),
            ),
            const SizedBox(height: 16),
            const Text(
              'Chưa có album nào',
              style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 6),
            Text(
              'Nhấn "Tạo Album" bên dưới để nhóm các khoảnh khắc',
              style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
            ),
          ],
        ),
      );
    }

    return GridView.builder(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 80),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 0.8,
        crossAxisSpacing: 16,
        mainAxisSpacing: 16,
      ),
      itemCount: state.albums.length,
      itemBuilder: (context, index) {
        final album = state.albums[index];
        return AlbumCard(
          album: album,
          onTap: () {
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => AlbumDetailScreen(
                  albumId: album.id,
                  albumName: album.name,
                ),
              ),
            );
          },
        );
      },
    );
  }

  String _formatGroupDate(String rawDate) {
    final now = DateTime.now();
    final parts = rawDate.split('-');
    if (parts.length == 3) {
      final year = int.tryParse(parts[0]) ?? now.year;
      final month = int.tryParse(parts[1]) ?? now.month;
      final day = int.tryParse(parts[2]) ?? now.day;
      final dt = DateTime(year, month, day);

      if (now.year == dt.year && now.month == dt.month && now.day == dt.day) {
        return 'Hôm nay';
      }
      final yesterday = now.subtract(const Duration(days: 1));
      if (yesterday.year == dt.year && yesterday.month == dt.month && yesterday.day == dt.day) {
        return 'Hôm qua';
      }
      return '$day tháng $month, $year';
    }
    return rawDate;
  }

  void _showCreateAlbumDialog(BuildContext context) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Tạo album mới'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(hintText: 'Tên album'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () {
              final name = controller.text.trim();
              if (name.isNotEmpty) {
                ref.read(photosControllerProvider.notifier).createAlbum(name);
              }
              Navigator.pop(ctx);
            },
            child: const Text('Tạo'),
          ),
        ],
      ),
    );
  }
}
