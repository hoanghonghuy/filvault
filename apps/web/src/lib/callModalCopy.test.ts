import { describe, expect, it } from 'vitest'
import { callModalCopy } from './callModalCopy'

describe('callModalCopy', () => {
  it('provides Vietnamese call modal chrome labels', () => {
    expect(callModalCopy('vi')).toEqual({
      incomingVideo: 'Cuộc gọi video đến',
      incomingAudio: 'Cuộc gọi thoại đến',
      outgoingVideo: 'Đang gọi video…',
      outgoingAudio: 'Đang gọi thoại…',
      videoCall: 'Cuộc gọi video',
      audioCall: 'Cuộc gọi thoại',
      member: 'Thành viên',
      conversation: 'Cuộc trò chuyện',
      declineCall: 'Từ chối cuộc gọi',
      decline: 'Từ chối',
      answerCall: 'Trả lời cuộc gọi',
      answer: 'Trả lời',
      ringing: 'Đang đổ chuông…',
      cancelCall: 'Hủy cuộc gọi',
      cancel: 'Hủy',
      minimize: 'Thu nhỏ',
      fullscreen: 'Toàn màn hình',
      remoteCameraOff: 'Camera đối phương đang tắt',
      inCall: 'Đang trong cuộc gọi',
      cameraOff: 'Camera tắt',
      muteMic: 'Tắt mic',
      unmuteMic: 'Bật mic',
      turnCameraOff: 'Tắt camera',
      turnCameraOn: 'Bật camera',
      endCall: 'Kết thúc cuộc gọi',
      callEnded: 'Cuộc gọi đã kết thúc',
    })
  })

  it('provides English call modal chrome labels', () => {
    expect(callModalCopy('en')).toEqual({
      incomingVideo: 'Incoming video call',
      incomingAudio: 'Incoming voice call',
      outgoingVideo: 'Calling video…',
      outgoingAudio: 'Calling…',
      videoCall: 'Video call',
      audioCall: 'Voice call',
      member: 'Member',
      conversation: 'Conversation',
      declineCall: 'Decline call',
      decline: 'Decline',
      answerCall: 'Answer call',
      answer: 'Answer',
      ringing: 'Ringing…',
      cancelCall: 'Cancel call',
      cancel: 'Cancel',
      minimize: 'Exit full screen',
      fullscreen: 'Full screen',
      remoteCameraOff: 'Their camera is off',
      inCall: 'In call',
      cameraOff: 'Camera off',
      muteMic: 'Mute microphone',
      unmuteMic: 'Unmute microphone',
      turnCameraOff: 'Turn camera off',
      turnCameraOn: 'Turn camera on',
      endCall: 'End call',
      callEnded: 'Call ended',
    })
  })
})
