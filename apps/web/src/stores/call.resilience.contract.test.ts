import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./call.ts', import.meta.url)), 'utf-8')

describe('call store resilience contract', () => {
  it('tracks LiveKit reconnect transitions without tearing down a transient call', () => {
    expect(source).toContain("export type CallConnectionStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting'")
    expect(source).toMatch(/RoomEvent\.Reconnecting[\s\S]*?connectionStatus\.value = 'reconnecting'/)
    expect(source).toMatch(/RoomEvent\.Reconnected[\s\S]*?connectionStatus\.value = 'connected'/)

    const reconnectHandler = source.match(/r\.on\(RoomEvent\.Reconnecting,[\s\S]*?\n\s*\}\)/)?.[0] ?? ''
    expect(reconnectHandler).not.toContain('endCall(')
  })

  it('distinguishes intentional hangup from unexpected room disconnect', () => {
    expect(source).toContain('let intentionalRoomDisconnect = false')
    expect(source).toMatch(/RoomEvent\.Disconnected[\s\S]*?if \(intentionalRoomDisconnect\)/)
    expect(source).toContain('Mất kết nối cuộc gọi. Hãy kiểm tra mạng và gọi lại.')
  })

  it('stores device failures and clears them after a successful retry', () => {
    expect(source).toContain('const deviceError = ref<string | null>(null)')
    expect(source).toMatch(/setMicrophoneEnabled\(next\)[\s\S]*?deviceError\.value = null[\s\S]*?Không thể thay đổi micro/)
    expect(source).toMatch(/setCameraEnabled\(next\)[\s\S]*?deviceError\.value = null[\s\S]*?Không thể thay đổi camera/)
    expect(source).toMatch(/return \{[\s\S]*?connectionStatus,[\s\S]*?deviceError,/)
  })
})
