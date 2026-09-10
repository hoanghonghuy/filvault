import { computed, onUnmounted, ref } from 'vue'

export function useResendCooldown(durationSeconds = 60) {
  const remainingSeconds = ref(0)
  let timer: ReturnType<typeof setInterval> | null = null

  const active = computed(() => remainingSeconds.value > 0)

  function stop() {
    if (timer !== null) {
      clearInterval(timer)
      timer = null
    }
  }

  function start() {
    stop()
    remainingSeconds.value = Math.max(0, Math.floor(durationSeconds))
    if (remainingSeconds.value === 0) return

    timer = setInterval(() => {
      remainingSeconds.value = Math.max(0, remainingSeconds.value - 1)
      if (remainingSeconds.value === 0) stop()
    }, 1000)
  }

  onUnmounted(stop)

  return {
    remainingSeconds,
    active,
    start,
    stop,
  }
}
