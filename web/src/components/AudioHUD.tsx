import React, { useEffect, useRef, useState } from 'react'
import { motion } from 'motion/react'
import { X, Play, Pause, RotateCcw, RotateCw } from 'lucide-react'
import { usePlayerStore } from '../store/playerStore'
import { audioEngine } from '../lib/audioEngine'

interface AudioHUDProps {
  className?: string
  compact?: boolean
  onClose?: () => void
}

export const AudioHUD: React.FC<AudioHUDProps> = ({ className = '', compact = false, onClose }) => {
  const { currentTrack, isPlaying, currentTime, duration, engineVariant, togglePlay, seek } = usePlayerStore()
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const animFrameId = useRef<number | null>(null)
  const [isDragging, setIsDragging] = useState(false)

  useEffect(() => {
    if (isPlaying) {
      audioEngine.resume()
    }
  }, [isPlaying])

  const formatPreciseTime = (seconds: number) => {
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    const ms = Math.floor((seconds % 1) * 100)
    return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs
      .toString()
      .padStart(2, '0')}.${ms.toString().padStart(2, '0')}`
  }

  // Polar Coordinate Scrubber Math
  const handleScrubAtPoint = (clientX: number, clientY: number) => {
    const canvas = canvasRef.current
    if (!canvas || !duration) return

    const rect = canvas.getBoundingClientRect()
    const centerX = rect.left + rect.width / 2
    const centerY = rect.top + rect.height / 2

    const dx = clientX - centerX
    const dy = clientY - centerY
    const dist = Math.hypot(dx, dy)

    // Center play/pause toggle click area
    const centerHitRadius = (rect.width / 2) * 0.28
    if (dist <= centerHitRadius) {
      togglePlay()
      return
    }

    // Polar angle: -PI/2 is top (0 radians)
    let angle = Math.atan2(dy, dx) + Math.PI / 2
    if (angle < 0) angle += 2 * Math.PI

    const progress = angle / (2 * Math.PI)
    const targetSeconds = progress * duration
    seek(Math.min(duration, Math.max(0, targetSeconds)))
  }

  const handlePointerDown = (e: React.PointerEvent<HTMLCanvasElement>) => {
    setIsDragging(true)
    handleScrubAtPoint(e.clientX, e.clientY)
  }

  const handlePointerMove = (e: React.PointerEvent<HTMLCanvasElement>) => {
    if (!isDragging) return
    handleScrubAtPoint(e.clientX, e.clientY)
  }

  const handlePointerUp = () => {
    setIsDragging(false)
  }

  // 60 FPS Visualizer & Interactive Dial Rendering
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const render = () => {
      const width = canvas.width
      const height = canvas.height
      const centerX = width / 2
      const centerY = height / 2
      const maxRadius = width * 0.44
      const progress = duration > 0 ? currentTime / duration : 0

      ctx.clearRect(0, 0, width, height)

      // 1. Fetch Real-time Audio Frequency Data from Built-in Studio DSP Engine
      const analyser = audioEngine.getAnalyser()
      let freqData = new Uint8Array(64)
      if (analyser && isPlaying) {
        freqData = new Uint8Array(analyser.frequencyBinCount)
        analyser.getByteFrequencyData(freqData)
      }

      // 2. Concentric Radar Rings
      const ringRadii = [0.22, 0.44, 0.68, 0.90, 1.0]
      ringRadii.forEach((factor, idx) => {
        ctx.beginPath()
        ctx.arc(centerX, centerY, maxRadius * factor, 0, Math.PI * 2)
        ctx.strokeStyle = idx === ringRadii.length - 1 ? 'rgba(255, 255, 255, 0.16)' : 'rgba(255, 255, 255, 0.05)'
        ctx.lineWidth = 1
        ctx.stroke()
      })

      // 3. Compass Crosshairs
      ctx.beginPath()
      ctx.moveTo(centerX, centerY - maxRadius * 1.05)
      ctx.lineTo(centerX, centerY + maxRadius * 1.05)
      ctx.moveTo(centerX - maxRadius * 1.05, centerY)
      ctx.lineTo(centerX + maxRadius * 1.05, centerY)
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)'
      ctx.lineWidth = 1
      ctx.stroke()

      // 4. Circular Audio Reactive Spectrum Bars
      const numTicks = 64
      for (let i = 0; i < numTicks; i++) {
        const angle = (i / numTicks) * Math.PI * 2 - Math.PI / 2
        const rawVal = isPlaying ? (freqData[i % freqData.length] || 0) / 255 : 0.08
        const dynamicLen = maxRadius * 0.12 + rawVal * (maxRadius * 0.28)

        const innerR = maxRadius * 0.72
        const outerR = innerR + dynamicLen

        const x1 = centerX + Math.cos(angle) * innerR
        const y1 = centerY + Math.sin(angle) * innerR
        const x2 = centerX + Math.cos(angle) * outerR
        const y2 = centerY + Math.sin(angle) * outerR

        const isScrubbed = i / numTicks <= progress

        ctx.beginPath()
        ctx.moveTo(x1, y1)
        ctx.lineTo(x2, y2)
        ctx.strokeStyle = isScrubbed
          ? 'rgba(255, 59, 48, 0.85)'
          : isPlaying
          ? 'rgba(255, 255, 255, 0.35)'
          : 'rgba(255, 255, 255, 0.12)'
        ctx.lineWidth = isScrubbed ? 2 : 1.5
        ctx.stroke()
      }

      // 5. Active Playhead Arc
      if (duration > 0) {
        ctx.beginPath()
        ctx.arc(centerX, centerY, maxRadius * 0.98, -Math.PI / 2, -Math.PI / 2 + progress * Math.PI * 2)
        ctx.strokeStyle = '#ff3b30'
        ctx.lineWidth = 2.5
        ctx.stroke()

        // Pointer Needle
        const needleAngle = -Math.PI / 2 + progress * Math.PI * 2
        const needleX = centerX + Math.cos(needleAngle) * (maxRadius * 0.98)
        const needleY = centerY + Math.sin(needleAngle) * (maxRadius * 0.98)

        ctx.beginPath()
        ctx.arc(needleX, needleY, 4, 0, Math.PI * 2)
        ctx.fillStyle = '#ffffff'
        ctx.fill()
        ctx.strokeStyle = '#ff3b30'
        ctx.lineWidth = 1.5
        ctx.stroke()
      }

      // 6. Center Tactical Recording Dot (Play / Pause Indicator)
      ctx.beginPath()
      ctx.arc(centerX, centerY, 7, 0, Math.PI * 2)
      ctx.fillStyle = isPlaying ? '#ff3b30' : 'rgba(255, 255, 255, 0.4)'
      ctx.fill()

      if (isPlaying) {
        ctx.beginPath()
        ctx.arc(centerX, centerY, 13, 0, Math.PI * 2)
        ctx.strokeStyle = 'rgba(255, 59, 48, 0.4)'
        ctx.lineWidth = 1.5
        ctx.stroke()
      }

      animFrameId.current = requestAnimationFrame(render)
    }

    render()

    return () => {
      if (animFrameId.current) cancelAnimationFrame(animFrameId.current)
    }
  }, [isPlaying, duration, currentTime])

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      className={`liquid-glass rounded-[28px] p-5 shadow-2xl relative select-none font-hud ${className}`}
    >
      {/* Top Telemetry Header */}
      <div className="flex items-center justify-between border-b border-white/[0.06] pb-3 mb-4 text-xs">
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${isPlaying ? 'bg-rose-500 animate-pulse' : 'bg-zinc-600'}`} />
          <span className="text-white font-bold tracking-widest">
            {isPlaying ? 'LIVE STREAM' : 'STANDBY'}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <span className={`px-2 py-0.5 rounded-full text-[9px] font-bold ${
            engineVariant === 'apple'
              ? 'bg-white/20 text-white border border-white/40 shadow-[0_0_10px_rgba(255,255,255,0.2)]'
              : engineVariant === 'spotify' 
              ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30' 
              : 'liquid-glass-pill text-zinc-400'
          }`}>
            {engineVariant === 'apple' ? 'HYPERION PRO 24-BIT' : engineVariant === 'spotify' ? 'BROADCAST ENGINE' : 'DIRECT BIT-PERFECT'}
          </span>

          {onClose && (
            <button
              onClick={onClose}
              className="p-1 text-zinc-400 hover:text-white rounded-full liquid-glass-pill transition-colors cursor-pointer"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Interactive Circular Jog Wheel & Radar Canvas */}
      <div className="relative flex flex-col items-center justify-center my-2">
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
          {isPlaying ? 'Master stream active.' : 'Stream idle. Waiting for playback...'}
        </p>
        <p className="text-xs text-zinc-400 font-normal mt-0.5">
          {isPlaying ? `Playing: ${currentTrack?.title || 'Audio'}` : 'Audio engine standby.'}
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
