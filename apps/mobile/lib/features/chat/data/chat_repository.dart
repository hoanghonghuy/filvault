import 'package:dio/dio.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/api/api_endpoints.dart';
import 'models/chat_model.dart';

abstract class ChatRepository {
  Future<List<ChatConversationModel>> getConversations();
  Future<ChatConversationModel> createDirectConversation(String targetUserId);
  Future<List<ChatMessageModel>> getMessages(String conversationId, {String? before, int limit = 50});
  Future<ChatMessageModel> sendMessage(
    String conversationId, {
    required String content,
    String? replyToMessageId,
    List<String>? attachmentFileIds,
  });
  Future<ChatMessageModel> editMessage(String conversationId, String messageId, String content);
  Future<void> deleteMessage(String conversationId, String messageId);
  Future<void> toggleReaction(String conversationId, String messageId, String emoji);
  Future<void> markAsRead(String conversationId);
  Future<void> sendTyping(String conversationId);
  Future<Map<String, dynamic>> createAttachmentSession({
    required String name,
    required int size,
    required String contentType,
  });
  Future<ChatAttachmentModel> completeAttachment(String fileId);
}

class ChatRepositoryImpl implements ChatRepository {
  final ApiClient _apiClient;

  ChatRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient.instance;

  @override
  Future<List<ChatConversationModel>> getConversations() async {
    try {
      final response = await _apiClient.get(ApiEndpoints.conversations);
      final list = (response.data['conversations'] as List<dynamic>?) ?? [];
      return list.map((e) => ChatConversationModel.fromJson(e as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<ChatConversationModel> createDirectConversation(String targetUserId) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.directConversations,
        data: {'targetUserId': targetUserId},
      );
      return ChatConversationModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<List<ChatMessageModel>> getMessages(
    String conversationId, {
    String? before,
    int limit = 50,
  }) async {
    try {
      final query = <String, dynamic>{'limit': limit};
      if (before != null) query['before'] = before;

      final response = await _apiClient.get(
        '${ApiEndpoints.conversations}/$conversationId/messages',
        queryParameters: query,
      );
      final list = (response.data['messages'] as List<dynamic>?) ?? [];
      return list.map((e) => ChatMessageModel.fromJson(e as Map<String, dynamic>)).toList();
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<ChatMessageModel> sendMessage(
    String conversationId, {
    required String content,
    String? replyToMessageId,
    List<String>? attachmentFileIds,
  }) async {
    try {
      final response = await _apiClient.post(
        '${ApiEndpoints.conversations}/$conversationId/messages',
        data: {
          'body': content,
          if (replyToMessageId != null) 'replyToMessageId': replyToMessageId,
          if (attachmentFileIds != null && attachmentFileIds.isNotEmpty)
            'attachmentFileIds': attachmentFileIds,
        },
      );
      return ChatMessageModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<ChatMessageModel> editMessage(
    String conversationId,
    String messageId,
    String content,
  ) async {
    try {
      final response = await _apiClient.patch(
        '${ApiEndpoints.conversations}/$conversationId/messages/$messageId',
        data: {'body': content},
      );
      return ChatMessageModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> deleteMessage(String conversationId, String messageId) async {
    try {
      await _apiClient.delete(
        '${ApiEndpoints.conversations}/$conversationId/messages/$messageId',
      );
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> toggleReaction(String conversationId, String messageId, String emoji) async {
    try {
      await _apiClient.post(
        '${ApiEndpoints.conversations}/$conversationId/messages/$messageId/reactions',
        data: {'emoji': emoji},
      );
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<void> markAsRead(String conversationId) async {
    try {
      await _apiClient.post('${ApiEndpoints.conversations}/$conversationId/read');
    } catch (_) {}
  }

  @override
  Future<void> sendTyping(String conversationId) async {
    try {
      await _apiClient.post('${ApiEndpoints.conversations}/$conversationId/typing');
    } catch (_) {}
  }

  @override
  Future<Map<String, dynamic>> createAttachmentSession({
    required String name,
    required int size,
    required String contentType,
  }) async {
    try {
      final response = await _apiClient.post(
        ApiEndpoints.chatUploadSessions,
        data: {
          'name': name,
          'size': size,
          'contentType': contentType,
        },
      );
      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }

  @override
  Future<ChatAttachmentModel> completeAttachment(String fileId) async {
    try {
      final response = await _apiClient.post(
        '/chat/attachments/$fileId/complete',
      );
      return ChatAttachmentModel.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw Exception(ApiClient.formatError(e));
    }
  }
}
