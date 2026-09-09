import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/photo_model.dart';
import '../../data/photos_repository.dart';

final photosRepositoryProvider = Provider<PhotosRepository>((ref) {
  return PhotosRepositoryImpl();
});

class PhotosState {
  final bool isLoading;
  final bool isLoadingMore;
  final String? errorMessage;
  final int selectedTab;
  final List<TimelineGroupModel> timelineGroups;
  final String? nextBefore;
  final List<AlbumModel> albums;

  const PhotosState({
    this.isLoading = false,
    this.isLoadingMore = false,
    this.errorMessage,
    this.selectedTab = 0,
    this.timelineGroups = const [],
    this.nextBefore,
    this.albums = const [],
  });

  bool get hasMore => nextBefore != null && nextBefore!.isNotEmpty;

  PhotosState copyWith({
    bool? isLoading,
    bool? isLoadingMore,
    String? errorMessage,
    int? selectedTab,
    List<TimelineGroupModel>? timelineGroups,
    String? nextBefore,
    bool clearNextBefore = false,
    List<AlbumModel>? albums,
  }) {
    return PhotosState(
      isLoading: isLoading ?? this.isLoading,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      errorMessage: errorMessage,
      selectedTab: selectedTab ?? this.selectedTab,
      timelineGroups: timelineGroups ?? this.timelineGroups,
      nextBefore: clearNextBefore ? null : (nextBefore ?? this.nextBefore),
      albums: albums ?? this.albums,
    );
  }
}

class PhotosController extends StateNotifier<PhotosState> {
  final PhotosRepository _repo;

  PhotosController(this._repo) : super(const PhotosState()) {
    init();
  }

  Future<void> init() async {
    await Future.wait([
      loadTimeline(),
      loadAlbums(),
    ]);
  }

  void switchTab(int index) {
    state = state.copyWith(selectedTab: index);
  }

  Future<void> loadTimeline() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final result = await _repo.getTimeline(limit: 50);
      state = state.copyWith(
        isLoading: false,
        timelineGroups: result.groups,
        nextBefore: result.nextBefore,
        clearNextBefore: result.nextBefore == null,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> loadMoreTimeline() async {
    if (state.isLoadingMore || !state.hasMore) return;
    state = state.copyWith(isLoadingMore: true);
    try {
      final result = await _repo.getTimeline(before: state.nextBefore, limit: 50);
      
      // Merge timeline groups
      final currentGroups = List<TimelineGroupModel>.from(state.timelineGroups);
      for (final newGroup in result.groups) {
        final existingIdx = currentGroups.indexWhere((g) => g.date == newGroup.date);
        if (existingIdx != -1) {
          final mergedItems = [...currentGroups[existingIdx].items, ...newGroup.items];
          currentGroups[existingIdx] = TimelineGroupModel(
            date: newGroup.date,
            items: mergedItems,
          );
        } else {
          currentGroups.add(newGroup);
        }
      }

      state = state.copyWith(
        isLoadingMore: false,
        timelineGroups: currentGroups,
        nextBefore: result.nextBefore,
        clearNextBefore: result.nextBefore == null,
      );
    } catch (e) {
      state = state.copyWith(
        isLoadingMore: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> loadAlbums() async {
    try {
      final albums = await _repo.getAlbums();
      state = state.copyWith(albums: albums);
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<bool> createAlbum(String name) async {
    try {
      await _repo.createAlbum(name);
      await loadAlbums();
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> deleteAlbum(String albumId) async {
    try {
      await _repo.deleteAlbum(albumId);
      state = state.copyWith(
        albums: state.albums.where((a) => a.id != albumId).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }
}

final photosControllerProvider =
    StateNotifierProvider<PhotosController, PhotosState>((ref) {
  final repo = ref.watch(photosRepositoryProvider);
  return PhotosController(repo);
});
