import 'package:flutter/material.dart';

class FileIconHelper {
  static IconData getIcon(String mimeType) {
    if (mimeType.startsWith('image/')) return Icons.image_rounded;
    if (mimeType.startsWith('video/')) return Icons.videocam_rounded;
    if (mimeType.startsWith('audio/')) return Icons.audiotrack_rounded;
    if (mimeType == 'application/pdf') return Icons.picture_as_pdf_rounded;
    if (mimeType.contains('zip') || mimeType.contains('tar') || mimeType.contains('compressed')) {
      return Icons.folder_zip_rounded;
    }
    if (mimeType.contains('word') || mimeType.contains('document')) {
      return Icons.description_rounded;
    }
    if (mimeType.contains('sheet') || mimeType.contains('excel')) {
      return Icons.table_chart_rounded;
    }
    if (mimeType.contains('presentation') || mimeType.contains('powerpoint')) {
      return Icons.slideshow_rounded;
    }
    if (mimeType.contains('text/')) return Icons.text_snippet_rounded;
    return Icons.insert_drive_file_rounded;
  }

  static Color getColor(String mimeType, BuildContext context) {
    if (mimeType.startsWith('image/')) return Colors.teal;
    if (mimeType.startsWith('video/')) return Colors.deepOrange;
    if (mimeType.startsWith('audio/')) return Colors.purple;
    if (mimeType == 'application/pdf') return Colors.redAccent;
    if (mimeType.contains('zip') || mimeType.contains('tar') || mimeType.contains('compressed')) {
      return Colors.amber.shade700;
    }
    if (mimeType.contains('word') || mimeType.contains('document')) {
      return Colors.blue;
    }
    if (mimeType.contains('sheet') || mimeType.contains('excel')) {
      return Colors.green;
    }
    if (mimeType.contains('presentation') || mimeType.contains('powerpoint')) {
      return Colors.orange;
    }
    return Theme.of(context).colorScheme.outline;
  }

  static String formatBytes(int bytes) {
    if (bytes <= 0) return '0 B';
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }
}
