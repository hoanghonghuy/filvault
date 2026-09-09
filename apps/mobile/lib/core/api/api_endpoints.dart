/// API Endpoints for Filvault Backend
abstract class ApiEndpoints {
  static const String defaultBaseUrl = 'http://10.0.2.2:8080/api/v1'; // 10.0.2.2 for Android Emulator

  // Auth
  static const String register = '/auth/register';
  static const String login = '/auth/login';
  static const String refresh = '/auth/refresh';
  static const String logout = '/auth/logout';
  static const String verifyEmail = '/auth/verify-email';
  static const String resendVerification = '/auth/resend-verification';
  static const String me = '/users/me';
  static const String changePassword = '/users/me/password';

  // Files & Folders
  static const String browser = '/browser';
  static const String folders = '/folders';
  static const String getOrCreateFolder = '/folders/get-or-create';
  static const String files = '/files';
  static const String uploadSessions = '/files/upload-sessions';
  static const String favorites = '/files/favorites';

  // Photos & Albums
  static const String photoTimeline = '/photos/timeline';
  static const String albums = '/photos/albums';

  // Direct Chat & Realtime
  static const String conversations = '/chat/conversations';
  static const String directConversations = '/chat/direct-conversations';
  static const String chatEvents = '/chat/events';
  static const String chatUploadSessions = '/chat/attachments/upload-sessions';

  // Storage & Trash
  static const String storage = '/storage';
  static const String trash = '/trash';
  static const String activity = '/activity';

  // Public & Sharing
  static const String shares = '/shares';
  static const String sharesWithMe = '/shares/with-me';
  static const String shareLinks = '/share-links';
}
