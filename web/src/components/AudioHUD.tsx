import React, { useEffect, useRef, useState } from 'react'
import { motion } from 'motion/react'
import { RotateCcw, RotateCw, Play, Pause } from 'lucide-react'
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
  const { currentTrack, isPlaying, currentTime, duration, togglePlay, seek } = usePlayerStore()
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const animFrameId = useRef<number | null>(null)
  const [isDragging, setIsDragging] = useState(false)

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
            // Already connected fallback
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

  const formatPreciseTime = (seconds: number) => {
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    const ms = Math.floor((seconds % 1) * 100)
    const pad = (n: number) => (n < 10 ? '0' + n : '' + n)
    return `${pad(hrs)}:${pad(mins)}:${pad(secs)}:${pad(ms)}`
  }

  // Handle circular scrub interaction (jog wheel)
  const handleScrubAtPoint = (clientX: number, clientY: number) => {
    const canvas = canvasRef.current
    if (!canvas || !duration || duration <= 0) return

    const rect = canvas.getBoundingClientRect()
    const x = clientX - rect.left - rect.width / 2
    const y = clientY - rect.top - rect.height / 2

    // Check if clicked in center (within 24px) -> toggle play
    const dist = Math.sqrt(x * x + y * y)
    if (dist < 24) {
      togglePlay()
      return
    }

    // Polar angle calculation
    let angle = Math.atan2(y, x) + Math.PI / 2
    if (angle < 0) angle += Math.PI * 2

    const progressRatio = angle / (Math.PI * 2)
    seek(progressRatio * duration)
  }

  const handlePointerDown = (e: React.PointerEvent<HTMLCanvasElement>) => {
    setIsDragging(true)
    handleScrubAtPoint(e.clientX, e.clientY)
  }

  const handlePointerMove = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (isDragging) {
      handleScrubAtPoint(e.clientX, e.clientY)
    }
  }

  const handlePointerUp = () => {
    setIsDragging(false)
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

      let hasRealFreqs = false
      if (analyserNode && isPlaying) {
        analyserNode.getByteFrequencyData(freqData)
        hasRealFreqs = freqData.some((v) => v > 0)
      }

      // Calculate progress angle (0 to 2*PI starting at top -PI/2)
      const progressFraction = duration > 0 ? currentTime / duration : 0
      const currentProgressAngle = progressFraction * Math.PI * 2 - Math.PI / 2

      // 1. Outer dashed concentric guide ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, baseRadius + 15, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)'
      ctx.lineWidth = 1
      ctx.setLineDash([2, 4])
      ctx.stroke()
      ctx.setLineDash([])

      // 2. Progress Arc along the radar perimeter
      if (progressFraction > 0) {
        ctx.beginPath()
        ctx.arc(centerX, centerY, baseRadius + 15, -Math.PI / 2, currentProgressAngle)
        ctx.strokeStyle = '#ff3b30'
        ctx.lineWidth = 2
        ctx.stroke()
      }

      // 3. Inner circular target ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, baseRadius - 12, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.12)'
      ctx.lineWidth = 1
      ctx.stroke()

      // 4. Center interactive boundary ring
      ctx.beginPath()
      ctx.arc(centerX, centerY, 20, 0, Math.PI * 2)
      ctx.fillStyle = 'rgba(255, 255, 255, 0.04)'
      ctx.fill()
      ctx.strokeStyle = isPlaying ? 'rgba(255, 59, 48, 0.4)' : 'rgba(255, 255, 255, 0.18)'
      ctx.lineWidth = 1
      ctx.stroke()

      // 5. Render radial tick marks with audio waveform spikes
      phase += isPlaying ? 0.04 : 0.005

      for (let i = 0; i < totalTicks; i++) {
        const angle = (i / totalTicks) * Math.PI * 2 - Math.PI / 2
        const cos = Math.cos(angle)
        const sin = Math.sin(angle)

        let tickHeight = 4
        let alpha = 0.35

        if (isPlaying) {
          if (hasRealFreqs) {
            const freqIdx = Math.floor((i / totalTicks) * 64)
            const rawVal = freqData[freqIdx] || 0
            const normalized = rawVal / 255
            tickHeight = 4 + normalized * 22
            alpha = 0.35 + normalized * 0.65
          } else {
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

        // Highlight ticks behind the current scrubber position
        if (angle <= currentProgressAngle) {
          ctx.strokeStyle = `rgba(255, 255, 255, ${Math.min(1, alpha + 0.3)})`
        } else {
          ctx.strokeStyle = `rgba(255, 255, 255, ${alpha})`
        }

        ctx.lineWidth = 1.25
        ctx.stroke()
      }

      // 6. Playhead needle / indicator on the scrub ring
      const needleCos = Math.cos(currentProgressAngle)
      const needleSin = Math.sin(currentProgressAngle)
      ctx.beginPath()
      ctx.arc(
        centerX + needleCos * (baseRadius + 15),
        centerY + needleSin * (baseRadius + 15),
        3.5,
        0,
        Math.PI * 2
      )
      ctx.fillStyle = '#ff3b30'
      ctx.shadowColor = '#ff3b30'
      ctx.shadowBlur = 10
      ctx.fill()
      ctx.shadowBlur = 0

      // 7. Center Red Live Status / Play-Pause Dot (matching Image 1)
      ctx.beginPath()
      ctx.arc(centerX, centerY, 6.5, 0, Math.PI * 2)
      ctx.fillStyle = isPlaying ? '#ff3b30' : 'rgba(255, 255, 255, 0.7)'
      ctx.shadowColor = isPlaying ? '#ff3b30' : 'rgba(255, 255, 255, 0.4)'
      ctx.shadowBlur = isPlaying ? 16 : 8
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
  }, [isPlaying, currentTime, duration])

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      className={`liquid-glass rounded-2xl p-5 shadow-2xl flex flex-col justify-between select-none text-zinc-300 font-hud ${className}`}
    >
      {/* Top Header: W - 01 and AUDIO_IN matching Image 1 */}
      <div className="flex items-center justify-between text-[11px] tracking-widest text-zinc-400 border-b border-white/[0.06] pb-3">
        <span className="font-semibold text-white">W - 01</span>
        <div className="flex items-center gap-2">
          <span className="inline-block w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse-dot" />
          <span className="text-zinc-400 font-medium">AUDIO_IN</span>
          {onClose && (
            <button
              onClick={onClose}
              className="ml-2 text-zinc-500 hover:text-white transition-colors cursor-pointer"
            >
              ✕
            </button>
          )}
        </div>
      </div>

      {/* Center Interactive Oscilloscope Radar Jog Wheel */}
      <div className="relative flex flex-col items-center justify-center my-3">
        <canvas
          ref={canvasRef}
          width={compact ? 200 : 260}
          height={compact ? 200 : 260}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onPointerLeave={handlePointerUp}
          className="w-full max-w-[260px] aspect-square cursor-pointer active:cursor-grabbing touch-none"
          title="Jog Wheel: Drag/Click circle to scrub. Click center to Pause/Play."
        />

        {/* Quick Rewind / Forward Tactile Scrub Bar */}
        <div className="flex items-center justify-center gap-6 mt-1 text-xs">
          <button
            onClick={() => seek(Math.max(0, currentTime - 5))}
            title="Rewind 5s"
            className="flex items-center gap-1 px-3 py-1 rounded-lg liquid-glass-pill text-zinc-400 hover:text-white transition-colors cursor-pointer"
          >
            <RotateCcw className="w-3 h-3" /> -5s
          </button>

          <button
            onClick={togglePlay}
            title={isPlaying ? 'Pause' : 'Play'}
            className="p-2 rounded-full liquid-glass-pill text-white hover:bg-white/[0.1] transition-colors cursor-pointer"
          >
            {isPlaying ? <Pause className="w-3.5 h-3.5 fill-current" /> : <Play className="w-3.5 h-3.5 fill-current ml-0.5" />}
          </button>

          <button
            onClick={() => seek(Math.min(duration, currentTime + 5))}
            title="Fast Forward 5s"
            className="flex items-center gap-1 px-3 py-1 rounded-lg liquid-glass-pill text-zinc-400 hover:text-white transition-colors cursor-pointer"
          >
            +5s <RotateCw className="w-3 h-3" />
          </button>
        </div>
      </div>

      {/* Technical Status Block matching Image 1 */}
      <div className="border-t border-white/[0.06] pt-3 mb-3">
        <span className="text-[10px] tracking-widest uppercase text-zinc-500 font-semibold block mb-1">
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
      <div className="flex items-center justify-between text-[11px] text-zinc-400 tracking-wider pt-2 border-t border-white/[0.06]">
        <span className="text-white font-medium">{formatPreciseTime(currentTime)}</span>
        <span className="text-zinc-400">
          {currentTrack?.sampleRate ? `${(currentTrack.sampleRate / 1000).toFixed(1)} kHz` : '44.1 kHz'}
          <span className="text-zinc-500 ml-1.5">• {currentTrack?.format?.toUpperCase() || 'MP3'}</span>
        </span>
      </div>
    </motion.div>
  )
}
