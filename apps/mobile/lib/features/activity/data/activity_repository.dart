import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/activity_model.dart';

abstract class ActivityRepository {
  Future<ActivityPageModel> getActivity({String? before, int limit = 50});
}

class ActivityRepositoryImpl implements ActivityRepository {
  final ApiClient _apiClient;

  ActivityRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<ActivityPageModel> getActivity({String? before, int limit = 50}) async {
    try {
      final query = <String, dynamic>{'limit': limit};
      if (before != null) query['before'] = before;

      final response = await _apiClient.get(
        ApiEndpoints.activity,
        queryParameters: query,
      );
      return ActivityPageModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
