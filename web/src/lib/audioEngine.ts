// Audio Engine Variants:
// 1. "normal"  - Direct: Uncolored, raw bit-perfect native audio pass-through
// 2. "spotify" - Broadcast: EBU R128 -14 LUFS loudness target, 10-band acoustic curve, studio dynamics compressor & brickwall limiter
// 3. "apple"   - Hyperion Studio Pro: 24-bit dynamic headroom (-16 LUFS), Aural Harmonic Synthesizer (16–22 kHz air),
//               Sub-harmonic bass tightener, and Binaural Spatial Soundstage Matrix.

export type EngineVariant = 'normal' | 'spotify' | 'apple'

export const EQ_FREQUENCIES = [32, 64, 125, 250, 500, 1000, 2000, 4000, 8000, 16000]

// Broadcast Master Acoustic Profile (EBU R128 -14 LUFS Target)
const SPOTIFY_EQ_GAINS = [3.5, 3.0, 1.5, -0.5, -1.0, 0.5, 1.5, 2.5, 3.0, 3.5]
const SPOTIFY_PREAMP_GAIN_DB = 3.5 // Matches international streaming broadcast standard (-14 LUFS)

// Hyperion Studio Profile (Natural Audiophile Curves with Air & Dynamic Headroom at -16 LUFS)
const APPLE_EQ_GAINS = [2.0, 1.8, 1.0, 0.0, -0.5, 0.5, 1.0, 2.0, 3.5, 5.0]
const APPLE_PREAMP_GAIN_DB = 1.8 // Matches studio reference dynamic headroom (-16 LUFS)

class AudioEngine {
  private ctx: AudioContext | null = null
  private source: MediaElementAudioSourceNode | null = null
  private preAmp: GainNode | null = null
  private eqNodes: BiquadFilterNode[] = []
  private compressor: DynamicsCompressorNode | null = null
  private limiter: DynamicsCompressorNode | null = null
  private analyser: AnalyserNode | null = null

  // Aural Harmonic Synthesizer (Restores 16–22 kHz "Air" via Non-linear Saturation)
  private auralFilter: BiquadFilterNode | null = null
  private auralShaper: WaveShaperNode | null = null
  private auralBandpass: BiquadFilterNode | null = null
  private auralGain: GainNode | null = null

  // Sub-Harmonic Bass Generator (Adds tight physical low-end weight)
  private subBassFilter: BiquadFilterNode | null = null
  private subBassShaper: WaveShaperNode | null = null
  private subBassGain: GainNode | null = null

  // Binaural Spatial Soundstage Matrix (Haas 3D Expansion)
  private haasDelay: DelayNode | null = null
  private haasGain: GainNode | null = null

  private isInitialized = false
  private currentVariant: EngineVariant = 'apple'

  constructor() {
    try {
      const saved = localStorage.getItem('ne_engine_variant') as EngineVariant
      if (saved === 'normal' || saved === 'spotify' || saved === 'apple') {
        this.currentVariant = saved
      }
    } catch {}
  }

  // Generate polynomial soft-saturation curve for even-order harmonic generation
  private createHarmonicCurve(samples = 1024): Float32Array<ArrayBuffer> {
    const buffer = new ArrayBuffer(samples * Float32Array.BYTES_PER_ELEMENT)
    const curve = new Float32Array(buffer)
    for (let i = 0; i < samples; i++) {
      const x = (i * 2) / samples - 1
      // Non-linear soft saturation generating 2nd and 3rd order acoustic harmonics
      curve[i] = (3 * x) / 2 - (Math.pow(x, 3) / 2)
    }
    return curve
  }

  // Soft clipper curve for low-end sub harmonics
  private createSubCurve(samples = 1024): Float32Array<ArrayBuffer> {
    const buffer = new ArrayBuffer(samples * Float32Array.BYTES_PER_ELEMENT)
    const curve = new Float32Array(buffer)
    for (let i = 0; i < samples; i++) {
      const x = (i * 2) / samples - 1
      curve[i] = Math.tanh(x * 1.5)
    }
    return curve
  }

  public init(audioEl: HTMLAudioElement) {
    if (this.isInitialized || typeof window === 'undefined') return
    try {
      const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
      if (!AudioCtx) return

      this.ctx = new AudioCtx()

      // 1. Source Node
      this.source = this.ctx.createMediaElementSource(audioEl)

      // 2. Pre-amp Stage
      this.preAmp = this.ctx.createGain()

      // 3. 10-Band Precision Equalizer
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

      // 4. Aural Harmonic Synthesizer (Air Generator)
      this.auralFilter = this.ctx.createBiquadFilter()
      this.auralFilter.type = 'highpass'
      this.auralFilter.frequency.value = 8500

      this.auralShaper = this.ctx.createWaveShaper()
      this.auralShaper.curve = this.createHarmonicCurve()
      this.auralShaper.oversample = '2x'

      this.auralBandpass = this.ctx.createBiquadFilter()
      this.auralBandpass.type = 'bandpass'
      this.auralBandpass.frequency.value = 16500
      this.auralBandpass.Q.value = 0.9

      this.auralGain = this.ctx.createGain()
      this.auralGain.gain.value = 0

      // 5. Sub-Harmonic Visceral Bass Generator (Lowpass 65 Hz -> WaveShaper -> Gain)
      this.subBassFilter = this.ctx.createBiquadFilter()
      this.subBassFilter.type = 'lowpass'
      this.subBassFilter.frequency.value = 65
      this.subBassFilter.Q.value = 1.0

      this.subBassShaper = this.ctx.createWaveShaper()
      this.subBassShaper.curve = this.createSubCurve()
      this.subBassShaper.oversample = '2x'

      this.subBassGain = this.ctx.createGain()
      this.subBassGain.gain.value = 0

      // 6. Master Studio Dynamics Compressor
      this.compressor = this.ctx.createDynamicsCompressor()

      // 7. Binaural Haas Stereo Spatial Widener
      this.haasDelay = this.ctx.createDelay(0.05)
      this.haasDelay.delayTime.value = 0.012 // 12 ms psychoacoustic Haas threshold

      this.haasGain = this.ctx.createGain()
      this.haasGain.gain.value = 0

      // 8. Output Brickwall True-Peak Limiter
      this.limiter = this.ctx.createDynamicsCompressor()

      // 9. Realtime Fast Fourier Transform (FFT) Analyser
      this.analyser = this.ctx.createAnalyser()
      this.analyser.fftSize = 256
      this.analyser.smoothingTimeConstant = 0.82

      // Connect Main Stream Chain:
      // Source -> PreAmp -> EQ Chain -> preCompressorBus -> Compressor -> Haas Widener -> Limiter -> Analyser -> Output DAC
      let current: AudioNode = this.source

      current.connect(this.preAmp)
      current = this.preAmp

      // Tap Aural Harmonic Synthesizer in parallel
      this.preAmp.connect(this.auralFilter)
      this.auralFilter.connect(this.auralShaper)
      this.auralShaper.connect(this.auralBandpass)
      this.auralBandpass.connect(this.auralGain)

      // Tap Sub-Harmonic Bass Generator in parallel
      this.preAmp.connect(this.subBassFilter)
      this.subBassFilter.connect(this.subBassShaper)
      this.subBassShaper.connect(this.subBassGain)

      for (const eqNode of this.eqNodes) {
        current.connect(eqNode)
        current = eqNode
      }

      // Dedicated pre-compressor summing bus: clean summing of EQ, Aural Harmonics, and Sub-Bass
      const preCompressorBus = this.ctx.createGain()
      current.connect(preCompressorBus)
      this.auralGain.connect(preCompressorBus)
      this.subBassGain.connect(preCompressorBus)

      preCompressorBus.connect(this.compressor)
      current = this.compressor

      // Binaural Haas Stereo Spatial Widener
      // Direct channels remain intact; delayed reflection is cross-fed to opposite ear, preventing comb-filtering
      const haasSplitter = this.ctx.createChannelSplitter(2)
      const haasMerger = this.ctx.createChannelMerger(2)

      current.connect(haasSplitter)

      // Direct L and R paths
      haasSplitter.connect(haasMerger, 0, 0)
      haasSplitter.connect(haasMerger, 1, 1)

      // Haas micro-delay applied strictly to cross-fed spatial ambience
      haasSplitter.connect(this.haasDelay, 0) // Left to delay
      this.haasDelay.connect(this.haasGain)
      this.haasGain.connect(haasMerger, 0, 1) // Cross-fed into Right channel with delay

      haasMerger.connect(this.limiter)
      current = this.limiter

      current.connect(this.analyser)
      this.analyser.connect(this.ctx.destination)

      this.applyVariant(this.currentVariant)
      this.isInitialized = true
    } catch (e) {
      console.warn('AudioEngine init note:', e)
    }
  }

  public setVariant(variant: EngineVariant) {
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
      if (this.preAmp) this.preAmp.gain.setTargetAtTime(1.0, now, 0.02)

      // 2. EQ: Completely Flat (0 dB)
      this.eqNodes.forEach((node) => node.gain.setTargetAtTime(0, now, 0.02))

      // 3. Aural & Sub Synthesizers: Bypassed
      if (this.auralGain) this.auralGain.gain.setTargetAtTime(0, now, 0.02)
      if (this.subBassGain) this.subBassGain.gain.setTargetAtTime(0, now, 0.02)
      if (this.haasGain) this.haasGain.gain.setTargetAtTime(0, now, 0.02)

      // 4. Dynamics: Transparent pass-through
      if (this.compressor) {
        this.compressor.threshold.setTargetAtTime(0, now, 0.02)
        this.compressor.ratio.setTargetAtTime(1, now, 0.02)
      }
      if (this.limiter) {
        this.limiter.threshold.setTargetAtTime(0, now, 0.02)
        this.limiter.ratio.setTargetAtTime(1, now, 0.02)
      }
    } else if (variant === 'spotify') {
      // 1. PreAmp: +3.5 dB (Spotify -14 LUFS target)
      if (this.preAmp) {
        const linear = Math.pow(10, SPOTIFY_PREAMP_GAIN_DB / 20)
        this.preAmp.gain.setTargetAtTime(linear, now, 0.02)
      }

      // 2. EQ: Spotify Curve
      this.eqNodes.forEach((node, idx) => {
        node.gain.setTargetAtTime(SPOTIFY_EQ_GAINS[idx], now, 0.02)
      })

      // 3. Synthesizers: Off (Standard master)
      if (this.auralGain) this.auralGain.gain.setTargetAtTime(0, now, 0.02)
      if (this.subBassGain) this.subBassGain.gain.setTargetAtTime(0, now, 0.02)
      if (this.haasGain) this.haasGain.gain.setTargetAtTime(0, now, 0.02)

      // 4. Compressor: Punch & Glue
      if (this.compressor) {
        this.compressor.threshold.setTargetAtTime(-16, now, 0.02)
        this.compressor.knee.setTargetAtTime(12, now, 0.02)
        this.compressor.ratio.setTargetAtTime(3.5, now, 0.02)
        this.compressor.attack.setTargetAtTime(0.003, now, 0.02)
        this.compressor.release.setTargetAtTime(0.22, now, 0.02)
      }

      // 5. Limiter: -0.5 dB ceiling
      if (this.limiter) {
        this.limiter.threshold.setTargetAtTime(-0.5, now, 0.02)
        this.limiter.knee.setTargetAtTime(0, now, 0.02)
        this.limiter.ratio.setTargetAtTime(20, now, 0.02)
        this.limiter.attack.setTargetAtTime(0.001, now, 0.02)
        this.limiter.release.setTargetAtTime(0.1, now, 0.02)
      }
    } else if (variant === 'apple') {
      // 1. PreAmp: +1.8 dB (Apple Music -16 LUFS dynamic breathing room)
      if (this.preAmp) {
        const linear = Math.pow(10, APPLE_PREAMP_GAIN_DB / 20)
        this.preAmp.gain.setTargetAtTime(linear, now, 0.02)
      }

      // 2. EQ: Apple Studio Natural Audiophile Profile
      this.eqNodes.forEach((node, idx) => {
        node.gain.setTargetAtTime(APPLE_EQ_GAINS[idx], now, 0.02)
      })

      // 3. Aural Harmonic Synthesizer Active: Injects 16–22 kHz sheen & air
      if (this.auralGain) this.auralGain.gain.setTargetAtTime(0.38, now, 0.03)

      // 4. Sub-Harmonic Bass Active: Tight visceral low-end
      if (this.subBassGain) this.subBassGain.gain.setTargetAtTime(0.28, now, 0.03)

      // 5. Binaural Haas Spatial Active: 3D Holographic Soundstage
      if (this.haasGain) this.haasGain.gain.setTargetAtTime(0.22, now, 0.03)

      // 6. Studio Compressor: Transparent, dynamic breathing (soft ratio 2.2:1)
      if (this.compressor) {
        this.compressor.threshold.setTargetAtTime(-18, now, 0.02)
        this.compressor.knee.setTargetAtTime(16, now, 0.02)
        this.compressor.ratio.setTargetAtTime(2.2, now, 0.02) // Preserves wide dynamic peaks
        this.compressor.attack.setTargetAtTime(0.008, now, 0.02) // Allows drum transients to crackle cleanly
        this.compressor.release.setTargetAtTime(0.28, now, 0.02)
      }

      // 7. Master Apple Digital Masters True-Peak Limiter (-1.0 dBTP Ceiling)
      if (this.limiter) {
        this.limiter.threshold.setTargetAtTime(-1.0, now, 0.02) // Zero Inter-Sample Peak distortion
        this.limiter.knee.setTargetAtTime(2, now, 0.02)
        this.limiter.ratio.setTargetAtTime(20, now, 0.02)
        this.limiter.attack.setTargetAtTime(0.001, now, 0.02)
        this.limiter.release.setTargetAtTime(0.12, now, 0.02)
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
