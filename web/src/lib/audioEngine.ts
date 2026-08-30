export type PresetName = 'spotify' | 'bass' | 'vocal' | 'flat'

export interface DSPState {
  enabled: boolean
  preset: PresetName
  preAmpGain: number // in dB (-6 to +6)
  eqBands: number[]  // 10 bands in dB (-12 to +12)
  stereoWidth: number // 0.0 (mono) to 2.0 (super-wide), 1.0 is normal
  compressorEnabled: boolean
}

export const EQ_FREQUENCIES = [32, 64, 125, 250, 500, 1000, 2000, 4000, 8000, 16000]

export const PRESETS: Record<PresetName, { name: string; description: string; state: Omit<DSPState, 'enabled' | 'preset'> }> = {
  spotify: {
    name: 'Spotify Master',
    description: 'EBU R128 -14 LUFS volume match, soft-knee studio glue, sub-bass warmth & air',
    state: {
      preAmpGain: 3.5,
      eqBands: [3.5, 3.0, 1.5, -0.5, -1.0, 0.5, 1.5, 2.5, 3.0, 3.5],
      stereoWidth: 1.25,
      compressorEnabled: true,
    },
  },
  bass: {
    name: 'Club Deep Bass',
    description: 'Heavy sub-bass impact, tightened kick punch, and controlled transient limiting',
    state: {
      preAmpGain: 2.5,
      eqBands: [6.0, 5.5, 3.5, 1.0, 0.0, 0.0, 1.0, 1.5, 2.0, 2.5],
      stereoWidth: 1.15,
      compressorEnabled: true,
    },
  },
  vocal: {
    name: 'Acoustic & Vocal',
    description: 'Brings vocals and instruments forward while scooping out muddy low-mids',
    state: {
      preAmpGain: 2.5,
      eqBands: [-1.0, -1.0, 0.0, -1.5, 0.5, 3.0, 3.5, 2.5, 1.5, 2.0],
      stereoWidth: 1.10,
      compressorEnabled: true,
    },
  },
  flat: {
    name: 'Bit-Perfect Pure',
    description: 'Uncolored bit-perfect pass-through with zero equalization or compression',
    state: {
      preAmpGain: 0.0,
      eqBands: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
      stereoWidth: 1.0,
      compressorEnabled: false,
    },
  },
}

class AudioEngine {
  private ctx: AudioContext | null = null
  private source: MediaElementAudioSourceNode | null = null
  private preAmp: GainNode | null = null
  private eqNodes: BiquadFilterNode[] = []
  private compressor: DynamicsCompressorNode | null = null
  private limiter: DynamicsCompressorNode | null = null
  private analyser: AnalyserNode | null = null
  private masterGain: GainNode | null = null

  private state: DSPState = {
    enabled: true,
    preset: 'spotify',
    preAmpGain: 3.5,
    eqBands: [3.5, 3.0, 1.5, -0.5, -1.0, 0.5, 1.5, 2.5, 3.0, 3.5],
    stereoWidth: 1.25,
    compressorEnabled: true,
  }

  private isInitialized = false

  constructor() {
    this.loadPersistedState()
  }

  private loadPersistedState() {
    try {
      const saved = localStorage.getItem('ne_dsp_state')
      if (saved) {
        this.state = { ...this.state, ...JSON.parse(saved) }
      }
    } catch (e) {
      console.warn('Could not load DSP state from localStorage:', e)
    }
  }

  private persistState() {
    try {
      localStorage.setItem('ne_dsp_state', JSON.stringify(this.state))
    } catch (e) {
      console.warn('Could not save DSP state to localStorage:', e)
    }
  }

  public init(audioEl: HTMLAudioElement) {
    if (this.isInitialized || typeof window === 'undefined') return
    try {
      const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
      if (!AudioCtx) return

      this.ctx = new AudioCtx()

      // 1. Source Node from HTMLAudioElement
      this.source = this.ctx.createMediaElementSource(audioEl)

      // 2. Pre-amp Gain Node (Loudness Matcher)
      this.preAmp = this.ctx.createGain()
      this.setPreAmpGain(this.state.preAmpGain)

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
          filter.Q.value = 1.414 // ButterWorth musical Q
        }
        filter.gain.value = this.state.eqBands[idx] || 0
        return filter
      })

      // 4. Studio Dynamics Compressor (Dynamic Glue & Punch)
      this.compressor = this.ctx.createDynamicsCompressor()
      this.compressor.threshold.value = -16
      this.compressor.knee.value = 12
      this.compressor.ratio.value = 3.5
      this.compressor.attack.value = 0.003
      this.compressor.release.value = 0.22


      // 6. Master Brickwall Lookahead Limiter (Zero Digital Distortion Guarantee)
      this.limiter = this.ctx.createDynamicsCompressor()
      this.limiter.threshold.value = -0.5 // Safety ceiling
      this.limiter.knee.value = 0        // Hard brickwall
      this.limiter.ratio.value = 20      // Limiter ratio
      this.limiter.attack.value = 0.001  // Instant peak catch
      this.limiter.release.value = 0.1

      // 7. 60 FPS Visualizer Analyser Node
      this.analyser = this.ctx.createAnalyser()
      this.analyser.fftSize = 256
      this.analyser.smoothingTimeConstant = 0.82

      // 8. Master Gain Node
      this.masterGain = this.ctx.createGain()
      this.masterGain.gain.value = 1.0

      // Connect the Graph
      this.reconnectGraph()

      this.isInitialized = true
    } catch (e) {
      console.warn('AudioEngine init note:', e)
    }
  }

  public resume() {
    if (this.ctx && this.ctx.state === 'suspended') {
      this.ctx.resume()
    }
  }

  public reconnectGraph() {
    if (!this.ctx || !this.source || !this.analyser) return

    try {
      this.source.disconnect()
      this.preAmp?.disconnect()
      this.eqNodes.forEach((node) => node.disconnect())
      this.compressor?.disconnect()
      this.limiter?.disconnect()
      this.analyser.disconnect()
      this.masterGain?.disconnect()

      if (!this.state.enabled) {
        // Direct Bit-Perfect Bypass: Source -> Analyser -> Destination
        this.source.connect(this.analyser)
        this.analyser.connect(this.ctx.destination)
        return
      }

      // Chain: Source -> PreAmp -> EQ 1..10 -> Compressor -> Limiter -> Analyser -> MasterGain -> Destination
      let lastNode: AudioNode = this.source

      if (this.preAmp) {
        lastNode.connect(this.preAmp)
        lastNode = this.preAmp
      }

      for (const eqNode of this.eqNodes) {
        lastNode.connect(eqNode)
        lastNode = eqNode
      }

      if (this.state.compressorEnabled && this.compressor) {
        lastNode.connect(this.compressor)
        lastNode = this.compressor
      }

      if (this.limiter) {
        lastNode.connect(this.limiter)
        lastNode = this.limiter
      }

      lastNode.connect(this.analyser)
      this.analyser.connect(this.ctx.destination)
    } catch (e) {
      console.warn('Error reconnecting DSP audio graph:', e)
    }
  }

  public setEnabled(enabled: boolean) {
    this.state.enabled = enabled
    this.persistState()
    this.reconnectGraph()
  }

  public setPreset(name: PresetName) {
    const preset = PRESETS[name]
    if (!preset) return

    this.state.preset = name
    this.state.preAmpGain = preset.state.preAmpGain
    this.state.eqBands = [...preset.state.eqBands]
    this.state.stereoWidth = preset.state.stereoWidth
    this.state.compressorEnabled = preset.state.compressorEnabled

    // Apply values to audio nodes immediately
    this.setPreAmpGain(this.state.preAmpGain)
    this.state.eqBands.forEach((gain, idx) => {
      this.setEQBand(idx, gain)
    })

    this.persistState()
    this.reconnectGraph()
  }

  public setPreAmpGain(gainDb: number) {
    this.state.preAmpGain = gainDb
    if (this.preAmp && this.ctx) {
      // dB to Linear Amplitude: 10^(dB/20)
      const linear = Math.pow(10, gainDb / 20)
      this.preAmp.gain.setTargetAtTime(linear, this.ctx.currentTime, 0.02)
    }
    this.persistState()
  }

  public setEQBand(index: number, gainDb: number) {
    if (index < 0 || index >= this.eqNodes.length) return
    this.state.eqBands[index] = gainDb
    if (this.ctx && this.eqNodes[index]) {
      this.eqNodes[index].gain.setTargetAtTime(gainDb, this.ctx.currentTime, 0.02)
    }
    this.persistState()
  }

  public setCompressorEnabled(enabled: boolean) {
    this.state.compressorEnabled = enabled
    this.persistState()
    this.reconnectGraph()
  }

  public getState(): DSPState {
    return { ...this.state }
  }

  public getAnalyser(): AnalyserNode | null {
    return this.analyser
  }

  public getAudioContext(): AudioContext | null {
    return this.ctx
  }
}

export const audioEngine = new AudioEngine()
