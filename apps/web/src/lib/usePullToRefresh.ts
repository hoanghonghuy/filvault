import { getCurrentInstance, onBeforeUnmount, ref, type Ref } from 'vue'

export interface PullToRefreshOptions {
  threshold?: number
  maxDistance?: number
  onRefresh: () => Promise<void>
}

export function usePullToRefresh(
  containerRef: Ref<HTMLElement | null>,
  options: PullToRefreshOptions,
) {
  const threshold = options.threshold ?? 60
  const maxDistance = options.maxDistance ?? 90
  const pullDistance = ref(0)
  const isPulling = ref(false)
  const isRefreshing = ref(false)

  let startY = 0
  let tracking = false
  let attachedElement: HTMLElement | null = null

  function isScrolledToTop(): boolean {
    const el = containerRef.value
    const elScrolled = el ? el.scrollTop > 0 : false
    const windowScrolled = window.scrollY > 0 || document.documentElement.scrollTop > 0
    return !elScrolled && !windowScrolled
  }

  function onTouchStart(e: TouchEvent) {
    if (isRefreshing.value) return
    if (e.touches.length !== 1) return
    if (!isScrolledToTop()) return
    const touch = e.touches[0]
    if (!touch) return
    startY = touch.clientY
    tracking = true
  }

  function onTouchMove(e: TouchEvent) {
    if (!tracking || isRefreshing.value) return
    if (e.touches.length !== 1) {
      tracking = false
      pullDistance.value = 0
      isPulling.value = false
      return
    }
    const touch = e.touches[0]
    if (!touch) return
    const deltaY = touch.clientY - startY
    if (deltaY > 0) {
      // Damping resistance formula
      const damping = 0.5
      pullDistance.value = Math.min(deltaY * damping, maxDistance)
      isPulling.value = true
      if (pullDistance.value > 10 && e.cancelable) {
        e.preventDefault()
      }
    } else {
      pullDistance.value = 0
      isPulling.value = false
      tracking = false
    }
  }

  async function onTouchEnd() {
    if (!tracking) return
    tracking = false
    isPulling.value = false

    if (pullDistance.value >= threshold && !isRefreshing.value) {
      isRefreshing.value = true
      pullDistance.value = threshold
      try {
        await options.onRefresh()
      } catch {
        // refresh failed, swallow to avoid unhandled rejection
      } finally {
        isRefreshing.value = false
        pullDistance.value = 0
      }
    } else {
      pullDistance.value = 0
    }
  }

  function attachListeners(el: HTMLElement) {
    if (attachedElement) detachListeners(attachedElement)
    attachedElement = el
    el.addEventListener('touchstart', onTouchStart, { passive: true })
    el.addEventListener('touchmove', onTouchMove, { passive: false })
    el.addEventListener('touchend', onTouchEnd, { passive: true })
    el.addEventListener('touchcancel', onTouchEnd, { passive: true })
  }

  function detachListeners(el: HTMLElement) {
    el.removeEventListener('touchstart', onTouchStart)
    el.removeEventListener('touchmove', onTouchMove)
    el.removeEventListener('touchend', onTouchEnd)
    el.removeEventListener('touchcancel', onTouchEnd)
  }

  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      if (attachedElement) {
        detachListeners(attachedElement)
        attachedElement = null
      }
    })
  }

  return {
    pullDistance,
    isPulling,
    isRefreshing,
    attachListeners,
    detachListeners,
  }
}
