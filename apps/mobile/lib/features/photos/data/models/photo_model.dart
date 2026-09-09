class TimelineItemModel {
  final String id;
  final String name;
  final String mimeType;
  final int sizeBytes;
  final DateTime createdAt;
  final String? thumbnailUrl;

  TimelineItemModel({
    required this.id,
    required this.name,
    required this.mimeType,
    required this.sizeBytes,
    required this.createdAt,
    this.thumbnailUrl,
  });

  factory TimelineItemModel.fromJson(Map<String, dynamic> json) {
    return TimelineItemModel(
      id: json['id'] as String,
      name: json['name'] as String,
      mimeType: json['mimeType'] as String? ?? 'image/jpeg',
      sizeBytes: (json['sizeBytes'] as num?)?.toInt() ?? 0,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      thumbnailUrl: json['thumbnailUrl'] as String?,
    );
  }

  bool get isVideo => mimeType.startsWith('video/');
}

class TimelineGroupModel {
  final String date;
  final List<TimelineItemModel> items;

  TimelineGroupModel({
    required this.date,
    required this.items,
  });

  factory TimelineGroupModel.fromJson(Map<String, dynamic> json) {
    return TimelineGroupModel(
      date: json['date'] as String,
      items: (json['items'] as List<dynamic>?)
              ?.map((e) => TimelineItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class TimelineDataModel {
  final List<TimelineGroupModel> groups;
  final String? nextBefore;

  TimelineDataModel({
    required this.groups,
    this.nextBefore,
  });

  factory TimelineDataModel.fromJson(Map<String, dynamic> json) {
    return TimelineDataModel(
      groups: (json['groups'] as List<dynamic>?)
              ?.map((e) => TimelineGroupModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      nextBefore: json['nextBefore'] as String?,
    );
  }
}

class AlbumModel {
  final String id;
  final String name;
  final int itemCount;
  final String? coverFileId;
  final String? coverUrl;
  final DateTime createdAt;
  final DateTime updatedAt;

  AlbumModel({
    required this.id,
    required this.name,
    required this.itemCount,
    this.coverFileId,
    this.coverUrl,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AlbumModel.fromJson(Map<String, dynamic> json) {
    return AlbumModel(
      id: json['id'] as String,
      name: json['name'] as String,
      itemCount: (json['itemCount'] as num?)?.toInt() ?? 0,
      coverFileId: json['coverFileId'] as String?,
      coverUrl: json['coverUrl'] as String?,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      updatedAt: DateTime.tryParse(json['updatedAt'] ?? '') ?? DateTime.now(),
    );
  }
}

class AlbumDetailModel {
  final AlbumModel album;
  final List<TimelineItemModel> items;

  AlbumDetailModel({
    required this.album,
    required this.items,
  });

  factory AlbumDetailModel.fromJson(Map<String, dynamic> json) {
    return AlbumDetailModel(
      album: AlbumModel.fromJson(json),
      items: (json['items'] as List<dynamic>?)
              ?.map((e) => TimelineItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
