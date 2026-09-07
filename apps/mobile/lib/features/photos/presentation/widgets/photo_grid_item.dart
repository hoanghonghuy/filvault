import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/photo_model.dart';

class PhotoGridItem extends StatelessWidget {
  final TimelineItemModel item;
  final VoidCallback onTap;

  const PhotoGridItem({
    super.key,
    required this.item,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Stack(
        fit: StackFit.expand,
        children: [
          Hero(
            tag: 'photo_${item.id}',
            child: ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: item.thumbnailUrl != null
                  ? CachedNetworkImage(
                      imageUrl: item.thumbnailUrl!,
                      fit: BoxFit.cover,
                      placeholder: (context, url) => Container(
                        color: Colors.grey.withOpacity(0.15),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey.withOpacity(0.15),
                        child: const Icon(Icons.broken_image_rounded, color: Colors.grey),
                      ),
                    )
                  : Container(
                      color: Colors.grey.withOpacity(0.15),
                      child: Icon(
                        item.isVideo ? Icons.videocam_rounded : Icons.image_rounded,
                        color: AppColors.muted,
                      ),
                    ),
            ),
          ),
          if (item.isVideo)
            Positioned(
              right: 6,
              bottom: 6,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.65),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.play_arrow_rounded, color: Colors.white, size: 14),
                    SizedBox(width: 2),
                    Text(
                      'Video',
                      style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }
}
