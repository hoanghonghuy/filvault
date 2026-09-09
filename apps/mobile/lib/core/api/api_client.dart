import 'package:dio/dio.dart';
import '../storage/secure_storage_service.dart';
import 'api_endpoints.dart';
import 'auth_interceptor.dart';

class ApiClient {
  late final Dio dio;
  final SecureStorageService storage;

  ApiClient({
    required this.storage,
    String? baseUrl,
    void Function()? onSessionExpired,
  }) {
    dio = Dio(
      BaseOptions(
        baseUrl: baseUrl ?? ApiEndpoints.defaultBaseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 30),
        sendTimeout: const Duration(minutes: 5), // for uploads
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    dio.interceptors.add(
      AuthInterceptor(
        storage: storage,
        dio: dio,
        onSessionExpired: onSessionExpired,
      ),
    );
  }

  /// Upload file stream or bytes directly to S3 via pre-signed PUT URL
  Future<Response<dynamic>> uploadToS3Direct({
    required String uploadUrl,
    required dynamic data,
    required String contentType,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  }) async {
    // Clean dio instance with no auth header or default baseUrl for S3
    final s3Dio = Dio();
    return s3Dio.put(
      uploadUrl,
      data: data,
      options: Options(
        headers: {
          'Content-Type': contentType,
        },
      ),
      onSendProgress: onProgress,
      cancelToken: cancelToken,
    );
  }

  /// Format backend error body into friendly human readable message
  static String formatError(dynamic error) {
    if (error is DioException) {
      final responseData = error.response?.data;
      if (responseData is Map<String, dynamic>) {
        final errObj = responseData['error'];
        if (errObj is Map<String, dynamic> && errObj['message'] != null) {
          return errObj['message'].toString();
        }
      }
      switch (error.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          return 'Kết nối mạng quá thời gian chờ. Vui lòng thử lại.';
        case DioExceptionType.connectionError:
          return 'Không thể kết nối đến máy chủ. Vui lòng kiểm tra mạng.';
        default:
          break;
      }
    }
    return error.toString();
  }
}
