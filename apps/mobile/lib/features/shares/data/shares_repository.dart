import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/share_model.dart';

abstract class SharesRepository {
  Future<List<IncomingShareModel>> getIncomingShares();
  Future<List<OutgoingShareModel>> getOutgoingShares();
  Future<void> revokeShare(String shareId);
  Future<void> createShare({
    required String resourceType,
    required String resourceId,
    required String email,
  });
}

class SharesRepositoryImpl implements SharesRepository {
  final ApiClient _apiClient;

  SharesRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<List<IncomingShareModel>> getIncomingShares() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.sharesWithMe);
      final list = (response.data as List<dynamic>?) ?? [];
      return list.map((e) => IncomingShareModel.fromJson(e as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<List<OutgoingShareModel>> getOutgoingShares() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.shares);
      final list = (response.data as List<dynamic>?) ?? [];
      return list.map((e) => OutgoingShareModel.fromJson(e as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> revokeShare(String shareId) async {
    try {
      await _apiClient.delete('${ApiEndpoints.shares}/$shareId');
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> createShare({
    required String resourceType,
    required String resourceId,
    required String email,
  }) async {
    try {
      await _apiClient.post(
        ApiEndpoints.shares,
        data: {
          'resourceType': resourceType,
          'resourceId': resourceId,
          'email': email,
        },
      );
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
