import 'package:flutter/material.dart';
import '../../data/models/file_item_model.dart';

class BreadcrumbBar extends StatelessWidget {
  final List<FolderModel> breadcrumb;
  final FolderModel? currentFolder;
  final ValueChanged<FolderModel?> onFolderTap;

  const BreadcrumbBar({
    super.key,
    required this.breadcrumb,
    this.currentFolder,
    required this.onFolderTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          // Root item
          InkWell(
            borderRadius: BorderRadius.circular(8),
            onTap: () => onFolderTap(null),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: Row(
                children: [
                  Icon(
                    Icons.cloud_outlined,
                    size: 18,
                    color: currentFolder == null
                        ? theme.colorScheme.primary
                        : theme.colorScheme.onSurfaceVariant,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'Tệp của tôi',
                    style: TextStyle(
                      fontWeight: currentFolder == null
                          ? FontWeight.bold
                          : FontWeight.normal,
                      color: currentFolder == null
                          ? theme.colorScheme.primary
                          : theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Intermediate breadcrumb items
          for (final f in breadcrumb) ...[
            Icon(
              Icons.chevron_right_rounded,
              size: 18,
              color: theme.colorScheme.outline,
            ),
            InkWell(
              borderRadius: BorderRadius.circular(8),
              onTap: () => onFolderTap(f),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                child: Text(
                  f.name,
                  style: TextStyle(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
            ),
          ],

          // Current active folder (if not root and not in crumb)
          if (currentFolder != null &&
              !breadcrumb.any((b) => b.id == currentFolder!.id)) ...[
            Icon(
              Icons.chevron_right_rounded,
              size: 18,
              color: theme.colorScheme.outline,
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              child: Text(
                currentFolder!.name,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.primary,
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
