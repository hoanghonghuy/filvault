class CallAudio {
  private ctx: AudioContext | null = null
  private ringTimer: number | null = null

  private getContext(): AudioContext | null {
    if (typeof window === 'undefined') return null
    try {
      if (!this.ctx || this.ctx.state === 'closed') {
        const AudioCtx =
          window.AudioContext ||
          (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
        if (!AudioCtx) return null
        this.ctx = new AudioCtx()
      }
      if (this.ctx.state === 'suspended') {
        void this.ctx.resume()
      }
      return this.ctx
    } catch {
      return null
    }
  }

  playIncomingRing() {
    this.stop()
    const playChime = () => {
      try {
        const ctx = this.getContext()
        if (!ctx) return
        const now = ctx.currentTime
        const freqs = [523.25, 659.25, 783.99, 1046.5] // C5, E5, G5, C6
        freqs.forEach((freq, idx) => {
          const osc = ctx.createOscillator()
          const gain = ctx.createGain()
          osc.type = 'sine'
          osc.frequency.setValueAtTime(freq, now + idx * 0.1)
          gain.gain.setValueAtTime(0.08, now + idx * 0.1)
          gain.gain.exponentialRampToValueAtTime(0.0001, now + idx * 0.1 + 0.4)
          osc.connect(gain)
          gain.connect(ctx.destination)
          osc.start(now + idx * 0.1)
          osc.stop(now + idx * 0.1 + 0.45)
        })
      } catch {
        // audio context blocked until user gesture
      }
    }
    playChime()
    if (typeof window !== 'undefined') {
      this.ringTimer = window.setInterval(playChime, 2500)
    }
  }

  playOutgoingRing() {
    this.stop()
    const playRingTone = () => {
      try {
        const ctx = this.getContext()
        if (!ctx) return
        const now = ctx.currentTime
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        osc.type = 'sine'
        osc.frequency.setValueAtTime(440, now)
        gain.gain.setValueAtTime(0.06, now)
        gain.gain.setValueAtTime(0.06, now + 1.2)
        gain.gain.exponentialRampToValueAtTime(0.0001, now + 1.3)
        osc.connect(gain)
        gain.connect(ctx.destination)
        osc.start(now)
        osc.stop(now + 1.35)
      } catch {
        // audio context blocked
      }
    }
    playRingTone()
    if (typeof window !== 'undefined') {
      this.ringTimer = window.setInterval(playRingTone, 3000)
    }
  }

  playEndCall() {
    this.stop()
    try {
      const ctx = this.getContext()
      if (!ctx) return
      const now = ctx.currentTime
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      osc.type = 'triangle'
      osc.frequency.setValueAtTime(320, now)
      osc.frequency.setValueAtTime(240, now + 0.15)
      gain.gain.setValueAtTime(0.08, now)
      gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.35)
      osc.connect(gain)
      gain.connect(ctx.destination)
      osc.start(now)
      osc.stop(now + 0.4)
    } catch {
      // ignore
    }
  }

  stop() {
    if (this.ringTimer !== null) {
      clearInterval(this.ringTimer)
      this.ringTimer = null
    }
  }
}

export const callAudio = new CallAudio()
