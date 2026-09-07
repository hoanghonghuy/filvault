import 'package:dio/dio.dart';
import '../storage/secure_storage_service.dart';
import 'api_endpoints.dart';

class AuthInterceptor extends QueuedInterceptor {
  final SecureStorageService _storage;
  final Dio _dio;
  final void Function()? _onSessionExpired;

  AuthInterceptor({
    required SecureStorageService storage,
    required Dio dio,
    void Function()? onSessionExpired,
  })  : _storage = storage,
        _dio = dio,
        _onSessionExpired = onSessionExpired;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // Exclude public/auth endpoints from auto-attaching bearer token if not needed
    final token = await _storage.getAccessToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final response = err.response;
    final path = err.requestOptions.path;

    // Only attempt token refresh on 401 and when not already on an auth endpoint
    final isAuthEndpoint = path.contains(ApiEndpoints.login) ||
        path.contains(ApiEndpoints.refresh) ||
        path.contains(ApiEndpoints.register);

    if (response?.statusCode == 401 && !isAuthEndpoint) {
      final refreshToken = await _storage.getRefreshToken();
      if (refreshToken != null && refreshToken.isNotEmpty) {
        try {
          // Use clean Dio options to avoid infinite interceptor loops
          final refreshDio = Dio(BaseOptions(
            baseUrl: _dio.options.baseUrl,
            headers: {'Content-Type': 'application/json'},
          ));

          final refreshResponse = await refreshDio.post(
            ApiEndpoints.refresh,
            data: {'refreshToken': refreshToken},
          );

          if (refreshResponse.statusCode == 200) {
            final data = refreshResponse.data as Map<String, dynamic>;
            final newAccessToken = data['accessToken'] as String;
            final newRefreshToken = data['refreshToken'] as String;

            await _storage.saveTokens(
              accessToken: newAccessToken,
              refreshToken: newRefreshToken,
            );

            // Retry original request with newly obtained access token
            final originalOptions = err.requestOptions;
            originalOptions.headers['Authorization'] = 'Bearer $newAccessToken';

            final retryResponse = await _dio.fetch(originalOptions);
            return handler.resolve(retryResponse);
          }
        } catch (_) {
          // Refresh failed - session has truly expired
          await _storage.clearAll();
          _onSessionExpired?.call();
        }
      } else {
        await _storage.clearAll();
        _onSessionExpired?.call();
      }
    }

    handler.next(err);
  }
}
