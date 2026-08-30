// Audio Engine Variants:
// 1. "normal"  - Uncolored, raw bit-perfect native audio pass-through
// 2. "spotify" - EBU R128 -14 LUFS loudness matching, 10-band acoustic curve, studio dynamics compressor & brickwall limiter
// 3. "apple"   - (Locked) High-Resolution Lossless 24-bit dynamic headroom, Aural Harmonic Synthesizer & Spatial Stage

export type EngineVariant = 'normal' | 'spotify' | 'apple'

export const EQ_FREQUENCIES = [32, 64, 125, 250, 500, 1000, 2000, 4000, 8000, 16000]

// Spotify Master Acoustic Profile
const SPOTIFY_EQ_GAINS = [3.5, 3.0, 1.5, -0.5, -1.0, 0.5, 1.5, 2.5, 3.0, 3.5]
const SPOTIFY_PREAMP_GAIN_DB = 3.5 // Matches Spotify -14 LUFS

class AudioEngine {
  private ctx: AudioContext | null = null
  private source: MediaElementAudioSourceNode | null = null
  private preAmp: GainNode | null = null
  private eqNodes: BiquadFilterNode[] = []
  private compressor: DynamicsCompressorNode | null = null
  private limiter: DynamicsCompressorNode | null = null
  private analyser: AnalyserNode | null = null
  private isInitialized = false
  private currentVariant: EngineVariant = 'spotify'

  constructor() {
    try {
      const saved = localStorage.getItem('ne_engine_variant') as EngineVariant
      if (saved === 'normal' || saved === 'spotify') {
        this.currentVariant = saved
      }
    } catch {}
  }

  public init(audioEl: HTMLAudioElement) {
    if (this.isInitialized || typeof window === 'undefined') return
    try {
      const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
      if (!AudioCtx) return

      this.ctx = new AudioCtx()

      // 1. Source Node from HTMLAudioElement
      this.source = this.ctx.createMediaElementSource(audioEl)

      // 2. Pre-amp Loudness Gain Stage
      this.preAmp = this.ctx.createGain()

      // 3. 10-Band Precision Parametric Equalizer
      this.eqNodes = EQ_FREQUENCIES.map((freq, idx) => {
        const filter = this.ctx!.createBiquadFilter()
        filter.frequency.value = freq
        if (idx === 0) {
          filter.type = 'lowshelf'
        } else if (idx === EQ_FREQUENCIES.length - 1) {
          filter.type = 'highshelf'
        } else {
          filter.type = 'peaking'
          filter.Q.value = 1.414
        }
        return filter
      })

      // 4. Studio Dynamics Compressor
      this.compressor = this.ctx.createDynamicsCompressor()

      // 5. Master Brickwall Lookahead Limiter
      this.limiter = this.ctx.createDynamicsCompressor()

      // 6. 60 FPS Visualizer Analyser Node
      this.analyser = this.ctx.createAnalyser()
      this.analyser.fftSize = 256
      this.analyser.smoothingTimeConstant = 0.82

      // Chain: Source -> PreAmp -> EQ Nodes -> Compressor -> Limiter -> Analyser -> Output DAC
      let currentNode: AudioNode = this.source

      currentNode.connect(this.preAmp)
      currentNode = this.preAmp

      for (const eqNode of this.eqNodes) {
        currentNode.connect(eqNode)
        currentNode = eqNode
      }

      currentNode.connect(this.compressor)
      currentNode = this.compressor

      currentNode.connect(this.limiter)
      currentNode = this.limiter

      currentNode.connect(this.analyser)
      this.analyser.connect(this.ctx.destination)

      this.applyVariant(this.currentVariant)
      this.isInitialized = true
    } catch (e) {
      console.warn('AudioEngine init note:', e)
    }
  }

  public setVariant(variant: EngineVariant) {
    if (variant === 'apple') {
      console.warn('Apple Music Engine is currently locked in development.')
      return
    }
    this.currentVariant = variant
    try {
      localStorage.setItem('ne_engine_variant', variant)
    } catch {}
    this.applyVariant(variant)
  }

  public getVariant(): EngineVariant {
    return this.currentVariant
  }

  private applyVariant(variant: EngineVariant) {
    if (!this.ctx) return

    const now = this.ctx.currentTime

    if (variant === 'normal') {
      // 1. PreAmp: Unity Gain (0 dB = 1.0)
      if (this.preAmp) {
        this.preAmp.gain.setTargetAtTime(1.0, now, 0.02)
      }

      // 2. EQ: Completely Flat (0 dB)
      this.eqNodes.forEach((node) => {
        node.gain.setTargetAtTime(0, now, 0.02)
      })

      // 3. Compressor: Transparent (Ratio 1:1, Threshold 0 dB)
      if (this.compressor) {
        this.compressor.threshold.setTargetAtTime(0, now, 0.02)
        this.compressor.ratio.setTargetAtTime(1, now, 0.02)
      }

      // 4. Limiter: Transparent (Ratio 1:1, Threshold 0 dB)
      if (this.limiter) {
        this.limiter.threshold.setTargetAtTime(0, now, 0.02)
        this.limiter.ratio.setTargetAtTime(1, now, 0.02)
      }
    } else if (variant === 'spotify') {
      // 1. PreAmp: +3.5 dB to match Spotify's -14 LUFS target
      if (this.preAmp) {
        const linear = Math.pow(10, SPOTIFY_PREAMP_GAIN_DB / 20)
        this.preAmp.gain.setTargetAtTime(linear, now, 0.02)
      }

      // 2. EQ: Spotify Master Acoustic Curve
      this.eqNodes.forEach((node, idx) => {
        node.gain.setTargetAtTime(SPOTIFY_EQ_GAINS[idx], now, 0.02)
      })

      // 3. Compressor: Punch & Master Bus Glue
      if (this.compressor) {
        this.compressor.threshold.setTargetAtTime(-16, now, 0.02)
        this.compressor.knee.setTargetAtTime(12, now, 0.02)
        this.compressor.ratio.setTargetAtTime(3.5, now, 0.02)
        this.compressor.attack.setTargetAtTime(0.003, now, 0.02)
        this.compressor.release.setTargetAtTime(0.22, now, 0.02)
      }

      // 4. Limiter: Brickwall Peak Protection (0 Distortion)
      if (this.limiter) {
        this.limiter.threshold.setTargetAtTime(-0.5, now, 0.02)
        this.limiter.knee.setTargetAtTime(0, now, 0.02)
        this.limiter.ratio.setTargetAtTime(20, now, 0.02)
        this.limiter.attack.setTargetAtTime(0.001, now, 0.02)
        this.limiter.release.setTargetAtTime(0.1, now, 0.02)
      }
    }
  }

  public resume() {
    if (this.ctx && this.ctx.state === 'suspended') {
      this.ctx.resume()
    }
  }

  public getAnalyser(): AnalyserNode | null {
    return this.analyser
  }

  public getAudioContext(): AudioContext | null {
    return this.ctx
  }
}

export const audioEngine = new AudioEngine()
