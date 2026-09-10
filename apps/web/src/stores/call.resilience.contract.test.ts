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

  it('scopes intentional disconnects to the exact LiveKit room instance', () => {
    expect(source).toContain('const intentionalDisconnectRooms = new WeakSet<Room>()')
    expect(source).toContain('intentionalDisconnectRooms.add(activeRoom)')
    expect(source).toMatch(/RoomEvent\.Disconnected[\s\S]*?intentionalDisconnectRooms\.has\(r\)/)
    expect(source).toContain('Mất kết nối cuộc gọi. Hãy kiểm tra mạng và gọi lại.')
  })

  it('stores mic and camera failures independently and clears each after retry', () => {
    expect(source).toContain('const microphoneError = ref<string | null>(null)')
    expect(source).toContain('const cameraError = ref<string | null>(null)')
    expect(source).toContain('const deviceError = computed(() => microphoneError.value ?? cameraError.value)')
    expect(source).toMatch(/setMicrophoneEnabled\(next\)[\s\S]*?microphoneError\.value = null[\s\S]*?Không thể thay đổi micro/)
    expect(source).toMatch(/setCameraEnabled\(next\)[\s\S]*?cameraError\.value = null[\s\S]*?Không thể thay đổi camera/)
    expect(source).toMatch(/return \{[\s\S]*?connectionStatus,[\s\S]*?deviceError,/)
  })
})
