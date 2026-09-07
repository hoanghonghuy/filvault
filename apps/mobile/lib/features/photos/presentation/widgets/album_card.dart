import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/photo_model.dart';

class AlbumCard extends StatelessWidget {
  final AlbumModel album;
  final VoidCallback onTap;
  final VoidCallback? onLongPress;

  const AlbumCard({
    super.key,
    required this.album,
    required this.onTap,
    this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return InkWell(
      onTap: onTap,
      onLongPress: onLongPress,
      borderRadius: BorderRadius.circular(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          AspectRatio(
            aspectRatio: 1,
            child: ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: album.coverUrl != null
                  ? CachedNetworkImage(
                      imageUrl: album.coverUrl!,
                      fit: BoxFit.cover,
                      placeholder: (context, url) => Container(
                        color: Colors.grey.withOpacity(0.15),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey.withOpacity(0.15),
                        child: const Icon(Icons.photo_library_rounded, color: AppColors.muted),
                      ),
                    )
                  : Container(
                      color: AppColors.accent.withOpacity(0.1),
                      child: const Center(
                        child: Icon(Icons.photo_library_rounded, color: AppColors.accent, size: 36),
                      ),
                    ),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            album.name,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            '${album.itemCount} mục',
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.outline,
            ),
          ),
        ],
      ),
    );
  }
}
