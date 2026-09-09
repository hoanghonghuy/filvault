import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/chat_repository.dart';
import '../../data/chat_sse_service.dart';
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
  final ChatSseService _sse;
  final String conversationId;
  Timer? _typingDebounce;
  Timer? _peerTypingTimer;
  StreamSubscription<ChatSseEvent>? _subscription;

  ChatRoomController(this._repo, this._sse, this.conversationId)
      : super(const ChatRoomState()) {
    loadMessages();
    _markRead();
    _subscribeSse();
  }

  @override
  void dispose() {
    _typingDebounce?.cancel();
    _peerTypingTimer?.cancel();
    _subscription?.cancel();
    super.dispose();
  }

  void _subscribeSse() {
    _subscription = _sse.eventStream.listen((event) {
      final data = event.data;
      final cid = data['conversationId'] as String?;
      if (cid != conversationId) return;

      if (event.type == 'message.new') {
        try {
          final newMsg = ChatMessageModel.fromJson(data);
          if (!state.messages.any((m) => m.id == newMsg.id)) {
            state = state.copyWith(messages: [newMsg, ...state.messages]);
            _markRead();
          }
        } catch (_) {}
      } else if (event.type == 'message.remove') {
        final mid = data['messageId'] as String?;
        if (mid != null) {
          state = state.copyWith(
            messages: state.messages.where((m) => m.id != mid).toList(),
          );
        }
      } else if (event.type == 'message.reaction' || event.type == 'message.edit') {
        loadMessages();
      } else if (event.type == 'typing') {
        final isTyping = data['typing'] as bool? ?? false;
        if (isTyping) {
          state = state.copyWith(isPeerTyping: true);
          _peerTypingTimer?.cancel();
          _peerTypingTimer = Timer(const Duration(seconds: 4), () {
            state = state.copyWith(isPeerTyping: false);
          });
        } else {
          state = state.copyWith(isPeerTyping: false);
        }
      }
    });
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

      if (!state.messages.any((m) => m.id == newMsg.id)) {
        state = state.copyWith(
          isSending: false,
          messages: [newMsg, ...state.messages],
          clearReply: true,
        );
      } else {
        state = state.copyWith(isSending: false, clearReply: true);
      }
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
  final sse = ref.watch(chatSseServiceProvider);
  return ChatRoomController(repo, sse, conversationId);
});
