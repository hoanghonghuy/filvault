import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/photo_model.dart';

abstract class PhotosRepository {
  Future<TimelineDataModel> getTimeline({String? before, int limit = 50});
  Future<List<AlbumModel>> getAlbums();
  Future<AlbumModel> createAlbum(String name);
  Future<AlbumDetailModel> getAlbumDetail(String albumId);
  Future<void> addItemsToAlbum(String albumId, List<String> fileIds);
  Future<void> removeItemFromAlbum(String albumId, String fileId);
  Future<void> deleteAlbum(String albumId);
  Future<String> getDownloadUrl(String fileId);
}

class PhotosRepositoryImpl implements PhotosRepository {
  final ApiClient _apiClient;

  PhotosRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<TimelineDataModel> getTimeline({String? before, int limit = 50}) async {
    try {
      final query = <String, dynamic>{'limit': limit};
      if (before != null) query['before'] = before;

      final response = await _apiClient.get(
        ApiEndpoints.photoTimeline,
        queryParameters: query,
      );
      return TimelineDataModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<List<AlbumModel>> getAlbums() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.albums);
      final list = (response.data['albums'] as List<dynamic>?) ?? [];
      return list.map((e) => AlbumModel.fromJson(e as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<AlbumModel> createAlbum(String name) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.albums,
        data: {'name': name},
      );
      return AlbumModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<AlbumDetailModel> getAlbumDetail(String albumId) async {
    try {
      final response = await _apiClient.get('${ApiEndpoints.albums}/$albumId');
      return AlbumDetailModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> addItemsToAlbum(String albumId, List<String> fileIds) async {
    try {
      await _apiClient.post(
        '${ApiEndpoints.albums}/$albumId/items',
        data: {'fileIds': fileIds},
      );
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> removeItemFromAlbum(String albumId, String fileId) async {
    try {
      await _apiClient.delete('${ApiEndpoints.albums}/$albumId/items/$fileId');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> deleteAlbum(String albumId) async {
    try {
      await _apiClient.delete('${ApiEndpoints.albums}/$albumId');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<String> getDownloadUrl(String fileId) async {
    try {
      final response = await _apiClient.get('${ApiEndpoints.files}/$fileId/download');
      return response.data['downloadUrl'] as String;
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
