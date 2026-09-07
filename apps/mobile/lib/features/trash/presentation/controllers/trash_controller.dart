import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/trash_model.dart';
import '../../data/trash_repository.dart';

final trashRepositoryProvider = Provider<TrashRepository>((ref) {
  return TrashRepositoryImpl();
});

class TrashState {
  final bool isLoading;
  final String? errorMessage;
  final List<TrashItemModel> folders;
  final List<TrashItemModel> files;

  const TrashState({
    this.isLoading = false,
    this.errorMessage,
    this.folders = const [],
    this.files = const [],
  });

  bool get isEmpty => folders.isEmpty && files.isEmpty;

  TrashState copyWith({
    bool? isLoading,
    String? errorMessage,
    List<TrashItemModel>? folders,
    List<TrashItemModel>? files,
  }) {
    return TrashState(
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      folders: folders ?? this.folders,
      files: files ?? this.files,
    );
  }
}

class TrashController extends StateNotifier<TrashState> {
  final TrashRepository _repo;

  TrashController(this._repo) : super(const TrashState()) {
    loadTrash();
  }

  Future<void> loadTrash() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final list = await _repo.getTrash();
      state = state.copyWith(
        isLoading: false,
        folders: list.folders,
        files: list.files,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<bool> restoreFile(String fileId) async {
    try {
      await _repo.restoreFile(fileId);
      state = state.copyWith(
        files: state.files.where((f) => f.id != fileId).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> restoreFolder(String folderId) async {
    try {
      await _repo.restoreFolder(folderId);
      state = state.copyWith(
        folders: state.folders.where((f) => f.id != folderId).toList(),
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }

  Future<bool> permanentDelete({required String type, required String id}) async {
    try {
      await _repo.permanentDelete(type: type, id: id);
      if (type == 'files') {
        state = state.copyWith(
          files: state.files.where((f) => f.id != id).toList(),
        );
      } else {
        state = state.copyWith(
          folders: state.folders.where((f) => f.id != id).toList(),
        );
      }
      return true;
    } catch (e) {
      state = state.copyWith(
        errorMessage: e.toString().replaceFirst('Exception: ', ''),
      );
      return false;
    }
  }
}

final trashControllerProvider =
    StateNotifierProvider<TrashController, TrashState>((ref) {
  final repo = ref.watch(trashRepositoryProvider);
  return TrashController(repo);
});
