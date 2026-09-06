import { getCurrentScope, onScopeDispose, ref } from 'vue'

export interface LongPressOptions {
  delay?: number
  moveThreshold?: number
  onLongPress: (payload?: unknown) => void
}

export function useLongPress(options: LongPressOptions) {
  const delay = options.delay ?? 450
  const moveThreshold = options.moveThreshold ?? 10
  const isPressing = ref(false)
  const shouldIgnoreClick = ref(false)

  let timer: ReturnType<typeof setTimeout> | null = null
  let startX = 0
  let startY = 0
  let longPressTriggered = false
  let currentPayload: unknown = null

  function start(e: TouchEvent | MouseEvent, payload?: unknown) {
    if (timer) clearTimeout(timer)
    longPressTriggered = false
    currentPayload = payload
    isPressing.value = true

    if ('touches' in e && e.touches[0]) {
      startX = e.touches[0].clientX
      startY = e.touches[0].clientY
    } else if ('clientX' in e) {
      startX = e.clientX
      startY = e.clientY
    }

    timer = setTimeout(() => {
      longPressTriggered = true
      isPressing.value = false
      shouldIgnoreClick.value = true
      // Reset ignore flag after browser dispatches the synthetic click
      requestAnimationFrame(() => {
        setTimeout(() => {
          shouldIgnoreClick.value = false
        }, 0)
      })
      if ('vibrate' in navigator) {
        try {
          navigator.vibrate(35)
        } catch {
          // ignore if vibration not permitted
        }
      }
      options.onLongPress(currentPayload)
    }, delay)
  }

  function move(e: TouchEvent | MouseEvent) {
    if (!timer) return
    let currentX = 0
    let currentY = 0

    if ('touches' in e && e.touches[0]) {
      currentX = e.touches[0].clientX
      currentY = e.touches[0].clientY
    } else if ('clientX' in e) {
      currentX = e.clientX
      currentY = e.clientY
    }

    const dist = Math.hypot(currentX - startX, currentY - startY)
    if (dist > moveThreshold) {
      cancel()
    }
  }

  function end() {
    const wasTriggered = longPressTriggered
    cancel()
    return wasTriggered
  }

  function cancel() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
    isPressing.value = false
  }

  if (getCurrentScope()) {
    onScopeDispose(() => {
      cancel()
    })
  }

  return {
    isPressing,
    shouldIgnoreClick,
    start,
    move,
    end,
    cancel,
  }
}
