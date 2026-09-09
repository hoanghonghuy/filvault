import 'dart:io';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/api/api_client.dart';
import '../../data/files_repository.dart';
import '../../data/models/file_item_model.dart';

final filesRepositoryProvider = Provider<FilesRepository>((ref) {
  return FilesRepositoryImpl();
});

class FilesState {
  final bool isLoading;
  final String? errorMessage;
  final FolderModel? currentFolder;
  final List<FolderModel> breadcrumb;
  final List<FolderModel> folders;
  final List<FileItemModel> files;
  final double? uploadProgress;
  final String? uploadingFileName;
  final bool isGridView;

  const FilesState({
    this.isLoading = false,
    this.errorMessage,
    this.currentFolder,
    this.breadcrumb = const [],
    this.folders = const [],
    this.files = const [],
    this.uploadProgress,
    this.uploadingFileName,
    this.isGridView = true,
  });

  FilesState copyWith({
    bool? isLoading,
    String? errorMessage,
    FolderModel? currentFolder,
    bool clearCurrentFolder = false,
    List<FolderModel>? breadcrumb,
    List<FolderModel>? folders,
    List<FileItemModel>? files,
    double? uploadProgress,
    bool clearUpload = false,
    String? uploadingFileName,
    bool? isGridView,
  }) {
    return FilesState(
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      currentFolder: clearCurrentFolder ? null : (currentFolder ?? this.currentFolder),
      breadcrumb: breadcrumb ?? this.breadcrumb,
      folders: folders ?? this.folders,
      files: files ?? this.files,
      uploadProgress: clearUpload ? null : (uploadProgress ?? this.uploadProgress),
      uploadingFileName: clearUpload ? null : (uploadingFileName ?? this.uploadingFileName),
      isGridView: isGridView ?? this.isGridView,
    );
  }
}

class FilesController extends StateNotifier<FilesState> {
  final FilesRepository _repo;
  final ApiClient _apiClient;

  FilesController(this._repo, {ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance,
        super(const FilesState()) {
    loadBrowser();
  }

  void toggleViewMode() {
    state = state.copyWith(isGridView: !state.isGridView);
  }

  Future<void> loadBrowser({String? folderId}) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final result = await _repo.getBrowser(folderId: folderId);
      state = state.copyWith(
        isLoading: false,
        currentFolder: result.currentFolder,
        clearCurrentFolder: result.currentFolder == null,
        breadcrumb: result.breadcrumb,
        folders: result.folders,
        files: result.files,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> openFolder(FolderModel folder) async {
    await loadBrowser(folderId: folder.id);
  }

  Future<void> navigateToBreadcrumb(FolderModel? folder) async {
    await loadBrowser(folderId: folder?.id);
  }

  Future<bool> createFolder(String name) async {
    try {
      await _repo.createFolder(
        name,
        parentId: state.currentFolder?.id,
      );
      await loadBrowser(folderId: state.currentFolder?.id);
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> renameFolder(String folderId, String newName) async {
    try {
      await _repo.renameFolder(folderId, newName);
      await loadBrowser(folderId: state.currentFolder?.id);
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> deleteFolder(String folderId) async {
    try {
      await _repo.deleteFolder(folderId);
      await loadBrowser(folderId: state.currentFolder?.id);
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> renameFile(String fileId, String newName) async {
    try {
      await _repo.renameFile(fileId, newName);
      await loadBrowser(folderId: state.currentFolder?.id);
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> toggleFavorite(String fileId, bool isFavorite) async {
    try {
      await _repo.toggleFavorite(fileId, isFavorite);
      state = state.copyWith(
        files: state.files.map((f) {
          if (f.id == fileId) {
            return f.copyWith(isFavorite: isFavorite);
          }
          return f;
        }).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> deleteFile(String fileId) async {
    try {
      await _repo.deleteFile(fileId);
      state = state.copyWith(
        files: state.files.where((f) => f.id != fileId).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<String?> getDownloadUrl(String fileId) async {
    try {
      return await _repo.getDownloadUrl(fileId);
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return null;
    }
  }

  Future<bool> uploadFile(File file, String name, String mimeType) async {
    final size = await file.length();
    state = state.copyWith(
      uploadProgress: 0.01,
      uploadingFileName: name,
      errorMessage: null,
    );

    try {
      // 1. Create upload session
      final session = await _repo.createUploadSession(
        name: name,
        size: size,
        contentType: mimeType,
        folderId: state.currentFolder?.id,
      );

      final uploadUrl = session['uploadUrl'] as String;
      final fileId = session['fileId'] as String;

      // 2. Direct S3 Upload via stream
      await _apiClient.uploadToS3Direct(
        uploadUrl: uploadUrl,
        fileStream: file.openRead(),
        contentLength: size,
        contentType: mimeType,
        onProgress: (sent, total) {
          if (total > 0) {
            state = state.copyWith(uploadProgress: sent / total);
          }
        },
      );

      // 3. Complete upload session
      await _repo.completeUpload(fileId);

      // 4. Reload folder contents
      state = state.copyWith(clearUpload: true);
      await loadBrowser(folderId: state.currentFolder?.id);
      return true;
    } catch (e) {
      state = state.copyWith(
        clearUpload: true,
        errorMessage: 'Lỗi tải lên: ${e.toString().replaceFirst('Exception: ', '')}',
      );
      return false;
    }
  }
}

final filesControllerProvider =
    StateNotifierProvider<FilesController, FilesState>((ref) {
  final repo = ref.watch(filesRepositoryProvider);
  return FilesController(repo);
});
