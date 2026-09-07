import 'user_model.dart';

class SessionModel {
  final UserModel user;
  final String accessToken;
  final String refreshToken;

  const SessionModel({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });

  factory SessionModel.fromJson(Map<String, dynamic> json) {
    return SessionModel(
      user: UserModel.fromJson(json['user'] as Map<String, dynamic>),
      accessToken: json['accessToken'] as String,
      refreshToken: json['refreshToken'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user': user.toJson(),
      'accessToken': accessToken,
      'refreshToken': refreshToken,
    };
  }
}
