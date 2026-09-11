import { describe, expect, it } from 'vitest'

import { getCallRuntimeCopy } from './callRuntimeCopy'

describe('call runtime copy', () => {
  it('localizes device failures and toggle recovery', () => {
    expect(getCallRuntimeCopy('vi').microphoneFailed).toContain('Không thể bật micro')
    expect(getCallRuntimeCopy('en').microphoneFailed).toContain('Could not enable the microphone')
    expect(getCallRuntimeCopy('vi').cameraFailed).toContain('Không thể bật camera')
    expect(getCallRuntimeCopy('en').cameraFailed).toContain('Could not enable the camera')
    expect(getCallRuntimeCopy('vi').microphoneToggleFailed).toContain('Không thể thay đổi micro')
    expect(getCallRuntimeCopy('en').cameraToggleFailed).toContain('Could not change the camera')
  })

  it('localizes connection recovery states', () => {
    expect(getCallRuntimeCopy('vi').reconnecting).toContain('Đang thử kết nối lại')
    expect(getCallRuntimeCopy('en').reconnecting).toContain('Reconnecting')
    expect(getCallRuntimeCopy('vi').disconnected).toContain('Mất kết nối')
    expect(getCallRuntimeCopy('en').disconnected).toContain('disconnected')
  })

  it('localizes signaling outcomes', () => {
    expect(getCallRuntimeCopy('vi').declined).toBe('Người nhận đã từ chối cuộc gọi')
    expect(getCallRuntimeCopy('en').declined).toBe('The recipient declined the call')
    expect(getCallRuntimeCopy('vi').ended).toBe('Cuộc gọi đã kết thúc')
    expect(getCallRuntimeCopy('en').ended).toBe('The call has ended')
  })
})
