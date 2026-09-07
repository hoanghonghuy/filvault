class TrashItemModel {
  final String id;
  final String name;
  final DateTime deletedAt;
  final String? mimeType;
  final int sizeBytes;

  TrashItemModel({
    required this.id,
    required this.name,
    required this.deletedAt,
    this.mimeType,
    this.sizeBytes = 0,
  });

  factory TrashItemModel.fromJson(Map<String, dynamic> json) {
    return TrashItemModel(
      id: json['id'] as String,
      name: json['name'] as String,
      deletedAt: DateTime.tryParse(json['deletedAt'] ?? '') ?? DateTime.now(),
      mimeType: json['mimeType'] as String?,
      sizeBytes: (json['sizeBytes'] as num?)?.toInt() ?? 0,
    );
  }
}

class TrashListModel {
  final List<TrashItemModel> folders;
  final List<TrashItemModel> files;

  TrashListModel({
    required this.folders,
    required this.files,
  });

  factory TrashListModel.fromJson(Map<String, dynamic> json) {
    return TrashListModel(
      folders: (json['folders'] as List<dynamic>?)
              ?.map((e) => TrashItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      files: (json['files'] as List<dynamic>?)
              ?.map((e) => TrashItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
