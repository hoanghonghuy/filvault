import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';

class CallTokenModel {
  final String token;
  final String room;
  final String url;

  CallTokenModel({
    required this.token,
    required this.room,
    required this.url,
  });

  factory CallTokenModel.fromJson(Map<String, dynamic> json) {
    return CallTokenModel(
      token: json['token'] as String? ?? '',
      room: json['room'] as String? ?? '',
      url: json['url'] as String? ?? 'ws://10.0.2.2:7880',
    );
  }
}

abstract class CallRepository {
  Future<CallTokenModel> getCallToken(String conversationId);
  Future<void> sendCallSignal(
    String conversationId, {
    required String action,
    required bool isVideo,
  });
}

class CallRepositoryImpl implements CallRepository {
  final ApiClient _apiClient;

  CallRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<CallTokenModel> getCallToken(String conversationId) async {
    try {
      final response = await _apiClient.post(
        '${ApiEndpoints.conversations}/$conversationId/call/token',
      );
      return CallTokenModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> sendCallSignal(
    String conversationId, {
    required String action,
    required bool isVideo,
  }) async {
    try {
      await _apiClient.post(
        '${ApiEndpoints.conversations}/$conversationId/call/signal',
        data: {
          'action': action,
          'isVideo': isVideo,
        },
      );
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
