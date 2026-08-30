import React, { useEffect, useRef } from 'react'
import { motion } from 'motion/react'
import { usePlayerStore, getAudio } from '../store/playerStore'

interface AudioHUDProps {
  className?: string
  compact?: boolean
  onClose?: () => void
}

let audioCtx: AudioContext | null = null
let analyserNode: AnalyserNode | null = null
let sourceNode: MediaElementAudioSourceNode | null = null

export const AudioHUD: React.FC<AudioHUDProps> = ({ className = '', compact = false, onClose }) => {
  const { currentTrack, isPlaying, currentTime } = usePlayerStore()
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const animFrameId = useRef<number | null>(null)

  // Initialize Web Audio API Analyser
  useEffect(() => {
    try {
      const audioEl = getAudio()
      if (!audioCtx && typeof window !== 'undefined') {
        const AudioContextClass = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
        if (AudioContextClass) {
          audioCtx = new AudioContextClass()
          analyserNode = audioCtx.createAnalyser()
          analyserNode.fftSize = 256
          analyserNode.smoothingTimeConstant = 0.8
          try {
            sourceNode = audioCtx.createMediaElementSource(audioEl)
            sourceNode.connect(analyserNode)
            analyserNode.connect(audioCtx.destination)
          } catch {
            // Already connected or cross-origin restriction fallback
          }
        }
      }
      if (audioCtx && audioCtx.state === 'suspended' && isPlaying) {
        audioCtx.resume()
      }
    } catch (e) {
      console.warn('AudioContext initialization note:', e)
    }
  }, [isPlaying])

  // Format high-precision time: HH:MM:SS:MS (e.g. 00:00:16:63)
  const formatPreciseTime = (seconds: number) => {
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    const ms = Math.floor((seconds % 1) * 100)
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n)
    return `${pad(hrs)}:${pad(mins)}:${pad(secs)}:${pad(ms)}`
  }

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const totalTicks = 84
    let phase = 0
    const freqData = new Uint8Array(128)

    const render = () => {
      const width = canvas.width
      const height = canvas.height
      const centerX = width / 2
      const centerY = height / 2
      const baseRadius = width * 0.32

      ctx.clearRect(0, 0, width, height)

      // Fetch real audio frequency data if available
      let hasRealFreqs = false
      if (analyserNode && isPlaying) {
        analyserNode.getByteFrequencyData(freqData)
        hasRealFreqs = freqData.some((v) => v > 0)
      }

      // 1. Outer dashed concentric guide ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, baseRadius + 14, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)'
      ctx.lineWidth = 1
      ctx.setLineDash([2, 4])
      ctx.stroke()
      ctx.setLineDash([])

      // 2. Inner circular target ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, baseRadius - 12, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.12)'
      ctx.lineWidth = 1
      ctx.stroke()

      // 3. Center recording / live boundary ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, 16, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.18)'
      ctx.lineWidth = 1
      ctx.stroke()

      // 4. Render radial tick marks with audio waveform spikes
      phase += isPlaying ? 0.04 : 0.005

      for (let i = 0; i < totalTicks; i++) {
        const angle = (i / totalTicks) * Math.PI * 2 - Math.PI / 2
        const cos = Math.cos(angle)
        const sin = Math.sin(angle)

        let tickHeight = 4
        let alpha = 0.35

        if (isPlaying) {
          if (hasRealFreqs) {
            // Map radial tick to audio frequency spectrum
            const freqIdx = Math.floor((i / totalTicks) * 64)
            const rawVal = freqData[freqIdx] || 0
            const normalized = rawVal / 255
            tickHeight = 4 + normalized * 22
            alpha = 0.35 + normalized * 0.65
          } else {
            // Harmonic fallback
            const wave1 = Math.sin(i * 0.35 + phase)
            const wave2 = Math.cos(i * 0.7 - phase * 1.5)
            const combined = Math.max(0, (wave1 + wave2 * 0.7) / 1.7)
            tickHeight = 4 + combined * 18
            alpha = 0.4 + combined * 0.6
          }
        } else {
          if (i % 6 === 0) {
            tickHeight = 8
            alpha = 0.6
          }
        }

        const startR = baseRadius
        const endR = baseRadius + tickHeight

        const x1 = centerX + cos * startR
        const y1 = centerY + sin * startR
        const x2 = centerX + cos * endR
        const y2 = centerY + sin * endR

        ctx.beginPath()
        ctx.moveTo(x1, y1)
        ctx.lineTo(x2, y2)
        ctx.strokeStyle = `rgba(245, 243, 240, ${alpha})`
        ctx.lineWidth = 1.25
        ctx.stroke()
      }

      // 5. Center Red Live Status Dot (matching Image 1)
      ctx.beginPath()
      ctx.arc(centerX, centerY, 5.5, 0, Math.PI * 2)
      ctx.fillStyle = isPlaying ? '#ff3b30' : 'rgba(255, 59, 48, 0.4)'
      ctx.shadowColor = '#ff3b30'
      ctx.shadowBlur = isPlaying ? 14 : 0
      ctx.fill()
      ctx.shadowBlur = 0

      animFrameId.current = requestAnimationFrame(render)
    }

    render()

    return () => {
      if (animFrameId.current) {
        cancelAnimationFrame(animFrameId.current)
      }
    }
  }, [isPlaying])

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      className={`bg-[#151619] border border-[#2c2e35] rounded-2xl p-5 shadow-2xl flex flex-col justify-between select-none text-zinc-300 font-hud ${className}`}
    >
      {/* Top Header: W - 01 and AUDIO_IN matching Image 1 */}
      <div className="flex items-center justify-between text-[11px] tracking-widest text-zinc-400 border-b border-white/5 pb-3">
        <span className="font-semibold text-zinc-300">W - 01</span>
        <div className="flex items-center gap-2">
          <span className="inline-block w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse-dot" />
          <span className="text-zinc-400 font-medium">AUDIO_IN</span>
          {onClose && (
            <button
              onClick={onClose}
              className="ml-2 text-zinc-400 hover:text-white transition-colors cursor-pointer"
            >
              ✕
            </button>
          )}
        </div>
      </div>

      {/* Center Oscilloscope Radar Canvas */}
      <div className="relative flex items-center justify-center my-4">
        <canvas
          ref={canvasRef}
          width={compact ? 200 : 260}
          height={compact ? 200 : 260}
          className="w-full max-w-[260px] aspect-square"
        />
      </div>

      {/* Technical Status Block (Image 1 replica) */}
      <div className="border-t border-white/5 pt-3 mb-4">
        <span className="text-[10px] tracking-widest uppercase text-zinc-400 font-semibold block mb-1">
          STATUS
        </span>
        <p className="text-xs text-white font-medium tracking-tight">
          {isPlaying ? 'Secure link established.' : 'Stream idle. Waiting for playback...'}
        </p>
        <p className="text-xs text-zinc-400 font-normal mt-0.5">
          {isPlaying ? `Capturing input stream: ${currentTrack?.title || 'Audio'}` : 'Audio engine standby.'}
        </p>
      </div>

      {/* Footer Telemetry: Precise Time + Sample Rate */}
      <div className="flex items-center justify-between text-[11px] text-zinc-400 tracking-wider pt-2 border-t border-white/5">
        <span className="text-zinc-300 font-medium">{formatPreciseTime(currentTime)}</span>
        <span className="text-zinc-400">
          {currentTrack?.sampleRate ? `${(currentTrack.sampleRate / 1000).toFixed(1)} kHz` : '44.1 kHz'}
          <span className="text-zinc-400 ml-1.5">• {currentTrack?.format?.toUpperCase() || 'MP3'}</span>
        </span>
      </div>
    </motion.div>
  )
}
