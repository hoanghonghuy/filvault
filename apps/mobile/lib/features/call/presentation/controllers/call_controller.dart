import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:livekit_client/livekit_client.dart';
import '../../data/call_repository.dart';

final callRepositoryProvider = Provider<CallRepository>((ref) {
  return CallRepositoryImpl();
});

class CallState {
  final bool isConnecting;
  final bool isConnected;
  final bool isEnded;
  final bool isMuted;
  final bool isCameraOff;
  final bool isSpeakerOn;
  final String? errorMessage;
  final Duration duration;
  final Room? room;
  final RemoteParticipant? remoteParticipant;
  final VideoTrack? remoteVideoTrack;
  final VideoTrack? localVideoTrack;

  const CallState({
    this.isConnecting = true,
    this.isConnected = false,
    this.isEnded = false,
    this.isMuted = false,
    this.isCameraOff = false,
    this.isSpeakerOn = true,
    this.errorMessage,
    this.duration = Duration.zero,
    this.room,
    this.remoteParticipant,
    this.remoteVideoTrack,
    this.localVideoTrack,
  });

  CallState copyWith({
    bool? isConnecting,
    bool? isConnected,
    bool? isEnded,
    bool? isMuted,
    bool? isCameraOff,
    bool? isSpeakerOn,
    String? errorMessage,
    Duration? duration,
    Room? room,
    RemoteParticipant? remoteParticipant,
    bool clearRemoteParticipant = false,
    VideoTrack? remoteVideoTrack,
    bool clearRemoteVideo = false,
    VideoTrack? localVideoTrack,
    bool clearLocalVideo = false,
  }) {
    return CallState(
      isConnecting: isConnecting ?? this.isConnecting,
      isConnected: isConnected ?? this.isConnected,
      isEnded: isEnded ?? this.isEnded,
      isMuted: isMuted ?? this.isMuted,
      isCameraOff: isCameraOff ?? this.isCameraOff,
      isSpeakerOn: isSpeakerOn ?? this.isSpeakerOn,
      errorMessage: errorMessage,
      duration: duration ?? this.duration,
      room: room ?? this.room,
      remoteParticipant: clearRemoteParticipant
          ? null
          : (remoteParticipant ?? this.remoteParticipant),
      remoteVideoTrack:
          clearRemoteVideo ? null : (remoteVideoTrack ?? this.remoteVideoTrack),
      localVideoTrack:
          clearLocalVideo ? null : (localVideoTrack ?? this.localVideoTrack),
    );
  }
}

class CallController extends StateNotifier<CallState> {
  final CallRepository _repo;
  final String conversationId;
  final bool isVideo;
  Timer? _durationTimer;
  EventsListener<RoomEvent>? _listener;

  CallController(this._repo, {required this.conversationId, required this.isVideo})
      : super(CallState(isCameraOff: !isVideo)) {
    _startCall();
  }

  @override
  void dispose() {
    _durationTimer?.cancel();
    _listener?.dispose();
    state.room?.disconnect();
    state.room?.dispose();
    super.dispose();
  }

  Future<void> _startCall() async {
    try {
      // 1. Send call offer signal
      await _repo.sendCallSignal(
        conversationId,
        action: 'offer',
        isVideo: isVideo,
      );

      // 2. Fetch LiveKit token
      final info = await _repo.getCallToken(conversationId);

      // 3. Connect to LiveKit Room
      final room = Room(
        roomOptions: const RoomOptions(
          adaptiveStream: true,
          dynacast: true,
          defaultCameraCaptureOptions: CameraCaptureOptions(
            maxFrameRate: 30,
            params: VideoParametersPresets.h720_169,
          ),
        ),
      );

      _listener = room.createListener();
      _setupRoomEvents(_listener!);

      // If emulator, replace localhost with 10.0.2.2
      var wsUrl = info.url;
      if (wsUrl.contains('localhost') || wsUrl.contains('127.0.0.1')) {
        wsUrl = wsUrl.replaceAll('localhost', '10.0.2.2').replaceAll('127.0.0.1', '10.0.2.2');
      }

      await room.connect(wsUrl, info.token);

      // 4. Publish local audio and video
      await room.localParticipant?.setMicrophoneEnabled(true);
      if (isVideo) {
        await room.localParticipant?.setCameraEnabled(true);
      }

      VideoTrack? localTrack;
      for (final pub in room.localParticipant?.videoTrackPublications ?? []) {
        if (pub.track != null) {
          localTrack = pub.track;
          break;
        }
      }

      state = state.copyWith(
        isConnecting: false,
        isConnected: true,
        room: room,
        localVideoTrack: localTrack,
      );

      _startTimer();
    } catch (e) {
      state = state.copyWith(
        isConnecting: false,
        errorMessage: 'Lỗi kết nối cuộc gọi: ${e.toString().replaceFirst("Exception: ", "")}',
      );
    }
  }

  void _setupRoomEvents(EventsListener<RoomEvent> listener) {
    listener
      ..on<ParticipantConnectedEvent>((event) {
        state = state.copyWith(remoteParticipant: event.participant);
      })
      ..on<ParticipantDisconnectedEvent>((event) {
        state = state.copyWith(
          clearRemoteParticipant: true,
          clearRemoteVideo: true,
        );
      })
      ..on<TrackSubscribedEvent>((event) {
        if (event.track is VideoTrack) {
          state = state.copyWith(remoteVideoTrack: event.track as VideoTrack);
        }
      })
      ..on<TrackUnsubscribedEvent>((event) {
        if (event.track is VideoTrack) {
          state = state.copyWith(clearRemoteVideo: true);
        }
      })
      ..on<RoomDisconnectedEvent>((event) {
        _endCallLocally();
      });
  }

  void _startTimer() {
    _durationTimer?.cancel();
    _durationTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      state = state.copyWith(duration: Duration(seconds: t.tick));
    });
  }

  Future<void> toggleMute() async {
    final newMute = !state.isMuted;
    await state.room?.localParticipant?.setMicrophoneEnabled(!newMute);
    state = state.copyWith(isMuted: newMute);
  }

  Future<void> toggleCamera() async {
    final newCamOff = !state.isCameraOff;
    await state.room?.localParticipant?.setCameraEnabled(!newCamOff);
    state = state.copyWith(isCameraOff: newCamOff);
  }

  Future<void> flipCamera() async {
    final videoPub = state.room?.localParticipant?.videoTrackPublications.firstOrNull;
    if (videoPub?.track is LocalVideoTrack) {
      final localTrack = videoPub!.track as LocalVideoTrack;
      // Camera switch supported by track
      final options = localTrack.currentOptions;
      if (options is CameraCaptureOptions) {
        final newPosition = options.cameraPosition == CameraPosition.front
            ? CameraPosition.back
            : CameraPosition.front;
        await localTrack.restartTrack(options.copyWith(cameraPosition: newPosition));
      }
    }
  }

  Future<void> endCall() async {
    try {
      await _repo.sendCallSignal(
        conversationId,
        action: 'end',
        isVideo: isVideo,
      );
    } catch (_) {}
    _endCallLocally();
  }

  void _endCallLocally() {
    _durationTimer?.cancel();
    state.room?.disconnect();
    state = state.copyWith(
      isConnected: false,
      isEnded: true,
      clearRemoteParticipant: true,
      clearRemoteVideo: true,
    );
  }
}

final callControllerProvider = StateNotifierProvider.autoDispose
    .family<CallController, CallState, ({String conversationId, bool isVideo})>((ref, args) {
  final repo = ref.watch(callRepositoryProvider);
  return CallController(
    repo,
    conversationId: args.conversationId,
    isVideo: args.isVideo,
  );
});
