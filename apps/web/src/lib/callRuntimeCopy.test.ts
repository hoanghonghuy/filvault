import { describe, expect, it } from 'vitest'

import { getCallRuntimeCopy } from './callRuntimeCopy'

describe('call runtime copy', () => {
  it('localizes device failures', () => {
    expect(getCallRuntimeCopy('vi').microphoneFailed).toContain('Không thể bật micro')
    expect(getCallRuntimeCopy('en').microphoneFailed).toContain('Could not enable the microphone')
    expect(getCallRuntimeCopy('vi').cameraFailed).toContain('Không thể bật camera')
    expect(getCallRuntimeCopy('en').cameraFailed).toContain('Could not enable the camera')
  })

  it('localizes signaling outcomes', () => {
    expect(getCallRuntimeCopy('vi').declined).toBe('Người nhận đã từ chối cuộc gọi')
    expect(getCallRuntimeCopy('en').declined).toBe('The recipient declined the call')
    expect(getCallRuntimeCopy('vi').ended).toBe('Cuộc gọi đã kết thúc')
    expect(getCallRuntimeCopy('en').ended).toBe('The call has ended')
  })
})
