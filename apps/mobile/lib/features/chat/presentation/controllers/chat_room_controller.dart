import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/chat_repository.dart';
import '../../data/models/chat_model.dart';
import 'chat_list_controller.dart';

class ChatRoomState {
  final bool isLoading;
  final bool isSending;
  final String? errorMessage;
  final List<ChatMessageModel> messages;
  final bool isPeerTyping;
  final ChatMessageModel? replyingTo;

  const ChatRoomState({
    this.isLoading = false,
    this.isSending = false,
    this.errorMessage,
    this.messages = const [],
    this.isPeerTyping = false,
    this.replyingTo,
  });

  ChatRoomState copyWith({
    bool? isLoading,
    bool? isSending,
    String? errorMessage,
    List<ChatMessageModel>? messages,
    bool? isPeerTyping,
    ChatMessageModel? replyingTo,
    bool clearReply = false,
  }) {
    return ChatRoomState(
      isLoading: isLoading ?? this.isLoading,
      isSending: isSending ?? this.isSending,
      errorMessage: errorMessage,
      messages: messages ?? this.messages,
      isPeerTyping: isPeerTyping ?? this.isPeerTyping,
      replyingTo: clearReply ? null : (replyingTo ?? this.replyingTo),
    );
  }
}

class ChatRoomController extends StateNotifier<ChatRoomState> {
  final ChatRepository _repo;
  final String conversationId;
  Timer? _typingDebounce;

  ChatRoomController(this._repo, this.conversationId)
      : super(const ChatRoomState()) {
    loadMessages();
    _markRead();
  }

  @override
  void dispose() {
    _typingDebounce?.cancel();
    super.dispose();
  }

  Future<void> loadMessages() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final list = await _repo.getMessages(conversationId, limit: 50);
      state = state.copyWith(isLoading: false, messages: list);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  void setReplyingTo(ChatMessageModel? message) {
    state = state.copyWith(replyingTo: message, clearReply: message == null);
  }

  void onUserTyping() {
    if (_typingDebounce?.isActive ?? false) return;
    _repo.sendTyping(conversationId);
    _typingDebounce = Timer(const Duration(seconds: 3), () {});
  }

  Future<bool> sendMessage(String content, {List<String>? attachmentFileIds}) async {
    final text = content.trim();
    if (text.isEmpty && (attachmentFileIds == null || attachmentFileIds.isEmpty)) {
      return false;
    }

    state = state.copyWith(isSending: true);
    try {
      final newMsg = await _repo.sendMessage(
        conversationId,
        content: text,
        replyToMessageId: state.replyingTo?.id,
        attachmentFileIds: attachmentFileIds,
      );

      state = state.copyWith(
        isSending: false,
        messages: [newMsg, ...state.messages],
        clearReply: true,
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        isSending: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<void> toggleReaction(String messageId, String emoji) async {
    try {
      await _repo.toggleReaction(conversationId, messageId, emoji);
      // Reload message list or update locally
      await loadMessages();
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> deleteMessage(String messageId) async {
    try {
      await _repo.deleteMessage(conversationId, messageId);
      state = state.copyWith(
        messages: state.messages.where((m) => m.id != messageId).toList(),
      );
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  void _markRead() {
    _repo.markAsRead(conversationId);
  }
}

final chatRoomControllerProvider = StateNotifierProvider.autoDispose
    .family<ChatRoomController, ChatRoomState, String>((ref, conversationId) {
  final repo = ref.watch(chatRepositoryProvider);
  return ChatRoomController(repo, conversationId);
});
