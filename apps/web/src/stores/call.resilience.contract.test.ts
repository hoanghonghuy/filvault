import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./call.ts', import.meta.url)), 'utf-8')
const runtimeCopy = readFileSync(fileURLToPath(new URL('../lib/callRuntimeCopy.ts', import.meta.url)), 'utf-8')

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
    expect(source).toContain('const message = runtimeCopy.value.disconnected')
    expect(runtimeCopy).toContain("disconnected: 'Mất kết nối cuộc gọi. Hãy kiểm tra mạng và gọi lại.'")
    expect(runtimeCopy).toContain("disconnected: 'The call disconnected. Check your network and call again.'")
  })

  it('identifies secure-context failures independently of localized copy', () => {
    expect(source).toContain("const MEDIA_REQUIRES_SECURE_CONTEXT = Symbol('media-requires-secure-context')")
    expect(source).toContain('throw MEDIA_REQUIRES_SECURE_CONTEXT')
    expect(source).toMatch(/e === MEDIA_REQUIRES_SECURE_CONTEXT[\s\S]*?runtimeCopy\.value\.mediaRequiresSecureContext/)
    expect(source).not.toContain('err?.message === runtimeCopy.value.mediaRequiresSecureContext')
  })

  it('stores mic and camera failures independently and clears each after retry', () => {
    expect(source).toContain('const microphoneError = ref<string | null>(null)')
    expect(source).toContain('const cameraError = ref<string | null>(null)')
    expect(source).toContain('const deviceError = computed(() => microphoneError.value ?? cameraError.value)')
    expect(source).toMatch(/setMicrophoneEnabled\(next\)[\s\S]*?microphoneError\.value = null[\s\S]*?microphoneError\.value = runtimeCopy\.value\.microphoneToggleFailed/)
    expect(source).toMatch(/setCameraEnabled\(next\)[\s\S]*?cameraError\.value = null[\s\S]*?cameraError\.value = runtimeCopy\.value\.cameraToggleFailed/)
    expect(runtimeCopy).toContain("microphoneToggleFailed: 'Không thể thay đổi micro")
    expect(runtimeCopy).toContain("cameraToggleFailed: 'Không thể thay đổi camera")
    expect(source).toMatch(/return \{[\s\S]*?connectionStatus,[\s\S]*?deviceError,/)
  })
})
