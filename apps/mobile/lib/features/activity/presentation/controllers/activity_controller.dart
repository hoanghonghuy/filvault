import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/activity_repository.dart';
import '../../data/models/activity_model.dart';

final activityRepositoryProvider = Provider<ActivityRepository>((ref) {
  return ActivityRepositoryImpl();
});

class ActivityState {
  final bool isLoading;
  final bool isLoadingMore;
  final String? errorMessage;
  final List<ActivityEventModel> events;
  final String? nextBefore;

  const ActivityState({
    this.isLoading = false,
    this.isLoadingMore = false,
    this.errorMessage,
    this.events = const [],
    this.nextBefore,
  });

  bool get hasMore => nextBefore != null && nextBefore!.isNotEmpty;

  ActivityState copyWith({
    bool? isLoading,
    bool? isLoadingMore,
    String? errorMessage,
    List<ActivityEventModel>? events,
    String? nextBefore,
    bool clearNextBefore = false,
  }) {
    return ActivityState(
      isLoading: isLoading ?? this.isLoading,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      errorMessage: errorMessage,
      events: events ?? this.events,
      nextBefore: clearNextBefore ? null : (nextBefore ?? this.nextBefore),
    );
  }
}

class ActivityController extends StateNotifier<ActivityState> {
  final ActivityRepository _repo;

  ActivityController(this._repo) : super(const ActivityState()) {
    loadActivity();
  }

  Future<void> loadActivity() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final page = await _repo.getActivity(limit: 50);
      state = state.copyWith(
        isLoading: false,
        events: page.events,
        nextBefore: page.nextBefore,
        clearNextBefore: page.nextBefore == null,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> loadMoreActivity() async {
    if (state.isLoadingMore || !state.hasMore) return;
    state = state.copyWith(isLoadingMore: true);
    try {
      final page = await _repo.getActivity(before: state.nextBefore, limit: 50);
      state = state.copyWith(
        isLoadingMore: false,
        events: [...state.events, ...page.events],
        nextBefore: page.nextBefore,
        clearNextBefore: page.nextBefore == null,
      );
    } catch (e) {
      state = state.copyWith(
        isLoadingMore: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }
}

final activityControllerProvider =
    StateNotifierProvider<ActivityController, ActivityState>((ref) {
  final repo = ref.watch(activityRepositoryProvider);
  return ActivityController(repo);
});
