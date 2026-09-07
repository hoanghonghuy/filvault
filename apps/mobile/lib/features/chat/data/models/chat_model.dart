class ChatPeerModel {
  final String id;
  final String name;
  final String email;
  final String? avatarUrl;
  final DateTime? lastSeenAt;

  ChatPeerModel({
    required this.id,
    required this.name,
    required this.email,
    this.avatarUrl,
    this.lastSeenAt,
  });

  factory ChatPeerModel.fromJson(Map<String, dynamic> json) {
    return ChatPeerModel(
      id: json['id'] as String,
      name: json['name'] as String? ?? 'Người dùng',
      email: json['email'] as String? ?? '',
      avatarUrl: json['avatarUrl'] as String?,
      lastSeenAt: json['lastSeenAt'] != null
          ? DateTime.tryParse(json['lastSeenAt'] as String)
          : null,
    );
  }
}

class ChatPreviewModel {
  final String messageId;
  final String body;
  final DateTime createdAt;
  final List<String> attachments;

  ChatPreviewModel({
    required this.messageId,
    required this.body,
    required this.createdAt,
    this.attachments = const [],
  });

  factory ChatPreviewModel.fromJson(Map<String, dynamic> json) {
    return ChatPreviewModel(
      messageId: json['messageId'] as String? ?? '',
      body: json['body'] as String? ?? '',
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      attachments: (json['attachments'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
    );
  }
}

class ChatConversationModel {
  final String id;
  final String title;
  final String? type;
  final ChatPeerModel? peer;
  final String? peerStatus;
  final DateTime? peerLastSeenAt;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? lastMessageAt;
  final int unreadCount;
  final ChatPreviewModel? preview;

  ChatConversationModel({
    required this.id,
    required this.title,
    this.type,
    this.peer,
    this.peerStatus,
    this.peerLastSeenAt,
    required this.createdAt,
    required this.updatedAt,
    this.lastMessageAt,
    this.unreadCount = 0,
    this.preview,
  });

  factory ChatConversationModel.fromJson(Map<String, dynamic> json) {
    return ChatConversationModel(
      id: json['id'] as String,
      title: json['title'] as String? ?? 'Trò chuyện',
      type: json['type'] as String?,
      peer: json['peer'] != null
          ? ChatPeerModel.fromJson(json['peer'] as Map<String, dynamic>)
          : null,
      peerStatus: json['peerStatus'] as String?,
      peerLastSeenAt: json['peerLastSeenAt'] != null
          ? DateTime.tryParse(json['peerLastSeenAt'] as String)
          : null,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      updatedAt: DateTime.tryParse(json['updatedAt'] ?? '') ?? DateTime.now(),
      lastMessageAt: json['lastMessageAt'] != null
          ? DateTime.tryParse(json['lastMessageAt'] as String)
          : null,
      unreadCount: (json['unreadCount'] as num?)?.toInt() ?? 0,
      preview: json['preview'] != null
          ? ChatPreviewModel.fromJson(json['preview'] as Map<String, dynamic>)
          : null,
    );
  }

  bool get isOnline => peerStatus == 'online';
}

class ChatAttachmentModel {
  final String id;
  final String? fileId;
  final String originalName;
  final String name;
  final String mimeType;
  final int sizeBytes;
  final DateTime createdAt;
  final String? thumbnailUrl;

  ChatAttachmentModel({
    required this.id,
    this.fileId,
    required this.originalName,
    required this.name,
    required this.mimeType,
    required this.sizeBytes,
    required this.createdAt,
    this.thumbnailUrl,
  });

  factory ChatAttachmentModel.fromJson(Map<String, dynamic> json) {
    return ChatAttachmentModel(
      id: json['id'] as String,
      fileId: json['fileId'] as String?,
      originalName: json['originalName'] as String? ?? json['name'] as String? ?? 'Tệp',
      name: json['name'] as String? ?? 'Tệp',
      mimeType: json['mimeType'] as String? ?? 'application/octet-stream',
      sizeBytes: (json['sizeBytes'] as num?)?.toInt() ?? 0,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      thumbnailUrl: json['thumbnailUrl'] as String?,
    );
  }

  bool get isImage => mimeType.startsWith('image/');
}

class ChatMessageReactionModel {
  final String reaction;
  final int count;
  final List<String> userIds;
  final bool reacted;

  ChatMessageReactionModel({
    required this.reaction,
    required this.count,
    required this.userIds,
    required this.reacted,
  });

  factory ChatMessageReactionModel.fromJson(Map<String, dynamic> json) {
    return ChatMessageReactionModel(
      reaction: json['reaction'] as String,
      count: (json['count'] as num?)?.toInt() ?? 0,
      userIds: (json['userIds'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? [],
      reacted: json['reacted'] as bool? ?? false,
    );
  }
}

class ChatMessageModel {
  final String id;
  final String conversationId;
  final String body;
  final String senderId;
  final String? clientMessageId;
  final DateTime? editedAt;
  final DateTime? removedAt;
  final DateTime createdAt;
  final List<ChatAttachmentModel> attachments;
  final List<ChatMessageReactionModel> reactions;

  ChatMessageModel({
    required this.id,
    required this.conversationId,
    required this.body,
    required this.senderId,
    this.clientMessageId,
    this.editedAt,
    this.removedAt,
    required this.createdAt,
    this.attachments = const [],
    this.reactions = const [],
  });

  factory ChatMessageModel.fromJson(Map<String, dynamic> json) {
    return ChatMessageModel(
      id: json['id'] as String,
      conversationId: json['conversationId'] as String,
      body: json['body'] as String? ?? '',
      senderId: json['senderId'] as String,
      clientMessageId: json['clientMessageId'] as String?,
      editedAt: json['editedAt'] != null ? DateTime.tryParse(json['editedAt'] as String) : null,
      removedAt: json['removedAt'] != null ? DateTime.tryParse(json['removedAt'] as String) : null,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      attachments: (json['attachments'] as List<dynamic>?)
              ?.map((e) => ChatAttachmentModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      reactions: (json['reactions'] as List<dynamic>?)
              ?.map((e) => ChatMessageReactionModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  ChatMessageModel copyWith({
    String? id,
    String? conversationId,
    String? body,
    String? senderId,
    String? clientMessageId,
    DateTime? editedAt,
    DateTime? removedAt,
    DateTime? createdAt,
    List<ChatAttachmentModel>? attachments,
    List<ChatMessageReactionModel>? reactions,
  }) {
    return ChatMessageModel(
      id: id ?? this.id,
      conversationId: conversationId ?? this.conversationId,
      body: body ?? this.body,
      senderId: senderId ?? this.senderId,
      clientMessageId: clientMessageId ?? this.clientMessageId,
      editedAt: editedAt ?? this.editedAt,
      removedAt: removedAt ?? this.removedAt,
      createdAt: createdAt ?? this.createdAt,
      attachments: attachments ?? this.attachments,
      reactions: reactions ?? this.reactions,
    );
  }
}
