import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/file_item_model.dart';

abstract class FilesRepository {
  Future<BrowserDataModel> getBrowser({String? folderId});
  Future<FolderModel> createFolder(String name, {String? parentId});
  Future<FolderModel> renameFolder(String folderId, String name);
  Future<void> deleteFolder(String folderId);
  Future<Map<String, dynamic>> createUploadSession({
    required String name,
    required int size,
    required String contentType,
    String? folderId,
  });
  Future<FileItemModel> completeUpload(String fileId);
  Future<String> getDownloadUrl(String fileId);
  Future<void> toggleFavorite(String fileId, bool isFavorite);
  Future<FileItemModel> renameFile(String fileId, String name);
  Future<void> deleteFile(String fileId);
}

class FilesRepositoryImpl implements FilesRepository {
  final ApiClient _apiClient;

  FilesRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<BrowserDataModel> getBrowser({String? folderId}) async {
    try {
      final response = await _apiClient.get(
        ApiEndpoints.browser,
        queryParameters: folderId != null ? {'folderId': folderId} : null,
      );
      return BrowserDataModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<FolderModel> createFolder(String name, {String? parentId}) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.folders,
        data: {
          'name': name,
          if (parentId != null) 'parentId': parentId,
        },
      );
      return FolderModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<FolderModel> renameFolder(String folderId, String name) async {
    try {
      final response = await _apiClient.patch(
        '${ApiEndpoints.folders}/$folderId',
        data: {'name': name},
      );
      return FolderModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> deleteFolder(String folderId) async {
    try {
      await _apiClient.delete('${ApiEndpoints.folders}/$folderId');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<Map<String, dynamic>> createUploadSession({
    required String name,
    required int size,
    required String contentType,
    String? folderId,
  }) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.uploadSessions,
        data: {
          'name': name,
          'size': size,
          'contentType': contentType,
          if (folderId != null) 'folderId': folderId,
        },
      );
      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<FileItemModel> completeUpload(String fileId) async {
    try {
      final response = await _apiClient.post(
        '${ApiEndpoints.files}/$fileId/complete',
      );
      return FileItemModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<String> getDownloadUrl(String fileId) async {
    try {
      final response = await _apiClient.get(
        '${ApiEndpoints.files}/$fileId/download',
      );
      final data = response.data as Map<String, dynamic>;
      return data['downloadUrl'] as String;
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> toggleFavorite(String fileId, bool isFavorite) async {
    try {
      if (isFavorite) {
        await _apiClient.put('${ApiEndpoints.files}/$fileId/favorite');
      } else {
        await _apiClient.delete('${ApiEndpoints.files}/$fileId/favorite');
      }
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<FileItemModel> renameFile(String fileId, String name) async {
    try {
      final response = await _apiClient.patch(
        '${ApiEndpoints.files}/$fileId',
        data: {'name': name},
      );
      return FileItemModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> deleteFile(String fileId) async {
    try {
      await _apiClient.delete('${ApiEndpoints.files}/$fileId');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
