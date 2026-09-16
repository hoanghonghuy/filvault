import { describe, expect, it } from 'vitest'
import { networkStatusCopy } from './networkStatusCopy'

describe('networkStatusCopy', () => {
  it('provides Vietnamese network status labels', () => {
    expect(networkStatusCopy('vi')).toEqual({
      offline: {
        title: 'Bạn đang ngoại tuyến',
        detail: 'Các thao tác cần mạng sẽ khả dụng lại khi thiết bị kết nối.',
      },
      recovered: {
        title: 'Đã kết nối lại',
        detail: 'Kết nối mạng của thiết bị đã được khôi phục.',
      },
    })
  })

  it('provides English network status labels', () => {
    expect(networkStatusCopy('en')).toEqual({
      offline: {
        title: 'You’re offline',
        detail: 'Network actions will be available again when your device reconnects.',
      },
      recovered: {
        title: 'Back online',
        detail: 'Your device connection has been restored.',
      },
    })
  })
})
