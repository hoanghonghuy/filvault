class ShareUserRefModel {
  final String id;
  final String email;
  final String displayName;

  ShareUserRefModel({
    required this.id,
    required this.email,
    required this.displayName,
  });

  factory ShareUserRefModel.fromJson(Map<String, dynamic> json) {
    return ShareUserRefModel(
      id: json['id'] as String? ?? '',
      email: json['email'] as String? ?? '',
      displayName: json['displayName'] as String? ?? json['email'] as String? ?? '',
    );
  }
}

class IncomingShareModel {
  final String id;
  final String resourceType;
  final String resourceId;
  final String resourceName;
  final ShareUserRefModel owner;
  final DateTime createdAt;

  IncomingShareModel({
    required this.id,
    required this.resourceType,
    required this.resourceId,
    required this.resourceName,
    required this.owner,
    required this.createdAt,
  });

  factory IncomingShareModel.fromJson(Map<String, dynamic> json) {
    return IncomingShareModel(
      id: json['id'] as String,
      resourceType: json['resourceType'] as String? ?? 'file',
      resourceId: json['resourceId'] as String,
      resourceName: json['resourceName'] as String? ?? 'Mục chia sẻ',
      owner: ShareUserRefModel.fromJson(json['owner'] as Map<String, dynamic>? ?? {}),
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
    );
  }

  bool get isFolder => resourceType == 'folder';
}

class OutgoingShareModel {
  final String id;
  final String resourceType;
  final String resourceId;
  final String resourceName;
  final ShareUserRefModel recipient;
  final DateTime createdAt;

  OutgoingShareModel({
    required this.id,
    required this.resourceType,
    required this.resourceId,
    required this.resourceName,
    required this.recipient,
    required this.createdAt,
  });

  factory OutgoingShareModel.fromJson(Map<String, dynamic> json) {
    return OutgoingShareModel(
      id: json['id'] as String,
      resourceType: json['resourceType'] as String? ?? 'file',
      resourceId: json['resourceId'] as String,
      resourceName: json['resourceName'] as String? ?? 'Mục chia sẻ',
      recipient: ShareUserRefModel.fromJson(json['recipient'] as Map<String, dynamic>? ?? {}),
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
    );
  }

  bool get isFolder => resourceType == 'folder';
}
