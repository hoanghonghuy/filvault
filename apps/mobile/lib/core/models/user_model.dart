class UserModel {
  final String id;
  final String email;
  final String displayName;
  final bool emailVerified;
  final int storageUsed;
  final int storageQuota;
  final bool imageThumbnailsEnabled;
  final bool videoThumbnailsEnabled;
  final bool trashAutoDeleteEnabled;
  final int trashRetentionDays;
  final bool? activeStatusEnabled;
  final String? avatarUrl;
  final DateTime createdAt;

  const UserModel({
    required this.id,
    required this.email,
    required this.displayName,
    required this.emailVerified,
    required this.storageUsed,
    required this.storageQuota,
    required this.imageThumbnailsEnabled,
    required this.videoThumbnailsEnabled,
    required this.trashAutoDeleteEnabled,
    required this.trashRetentionDays,
    this.activeStatusEnabled,
    this.avatarUrl,
    required this.createdAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] as String,
      email: json['email'] as String,
      displayName: json['displayName'] as String? ?? '',
      emailVerified: json['emailVerified'] as bool? ?? false,
      storageUsed: (json['storageUsed'] as num?)?.toInt() ?? 0,
      storageQuota: (json['storageQuota'] as num?)?.toInt() ?? 0,
      imageThumbnailsEnabled: json['imageThumbnailsEnabled'] as bool? ?? true,
      videoThumbnailsEnabled: json['videoThumbnailsEnabled'] as bool? ?? true,
      trashAutoDeleteEnabled: json['trashAutoDeleteEnabled'] as bool? ?? false,
      trashRetentionDays: (json['trashRetentionDays'] as num?)?.toInt() ?? 30,
      activeStatusEnabled: json['activeStatusEnabled'] as bool?,
      avatarUrl: json['avatarUrl'] as String?,
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'email': email,
      'displayName': displayName,
      'emailVerified': emailVerified,
      'storageUsed': storageUsed,
      'storageQuota': storageQuota,
      'imageThumbnailsEnabled': imageThumbnailsEnabled,
      'videoThumbnailsEnabled': videoThumbnailsEnabled,
      'trashAutoDeleteEnabled': trashAutoDeleteEnabled,
      'trashRetentionDays': trashRetentionDays,
      'activeStatusEnabled': activeStatusEnabled,
      'avatarUrl': avatarUrl,
      'createdAt': createdAt.toIso8601String(),
    };
  }

  UserModel copyWith({
    String? id,
    String? email,
    String? displayName,
    bool? emailVerified,
    int? storageUsed,
    int? storageQuota,
    bool? imageThumbnailsEnabled,
    bool? videoThumbnailsEnabled,
    bool? trashAutoDeleteEnabled,
    int? trashRetentionDays,
    bool? activeStatusEnabled,
    String? avatarUrl,
    DateTime? createdAt,
  }) {
    return UserModel(
      id: id ?? this.id,
      email: email ?? this.email,
      displayName: displayName ?? this.displayName,
      emailVerified: emailVerified ?? this.emailVerified,
      storageUsed: storageUsed ?? this.storageUsed,
      storageQuota: storageQuota ?? this.storageQuota,
      imageThumbnailsEnabled: imageThumbnailsEnabled ?? this.imageThumbnailsEnabled,
      videoThumbnailsEnabled: videoThumbnailsEnabled ?? this.videoThumbnailsEnabled,
      trashAutoDeleteEnabled: trashAutoDeleteEnabled ?? this.trashAutoDeleteEnabled,
      trashRetentionDays: trashRetentionDays ?? this.trashRetentionDays,
      activeStatusEnabled: activeStatusEnabled ?? this.activeStatusEnabled,
      avatarUrl: avatarUrl ?? this.avatarUrl,
      createdAt: createdAt ?? this.createdAt,
    );
  }
}
