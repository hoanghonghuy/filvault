import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import '../../../../core/models/session_model.dart';
import '../../../../core/models/user_model.dart';
import '../../../../core/storage/secure_storage_service.dart';

abstract class IAuthRepository {
  Future<SessionModel> login({required String email, required String password});
  Future<SessionModel> register({required String email, required String password, String? displayName});
  Future<UserModel> verifyEmail(String code);
  Future<void> resendVerification();
  Future<UserModel> getMe();
  Future<void> logout();
  Future<void> changePassword({required String oldPassword, required String newPassword});
  Future<UserModel> updateProfile({String? displayName, String? avatarUrl});
}

class AuthRepository implements IAuthRepository {
  final ApiClient _apiClient;
  final SecureStorageService _storage;

  AuthRepository({
    required ApiClient apiClient,
    required SecureStorageService storage,
  })  : _apiClient = apiClient,
        _storage = storage;

  @override
  Future<SessionModel> login({
    required String email,
    required String password,
  }) async {
    final response = await _apiClient.dio.post(
      ApiEndpoints.login,
      data: {
        'email': email.trim(),
        'password': password,
      },
    );
    final session = SessionModel.fromJson(response.data as Map<String, dynamic>);
    await _storage.saveTokens(
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
      userId: session.user.id,
    );
    return session;
  }

  @override
  Future<SessionModel> register({
    required String email,
    required String password,
    String? displayName,
  }) async {
    final response = await _apiClient.dio.post(
      ApiEndpoints.register,
      data: {
        'email': email.trim(),
        'password': password,
        if (displayName != null && displayName.trim().isNotEmpty)
          'displayName': displayName.trim(),
      },
    );
    final session = SessionModel.fromJson(response.data as Map<String, dynamic>);
    await _storage.saveTokens(
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
      userId: session.user.id,
    );
    return session;
  }

  @override
  Future<UserModel> verifyEmail(String code) async {
    final response = await _apiClient.dio.post(
      ApiEndpoints.verifyEmail,
      data: {'code': code.trim()},
    );
    return UserModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<void> resendVerification() async {
    await _apiClient.dio.post(ApiEndpoints.resendVerification);
  }

  @override
  Future<UserModel> getMe() async {
    final response = await _apiClient.dio.get(ApiEndpoints.me);
    return UserModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<void> logout() async {
    try {
      await _apiClient.dio.post(ApiEndpoints.logout);
    } catch (_) {
      // Ignore network errors on logout
    } finally {
      await _storage.clearAll();
    }
  }

  @override
  Future<void> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    await _apiClient.dio.post(
      ApiEndpoints.changePassword,
      data: {
        'oldPassword': oldPassword,
        'newPassword': newPassword,
      },
    );
  }

  @override
  Future<UserModel> updateProfile({
    String? displayName,
    String? avatarUrl,
  }) async {
    final payload = <String, dynamic>{};
    if (displayName != null) payload['displayName'] = displayName.trim();
    if (avatarUrl != null) payload['avatarUrl'] = avatarUrl;

    final response = await _apiClient.dio.patch(
      ApiEndpoints.me,
      data: payload,
    );
    return UserModel.fromJson(response.data as Map<String, dynamic>);
  }
}
