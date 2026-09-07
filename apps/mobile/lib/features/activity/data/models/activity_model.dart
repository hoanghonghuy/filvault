class ActivityEventModel {
  final String id;
  final String type;
  final String targetName;
  final DateTime createdAt;

  ActivityEventModel({
    required this.id,
    required this.type,
    required this.targetName,
    required this.createdAt,
  });

  factory ActivityEventModel.fromJson(Map<String, dynamic> json) {
    return ActivityEventModel(
      id: json['id'] as String,
      type: json['type'] as String? ?? '',
      targetName: json['targetName'] as String? ?? '',
      createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
    );
  }
}

class ActivityPageModel {
  final List<ActivityEventModel> events;
  final String? nextBefore;

  ActivityPageModel({
    required this.events,
    this.nextBefore,
  });

  factory ActivityPageModel.fromJson(Map<String, dynamic> json) {
    return ActivityPageModel(
      events: (json['events'] as List<dynamic>?)
              ?.map((e) => ActivityEventModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      nextBefore: json['nextBefore'] as String?,
    );
  }
}
