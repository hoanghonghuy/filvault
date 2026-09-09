import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:livekit_client/livekit_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../controllers/call_controller.dart';

class CallScreen extends ConsumerWidget {
  final String conversationId;
  final String peerName;
  final bool isVideo;

  const CallScreen({
    super.key,
    required this.conversationId,
    required this.peerName,
    required this.isVideo,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final args = (conversationId: conversationId, isVideo: isVideo);
    final state = ref.watch(callControllerProvider(args));
    final ctrl = ref.read(callControllerProvider(args).notifier);

    // Auto-pop when call is ended
    if (state.isEnded) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (context.mounted) Navigator.pop(context);
      });
    }

    return Scaffold(
      backgroundColor: const Color(0xFF0F172A),
      body: SafeArea(
        child: Stack(
          fit: StackFit.expand,
          children: [
            // Remote video or Avatar placeholder
            if (state.remoteVideoTrack != null && isVideo)
              VideoTrackRenderer(state.remoteVideoTrack!)
            else
              _buildAvatarPlaceholder(context, state),

            // Top Header: Peer info & Call duration
            Positioned(
              top: 20,
              left: 20,
              right: 20,
              child: Column(
                children: [
                  Text(
                    peerName,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 22,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.black45,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      state.isConnecting
                          ? 'Đang kết nối...'
                          : state.isConnected
                              ? _formatDuration(state.duration)
                              : 'Đang gọi...',
                      style: const TextStyle(
                        color: Colors.white70,
                        fontSize: 13,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Error banner if any
            if (state.errorMessage != null)
              Positioned(
                top: 90,
                left: 20,
                right: 20,
                child: Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.redAccent.withOpacity(0.9),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    state.errorMessage!,
                    style: const TextStyle(color: Colors.white, fontSize: 13),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),

            // Local Video PiP (Picture in Picture)
            if (isVideo && state.localVideoTrack != null && !state.isCameraOff)
              Positioned(
                right: 16,
                top: 80,
                width: 100,
                height: 150,
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Container(
                    decoration: BoxDecoration(
                      color: Colors.black54,
                      border: Border.all(color: Colors.white24, width: 1.5),
                    ),
                    child: VideoTrackRenderer(state.localVideoTrack!),
                  ),
                ),
              ),

            // Bottom Call Action Controls
            Positioned(
              bottom: 36,
              left: 0,
              right: 0,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  // Mute / Unmute
                  _buildCircleButton(
                    icon: state.isMuted ? Icons.mic_off_rounded : Icons.mic_rounded,
                    isActive: state.isMuted,
                    onTap: () => ctrl.toggleMute(),
                  ),

                  // Camera On/Off (if video call)
                  if (isVideo)
                    _buildCircleButton(
                      icon: state.isCameraOff ? Icons.videocam_off_rounded : Icons.videocam_rounded,
                      isActive: state.isCameraOff,
                      onTap: () => ctrl.toggleCamera(),
                    ),

                  // Flip Camera (if video call)
                  if (isVideo)
                    _buildCircleButton(
                      icon: Icons.flip_camera_ios_rounded,
                      isActive: false,
                      onTap: () => ctrl.flipCamera(),
                    ),

                  // End Call (Red button)
                  _buildCircleButton(
                    icon: Icons.call_end_rounded,
                    backgroundColor: Colors.redAccent,
                    iconColor: Colors.white,
                    size: 64,
                    onTap: () => ctrl.endCall(),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAvatarPlaceholder(BuildContext context, CallState state) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          CircleAvatar(
            radius: 54,
            backgroundColor: AppColors.accent.withOpacity(0.3),
            child: Text(
              peerName.isNotEmpty ? peerName[0].toUpperCase() : 'U',
              style: const TextStyle(fontSize: 42, color: Colors.white, fontWeight: FontWeight.bold),
            ),
          ),
          const SizedBox(height: 20),
          Text(
            state.isConnected ? 'Đang trong cuộc gọi thoại' : 'Đang đổ chuông...',
            style: const TextStyle(color: Colors.white54, fontSize: 14),
          ),
        ],
      ),
    );
  }

  Widget _buildCircleButton({
    required IconData icon,
    bool isActive = false,
    Color? backgroundColor,
    Color? iconColor,
    double size = 52,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: backgroundColor ?? (isActive ? Colors.white : Colors.white24),
          shape: BoxShape.circle,
        ),
        child: Icon(
          icon,
          color: iconColor ?? (isActive ? Colors.black87 : Colors.white),
          size: size * 0.48,
        ),
      ),
    );
  }

  String _formatDuration(Duration d) {
    final minutes = d.inMinutes.toString().padLeft(2, '0');
    final seconds = (d.inSeconds % 60).toString().padLeft(2, '0');
    return '$minutes:$seconds';
  }
}
