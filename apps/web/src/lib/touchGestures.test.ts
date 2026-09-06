/**
 * @vitest-environment jsdom
 */
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useLongPress } from './useLongPress'
import { usePullToRefresh } from './usePullToRefresh'

describe('useLongPress', () => {
  it('triggers onLongPress callback after delay', async () => {
    vi.useFakeTimers()
    const onLongPress = vi.fn<() => void>()
    const { start, end } = useLongPress({ delay: 300, onLongPress })

    const fakeEvent = { clientX: 100, clientY: 100 } as MouseEvent
    start(fakeEvent, { id: 'file-1' })

    expect(onLongPress).not.toHaveBeenCalled()
    vi.advanceTimersByTime(310)

    expect(onLongPress).toHaveBeenCalledWith({ id: 'file-1' })
    end()
    vi.useRealTimers()
  })

  it('cancels if touch moves past threshold', () => {
    vi.useFakeTimers()
    const onLongPress = vi.fn<() => void>()
    const { start, move } = useLongPress({ delay: 300, moveThreshold: 10, onLongPress })

    start({ clientX: 100, clientY: 100 } as MouseEvent)
    move({ clientX: 100, clientY: 125 } as MouseEvent)

    vi.advanceTimersByTime(310)
    expect(onLongPress).not.toHaveBeenCalled()
    vi.useRealTimers()
  })
})

describe('usePullToRefresh', () => {
  it('tracks pullDistance on downward touch moves', () => {
    const el = document.createElement('div')
    const containerRef = ref(el)
    const onRefresh = vi.fn<() => Promise<void>>().mockResolvedValue(undefined)

    const { pullDistance, attachListeners } = usePullToRefresh(containerRef, {
      threshold: 50,
      onRefresh,
    })
    attachListeners(el)

    el.dispatchEvent(
      new TouchEvent('touchstart', {
        touches: [{ clientX: 50, clientY: 50 } as Touch],
      }),
    )

    el.dispatchEvent(
      new TouchEvent('touchmove', {
        touches: [{ clientX: 50, clientY: 150 } as Touch],
      }),
    )

    expect(pullDistance.value).toBeGreaterThan(0)
  })
})
