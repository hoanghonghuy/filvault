class FolderModel {
  final String id;
  final String? parentId;
  final String name;
  final DateTime createdAt;
  final DateTime updatedAt;

  FolderModel({
    required this.id,
    this.parentId,
    required this.name,
    required this.createdAt,
    required this.updatedAt,
  });

  factory FolderModel.fromJson(Map<String, dynamic> json) {
    return FolderModel(
      id: json['id'] as String,
      parentId: json['parentId'] as String?,
      name: json['name'] as String,
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
      updatedAt: DateTime.tryParse(json['updatedAt'] ?? '') ?? DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'parentId': parentId,
    'name': name,
    'createdAt': createdAt.toIso8601String(),
    'updatedAt': updatedAt.toIso8601String(),
  };
}

class FileItemModel {
  final String id;
  final String name;
  final String mimeType;
  final int sizeBytes;
  final DateTime updatedAt;
  final bool isFavorite;

  FileItemModel({
    required this.id,
    required this.name,
    required this.mimeType,
    required this.sizeBytes,
    required this.updatedAt,
    this.isFavorite = false,
  });

  factory FileItemModel.fromJson(Map<String, dynamic> json) {
    return FileItemModel(
      id: json['id'] as String,
      name: json['name'] as String,
      mimeType: json['mimeType'] as String? ?? 'application/octet-stream',
      sizeBytes: (json['sizeBytes'] as num?)?.toInt() ?? 0,
      updatedAt: DateTime.tryParse(json['updatedAt'] ?? '') ?? DateTime.now(),
      isFavorite: json['isFavorite'] as bool? ?? false,
    );
  }

  FileItemModel copyWith({
    String? id,
    String? name,
    String? mimeType,
    int? sizeBytes,
    DateTime? updatedAt,
    bool? isFavorite,
  }) {
    return FileItemModel(
      id: id ?? this.id,
      name: name ?? this.name,
      mimeType: mimeType ?? this.mimeType,
      sizeBytes: sizeBytes ?? this.sizeBytes,
      updatedAt: updatedAt ?? this.updatedAt,
      isFavorite: isFavorite ?? this.isFavorite,
    );
  }

  bool get isImage => mimeType.startsWith('image/');
  bool get isVideo => mimeType.startsWith('video/');
  bool get isAudio => mimeType.startsWith('audio/');
  bool get isPdf => mimeType == 'application/pdf';
  bool get isZip => mimeType.contains('zip') || mimeType.contains('compressed') || mimeType.contains('tar');
  bool get isDocument =>
      mimeType.contains('word') ||
      mimeType.contains('document') ||
      mimeType.contains('sheet') ||
      mimeType.contains('presentation') ||
      mimeType.contains('text/');
}

class BrowserDataModel {
  final FolderModel? currentFolder;
  final List<FolderModel> breadcrumb;
  final List<FolderModel> folders;
  final List<FileItemModel> files;

  BrowserDataModel({
    this.currentFolder,
    required this.breadcrumb,
    required this.folders,
    required this.files,
  });

  factory BrowserDataModel.fromJson(Map<String, dynamic> json) {
    return BrowserDataModel(
      currentFolder: json['folder'] != null
          ? FolderModel.fromJson(json['folder'] as Map<String, dynamic>)
          : null,
      breadcrumb: (json['breadcrumb'] as List<dynamic>?)
              ?.map((e) => FolderModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      folders: (json['folders'] as List<dynamic>?)
              ?.map((e) => FolderModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      files: (json['files'] as List<dynamic>?)
              ?.map((e) => FileItemModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
