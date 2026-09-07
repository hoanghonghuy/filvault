import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/share_model.dart';
import '../../data/shares_repository.dart';

final sharesRepositoryProvider = Provider<SharesRepository>((ref) {
  return SharesRepositoryImpl();
});

class SharesState {
  final bool isLoading;
  final String? errorMessage;
  final List<IncomingShareModel> incoming;
  final List<OutgoingShareModel> outgoing;

  const SharesState({
    this.isLoading = false,
    this.errorMessage,
    this.incoming = const [],
    this.outgoing = const [],
  });

  SharesState copyWith({
    bool? isLoading,
    String? errorMessage,
    List<IncomingShareModel>? incoming,
    List<OutgoingShareModel>? outgoing,
  }) {
    return SharesState(
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      incoming: incoming ?? this.incoming,
      outgoing: outgoing ?? this.outgoing,
    );
  }
}

class SharesController extends StateNotifier<SharesState> {
  final SharesRepository _repo;

  SharesController(this._repo) : super(const SharesState()) {
    loadShares();
  }

  Future<void> loadShares() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final results = await Future.wait([
        _repo.getIncomingShares(),
        _repo.getOutgoingShares(),
      ]);
      state = state.copyWith(
        isLoading: false,
        incoming: results[0] as List<IncomingShareModel>,
        outgoing: results[1] as List<OutgoingShareModel>,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<bool> revokeShare(String shareId) async {
    try {
      await _repo.revokeShare(shareId);
      state = state.copyWith(
        outgoing: state.outgoing.where((s) => s.id != shareId).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> createShare({
    required String resourceType,
    required String resourceId,
    required String email,
  }) async {
    try {
      await _repo.createShare(
        resourceType: resourceType,
        resourceId: resourceId,
        email: email,
      );
      await loadShares();
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }
}

final sharesControllerProvider =
    StateNotifierProvider<SharesController, SharesState>((ref) {
  final repo = ref.watch(sharesRepositoryProvider);
  return SharesController(repo);
});
