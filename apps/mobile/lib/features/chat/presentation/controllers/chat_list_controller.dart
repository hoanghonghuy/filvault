import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/chat_repository.dart';
import '../../data/models/chat_model.dart';

final chatRepositoryProvider = Provider<ChatRepository>((ref) {
  return ChatRepositoryImpl();
});

class ChatListState {
  final bool isLoading;
  final String? errorMessage;
  final List<ChatConversationModel> conversations;

  const ChatListState({
    this.isLoading = false,
    this.errorMessage,
    this.conversations = const [],
  });

  int get totalUnread => conversations.fold(0, (sum, c) => sum + c.unreadCount);

  ChatListState copyWith({
    bool? isLoading,
    String? errorMessage,
    List<ChatConversationModel>? conversations,
  }) {
    return ChatListState(
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      conversations: conversations ?? this.conversations,
    );
  }
}

class ChatListController extends StateNotifier<ChatListState> {
  final ChatRepository _repo;

  ChatListController(this._repo) : super(const ChatListState()) {
    loadConversations();
  }

  Future<void> loadConversations() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final list = await _repo.getConversations();
      state = state.copyWith(
        isLoading: false,
        conversations: list,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<ChatConversationModel?> startDirectChat(String targetUserId) async {
    try {
      final convo = await _repo.createDirectConversation(targetUserId);
      await loadConversations();
      return convo;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return null;
    }
  }
}

final chatListControllerProvider =
    StateNotifierProvider<ChatListController, ChatListState>((ref) {
  final repo = ref.watch(chatRepositoryProvider);
  return ChatListController(repo);
});
