import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/models/user_model.dart';
import '../../../../core/storage/secure_storage_service.dart';
import '../../data/auth_repository.dart';

// Providers
final secureStorageProvider = Provider<SecureStorageService>((ref) {
  return SecureStorageService();
});

final apiClientProvider = Provider<ApiClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return ApiClient(
    storage: storage,
    onSessionExpired: () {
      ref.read(authControllerProvider.notifier).handleSessionExpired();
    },
  );
});

final authRepositoryProvider = Provider<IAuthRepository>((ref) {
  return AuthRepository(
    apiClient: ref.watch(apiClientProvider),
    storage: ref.watch(secureStorageProvider),
  );
});

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthState>((ref) {
  return AuthController(
    authRepo: ref.watch(authRepositoryProvider),
    storage: ref.watch(secureStorageProvider),
  );
});

// State definitions
enum AuthStatus {
  initial,
  loading,
  authenticated,
  unverified,
  unauthenticated,
}

class AuthState {
  final AuthStatus status;
  final UserModel? user;
  final String? errorMessage;

  const AuthState({
    required this.status,
    this.user,
    this.errorMessage,
  });

  factory AuthState.initial() => const AuthState(status: AuthStatus.initial);
  factory AuthState.loading() => const AuthState(status: AuthStatus.loading);
  factory AuthState.authenticated(UserModel user) =>
      AuthState(status: AuthStatus.authenticated, user: user);
  factory AuthState.unverified(UserModel user) =>
      AuthState(status: AuthStatus.unverified, user: user);
  factory AuthState.unauthenticated([String? message]) =>
      AuthState(status: AuthStatus.unauthenticated, errorMessage: message);

  bool get isAuthenticated => status == AuthStatus.authenticated;
  bool get isUnverified => status == AuthStatus.unverified;
  bool get isLoading => status == AuthStatus.loading;
}

class AuthController extends StateNotifier<AuthState> {
  final IAuthRepository _authRepo;
  final SecureStorageService _storage;

  AuthController({
    required IAuthRepository authRepo,
    required SecureStorageService storage,
  })  : _authRepo = authRepo,
        _storage = storage,
        super(AuthState.initial());

  Future<void> checkSession() async {
    state = AuthState.loading();
    final token = await _storage.getAccessToken();
    if (token == null || token.isEmpty) {
      state = AuthState.unauthenticated();
      return;
    }
    try {
      final user = await _authRepo.getMe();
      if (!user.emailVerified) {
        state = AuthState.unverified(user);
      } else {
        state = AuthState.authenticated(user);
      }
    } catch (_) {
      await _storage.clearAll();
      state = AuthState.unauthenticated();
    }
  }

  Future<bool> login({
    required String email,
    required String password,
  }) async {
    state = AuthState.loading();
    try {
      final session = await _authRepo.login(email: email, password: password);
      if (!session.user.emailVerified) {
        state = AuthState.unverified(session.user);
      } else {
        state = AuthState.authenticated(session.user);
      }
      return true;
    } catch (e) {
      final errorMsg = ApiClient.formatError(e);
      state = AuthState.unauthenticated(errorMsg);
      return false;
    }
  }

  Future<bool> register({
    required String email,
    required String password,
    String? displayName,
  }) async {
    state = AuthState.loading();
    try {
      final session = await _authRepo.register(
        email: email,
        password: password,
        displayName: displayName,
      );
      state = AuthState.unverified(session.user);
      return true;
    } catch (e) {
      final errorMsg = ApiClient.formatError(e);
      state = AuthState.unauthenticated(errorMsg);
      return false;
    }
  }

  Future<bool> verifyEmail(String code) async {
    state = AuthState.loading();
    try {
      final user = await _authRepo.verifyEmail(code);
      state = AuthState.authenticated(user);
      return true;
    } catch (e) {
      final errorMsg = ApiClient.formatError(e);
      final currentUser = state.user;
      if (currentUser != null) {
        state = AuthState.unverified(currentUser);
      } else {
        state = AuthState.unauthenticated(errorMsg);
      }
      return false;
    }
  }

  Future<void> resendVerification() async {
    try {
      await _authRepo.resendVerification();
    } catch (_) {}
  }

  Future<void> logout() async {
    state = AuthState.loading();
    await _authRepo.logout();
    state = AuthState.unauthenticated();
  }

  void handleSessionExpired() {
    state = AuthState.unauthenticated('Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.');
  }
}
