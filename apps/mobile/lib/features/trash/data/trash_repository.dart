import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/trash_model.dart';

abstract class TrashRepository {
  Future<TrashListModel> getTrash();
  Future<void> restoreFile(String fileId);
  Future<void> restoreFolder(String folderId);
  Future<void> permanentDelete({required String type, required String id});
}

class TrashRepositoryImpl implements TrashRepository {
  final ApiClient _apiClient;

  TrashRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<TrashListModel> getTrash() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.trash);
      return TrashListModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> restoreFile(String fileId) async {
    try {
      await _apiClient.post('${ApiEndpoints.files}/$fileId/restore');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> restoreFolder(String folderId) async {
    try {
      await _apiClient.post('${ApiEndpoints.folders}/$folderId/restore');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> permanentDelete({required String type, required String id}) async {
    try {
      await _apiClient.delete('${ApiEndpoints.trash}/$type/$id');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
