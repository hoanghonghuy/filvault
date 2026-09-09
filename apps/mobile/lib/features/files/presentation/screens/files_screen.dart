import 'dart:io';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/models/file_item_model.dart';
import '../controllers/files_controller.dart';
import '../widgets/breadcrumb_bar.dart';
import '../widgets/file_actions_bottom_sheet.dart';
import '../widgets/file_icon_helper.dart';
import '../widgets/folder_actions_bottom_sheet.dart';

class FilesScreen extends ConsumerStatefulWidget {
  const FilesScreen({super.key});

  @override
  ConsumerState<FilesScreen> createState() => _FilesScreenState();
}

class _FilesScreenState extends ConsumerState<FilesScreen> {
  final TextEditingController _searchController = TextEditingController();
  bool _isSearching = false;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(filesControllerProvider);
    final filesCtrl = ref.read(filesControllerProvider.notifier);
    final theme = Theme.of(context);

    final searchQuery = _searchController.text.trim().toLowerCase();
    final filteredFolders = searchQuery.isEmpty
        ? state.folders
        : state.folders.where((f) => f.name.toLowerCase().contains(searchQuery)).toList();
    final filteredFiles = searchQuery.isEmpty
        ? state.files
        : state.files.where((f) => f.name.toLowerCase().contains(searchQuery)).toList();

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        title: _isSearching
            ? TextField(
                controller: _searchController,
                autofocus: true,
                decoration: const InputDecoration(
                  hintText: 'Tìm kiếm tệp, thư mục...',
                  border: InputBorder.none,
                ),
                onChanged: (_) => setState(() {}),
              )
            : Text(
                state.currentFolder?.name ?? 'Tệp của tôi',
                style: const TextStyle(fontWeight: FontWeight.w700),
              ),
        actions: [
          IconButton(
            icon: Icon(_isSearching ? Icons.close : Icons.search_rounded),
            onPressed: () {
              setState(() {
                if (_isSearching) {
                  _searchController.clear();
                  _isSearching = false;
                } else {
                  _isSearching = true;
                }
              });
            },
          ),
          IconButton(
            icon: Icon(
              state.isGridView ? Icons.view_list_rounded : Icons.grid_view_rounded,
            ),
            tooltip: state.isGridView ? 'Chế độ danh sách' : 'Chế độ lưới',
            onPressed: () => filesCtrl.toggleViewMode(),
          ),
        ],
      ),
      body: Column(
        children: [
          // Breadcrumb Navigation
          BreadcrumbBar(
            breadcrumb: state.breadcrumb,
            currentFolder: state.currentFolder,
            onFolderTap: (folder) => filesCtrl.navigateToBreadcrumb(folder),
          ),

          // Upload Progress Banner
          if (state.uploadProgress != null)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              color: AppColors.accent.withOpacity(0.12),
              child: Row(
                children: [
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2.5),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Đang tải lên: ${state.uploadingFileName ?? "Tệp"}',
                          style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        LinearProgressIndicator(
                          value: state.uploadProgress,
                          backgroundColor: Colors.grey.withOpacity(0.2),
                          valueColor: const AlwaysStoppedAnimation<Color>(AppColors.accent),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    '${((state.uploadProgress ?? 0) * 100).toInt()}%',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
            ),

          // Error Banner
          if (state.errorMessage != null)
            Container(
              margin: const EdgeInsets.all(12),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: theme.colorScheme.errorContainer,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                children: [
                  Icon(Icons.error_outline_rounded, color: theme.colorScheme.error),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      state.errorMessage!,
                      style: TextStyle(color: theme.colorScheme.onErrorContainer, fontSize: 13),
                    ),
                  ),
                ],
              ),
            ),

          // Main Browser Body
          Expanded(
            child: RefreshIndicator(
              onRefresh: () => filesCtrl.loadBrowser(folderId: state.currentFolder?.id),
              child: state.isLoading && state.folders.isEmpty && state.files.isEmpty
                  ? const Center(child: CircularProgressIndicator())
                  : (filteredFolders.isEmpty && filteredFiles.isEmpty)
                      ? _buildEmptyState(context)
                      : state.isGridView
                          ? _buildGridView(context, filteredFolders, filteredFiles)
                          : _buildListView(context, filteredFolders, filteredFiles),
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        backgroundColor: AppColors.accent,
        foregroundColor: Colors.white,
        elevation: 4,
        onPressed: () => _showAddOptions(context),
        child: const Icon(Icons.add_rounded, size: 28),
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return ListView(
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.2),
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
                child: const Icon(Icons.folder_open_rounded, size: 38, color: AppColors.accent),
              ),
              const SizedBox(height: 16),
              const Text(
                'Thư mục này trống',
                style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 6),
              Text(
                'Nhấn nút + bên dưới để tạo thư mục hoặc tải tệp lên',
                style: TextStyle(fontSize: 13, color: Theme.of(context).colorScheme.outline),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildGridView(
    BuildContext context,
    List<FolderModel> folders,
    List<FileItemModel> files,
  ) {
    return CustomScrollView(
      slivers: [
        if (folders.isNotEmpty) ...[
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: Text(
                'Thư mục (${folders.length})',
                style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverGrid(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 2.2,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
              ),
              delegate: SliverChildBuilderDelegate(
                (context, index) => _buildFolderCard(context, folders[index]),
                childCount: folders.length,
              ),
            ),
          ),
        ],
        if (files.isNotEmpty) ...[
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
              child: Text(
                'Tệp (${files.length})',
                style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverGrid(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 1.1,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
              ),
              delegate: SliverChildBuilderDelegate(
                (context, index) => _buildFileGridCard(context, files[index]),
                childCount: files.length,
              ),
            ),
          ),
        ],
        const SliverToBoxAdapter(child: SizedBox(height: 80)),
      ],
    );
  }

  Widget _buildListView(
    BuildContext context,
    List<FolderModel> folders,
    List<FileItemModel> files,
  ) {
    return ListView(
      padding: const EdgeInsets.only(bottom: 80),
      children: [
        if (folders.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: Text(
              'Thư mục (${folders.length})',
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
            ),
          ),
          ...folders.map((f) => _buildFolderListTile(context, f)),
        ],
        if (files.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text(
              'Tệp (${files.length})',
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
            ),
          ),
          ...files.map((f) => _buildFileListTile(context, f)),
        ],
      ],
    );
  }

  Widget _buildFolderCard(BuildContext context, FolderModel folder) {
    return Material(
      color: Theme.of(context).cardColor,
      borderRadius: BorderRadius.circular(12),
      elevation: 0.5,
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => ref.read(filesControllerProvider.notifier).openFolder(folder),
        onLongPress: () => FolderActionsBottomSheet.show(context, folder),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            children: [
              const Icon(Icons.folder_rounded, color: AppColors.accent, size: 30),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  folder.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                ),
              ),
              InkWell(
                borderRadius: BorderRadius.circular(20),
                onTap: () => FolderActionsBottomSheet.show(context, folder),
                child: const Padding(
                  padding: EdgeInsets.all(4),
                  child: Icon(Icons.more_vert_rounded, size: 18),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFolderListTile(BuildContext context, FolderModel folder) {
    return ListTile(
      leading: const Icon(Icons.folder_rounded, color: AppColors.accent, size: 32),
      title: Text(folder.name, style: const TextStyle(fontWeight: FontWeight.w600)),
      trailing: IconButton(
        icon: const Icon(Icons.more_vert_rounded, size: 20),
        onPressed: () => FolderActionsBottomSheet.show(context, folder),
      ),
      onTap: () => ref.read(filesControllerProvider.notifier).openFolder(folder),
      onLongPress: () => FolderActionsBottomSheet.show(context, folder),
    );
  }

  Widget _buildFileGridCard(BuildContext context, FileItemModel file) {
    final color = FileIconHelper.getColor(file.mimeType, context);
    final icon = FileIconHelper.getIcon(file.mimeType);

    return Material(
      color: Theme.of(context).cardColor,
      borderRadius: BorderRadius.circular(14),
      elevation: 0.5,
      child: InkWell(
        borderRadius: BorderRadius.circular(14),
        onTap: () => FileActionsBottomSheet.show(context, file),
        onLongPress: () => FileActionsBottomSheet.show(context, file),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.12),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Icon(icon, color: color, size: 24),
                  ),
                  if (file.isFavorite)
                    const Icon(Icons.star_rounded, color: Colors.amber, size: 18)
                  else
                    InkWell(
                      borderRadius: BorderRadius.circular(20),
                      onTap: () => FileActionsBottomSheet.show(context, file),
                      child: const Icon(Icons.more_vert_rounded, size: 18),
                    ),
                ],
              ),
              const Spacer(),
              Text(
                file.name,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
              ),
              const SizedBox(height: 4),
              Text(
                FileIconHelper.formatBytes(file.sizeBytes),
                style: TextStyle(
                  fontSize: 11,
                  color: Theme.of(context).colorScheme.outline,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFileListTile(BuildContext context, FileItemModel file) {
    final color = FileIconHelper.getColor(file.mimeType, context);
    final icon = FileIconHelper.getIcon(file.mimeType);

    return ListTile(
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(icon, color: color, size: 22),
      ),
      title: Text(
        file.name,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      subtitle: Text(
        '${FileIconHelper.formatBytes(file.sizeBytes)} • ${_formatDate(file.updatedAt)}',
        style: const TextStyle(fontSize: 12),
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (file.isFavorite)
            const Icon(Icons.star_rounded, color: Colors.amber, size: 18),
          IconButton(
            icon: const Icon(Icons.more_vert_rounded, size: 20),
            onPressed: () => FileActionsBottomSheet.show(context, file),
          ),
        ],
      ),
      onTap: () => FileActionsBottomSheet.show(context, file),
      onLongPress: () => FileActionsBottomSheet.show(context, file),
    );
  }

  String _formatDate(DateTime dt) {
    return '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year}';
  }

  void _showAddOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 36,
              height: 4,
              margin: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.create_new_folder_rounded, color: AppColors.accent),
              title: const Text('Thư mục mới'),
              onTap: () {
                Navigator.pop(ctx);
                _showCreateFolderDialog(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.upload_file_rounded, color: Colors.blue),
              title: const Text('Tải tệp lên'),
              onTap: () {
                Navigator.pop(ctx);
                _pickAndUploadFile();
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showCreateFolderDialog(BuildContext context) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Tạo thư mục mới'),
        content: TextField(
          controller: controller,
          autofocus: true,
          decoration: const InputDecoration(hintText: 'Nhập tên thư mục'),
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
                ref.read(filesControllerProvider.notifier).createFolder(name);
              }
              Navigator.pop(ctx);
            },
            child: const Text('Tạo'),
          ),
        ],
      ),
    );
  }

  Future<void> _pickAndUploadFile() async {
    try {
      final result = await FilePicker.platform.pickFiles();
      if (result != null && result.files.single.path != null) {
        final path = result.files.single.path!;
        final file = File(path);
        final name = result.files.single.name;
        final extension = result.files.single.extension ?? 'bin';
        final mimeType = _getMimeTypeForExtension(extension);

        await ref.read(filesControllerProvider.notifier).uploadFile(file, name, mimeType);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Lỗi chọn tệp: $e')),
        );
      }
    }
  }

  String _getMimeTypeForExtension(String ext) {
    switch (ext.toLowerCase()) {
      case 'jpg':
      case 'jpeg':
        return 'image/jpeg';
      case 'png':
        return 'image/png';
      case 'gif':
        return 'image/gif';
      case 'webp':
        return 'image/webp';
      case 'mp4':
        return 'video/mp4';
      case 'mp3':
        return 'audio/mpeg';
      case 'pdf':
        return 'application/pdf';
      case 'zip':
        return 'application/zip';
      case 'txt':
        return 'text/plain';
      default:
        return 'application/octet-stream';
    }
  }
}
