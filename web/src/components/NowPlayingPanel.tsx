import React, { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Shuffle,
  Repeat,
  Repeat1,
  Heart,
  Download,
  MoreHorizontal,
  Activity,
  Disc,
  Volume2,
  VolumeX,
  ListMusic,
  Sliders,
} from 'lucide-react'
import * as Slider from '@radix-ui/react-slider'
import * as Tooltip from '@radix-ui/react-tooltip'
import * as DropdownMenu from '@radix-ui/react-dropdown-menu'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'
import { AudioHUD } from './AudioHUD'
import { AudioDSPModal } from './AudioDSPModal'

type PanelMode = 'player' | 'queue' | 'hud'

export const NowPlayingPanel: React.FC = () => {
  const {
    currentTrack,
    isPlaying,
    currentTime,
    duration,
    volume,
    isMuted,
    isShuffle,
    repeatMode,
    queue,
    dspState,
    playTrack,
    togglePlay,
    next,
    previous,
    seek,
    setVolume,
    toggleMute,
    toggleShuffle,
    toggleRepeat,
    toggleStarCurrentTrack,
  } = usePlayerStore()

  const [mode, setMode] = useState<PanelMode>('player')
  const [showDSP, setShowDSP] = useState(false)

  const formatTime = (seconds: number) => {
    if (!seconds || isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`
  }

  const artworkUrl = currentTrack ? api.getArtworkUrl('track', currentTrack.id, 600) : ''

  return (
    <aside className="w-80 md:w-96 bg-black/85 backdrop-blur-2xl border-l border-white/[0.06] p-6 flex flex-col justify-between shrink-0 select-none h-full overflow-y-auto relative">
      {/* Top Segment Mode Switcher + DSP Studio Button */}
      <div className="flex items-center justify-between mb-5 pb-3 border-b border-white/[0.06]">
        <div className="flex items-center gap-1 liquid-glass-pill p-1 rounded-xl text-xs font-semibold">
          <button
            onClick={() => setMode('player')}
            className={`px-3 py-1 rounded-lg transition-all cursor-pointer ${
              mode === 'player' ? 'bg-white text-black shadow-md font-bold' : 'text-zinc-400 hover:text-white'
            }`}
          >
            Playing
          </button>
          <button
            onClick={() => setMode('queue')}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-lg transition-all cursor-pointer ${
              mode === 'queue' ? 'bg-white text-black shadow-md font-bold' : 'text-zinc-400 hover:text-white'
            }`}
          >
            <ListMusic className="w-3 h-3" /> Queue
          </button>
          <button
            onClick={() => setMode('hud')}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-lg transition-all cursor-pointer font-hud text-[11px] ${
              mode === 'hud' ? 'bg-white text-black shadow-md font-bold' : 'text-zinc-400 hover:text-white'
            }`}
          >
            <Activity className="w-3 h-3 text-rose-500" /> HUD
          </button>
        </div>

        {/* DSP Studio Trigger Button */}
        <button
          onClick={() => setShowDSP(true)}
          title="Studio DSP Master Engine (Equalizer & Dynamics)"
          className={`p-1.5 rounded-lg liquid-glass-pill transition-all cursor-pointer flex items-center gap-1 text-xs ${
            dspState.enabled ? 'text-rose-500 border-rose-500/40 shadow-lg' : 'text-zinc-400 hover:text-white'
          }`}
        >
          <Sliders className="w-4 h-4" />
        </button>
      </div>

      <AnimatePresence mode="wait">
        {mode === 'hud' && (
          <AudioHUD key="hud" onClose={() => setMode('player')} onOpenDSP={() => setShowDSP(true)} className="w-full my-auto" />
        )}

        {mode === 'queue' && (
          <motion.div
            key="queue"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 8 }}
            className="flex-1 flex flex-col justify-between min-h-0"
          >
            <div>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-bold text-white tracking-tight">Play Queue</h3>
                <span className="text-xs text-zinc-500 font-hud">{queue.length} songs</span>
              </div>
              <div className="space-y-1 overflow-y-auto max-h-[480px] pr-1">
                {queue.map((t, idx) => {
                  const isCurrent = currentTrack?.id === t.id
                  return (
                    <div
                      key={t.id}
                      onClick={() => playTrack(t, queue)}
                      className={`flex items-center gap-3 p-2.5 rounded-xl cursor-pointer transition-all ${
                        isCurrent
                          ? 'liquid-glass-pill text-white font-semibold'
                          : 'hover:bg-white/[0.04] text-zinc-400 hover:text-white'
                      }`}
                    >
                      <span className="w-5 text-center text-xs font-hud text-zinc-600">
                        {idx + 1}
                      </span>
                      <div className="min-w-0 flex-1">
                        <p className={`text-xs truncate ${isCurrent ? 'text-white font-bold' : 'text-zinc-300'}`}>
                          {t.title}
                        </p>
                        <p className="text-[11px] text-zinc-500 truncate">{t.rawArtist}</p>
                      </div>
                      <span className="text-[11px] font-hud text-zinc-500">
                        {formatTime(t.duration)}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
            <button
              onClick={() => setMode('player')}
              className="mt-4 w-full py-2.5 rounded-xl liquid-glass-pill text-xs text-white font-bold cursor-pointer hover:bg-white/[0.1] transition-colors"
            >
              Back to Player
            </button>
          </motion.div>
        )}

        {mode === 'player' && (
          <motion.div
            key="player"
            initial={{ opacity: 0, scale: 0.98 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.98 }}
            className="flex flex-col justify-between flex-1"
          >
            {/* Top: Large Vinyl Artwork Card with Liquid Glass Frame */}
            <div className="relative w-full aspect-square mb-6">
              <div className="relative w-full h-full rounded-[26px] overflow-hidden shadow-2xl bg-black border border-white/10 group">
                {artworkUrl ? (
                  <img
                    src={artworkUrl}
                    alt={currentTrack?.title || 'Album Art'}
                    className="w-full h-full object-cover"
                    onError={(e) => {
                      ;(e.target as HTMLElement).style.display = 'none'
                    }}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center bg-zinc-950">
                    <Disc className="w-24 h-24 text-zinc-800" />
                  </div>
                )}

                {/* Subtle top edge glossy reflection */}
                <div className="absolute inset-0 bg-gradient-to-b from-white/15 via-transparent to-black/70 pointer-events-none" />

                {/* HUD Radar Quick Trigger Icon */}
                <button
                  onClick={() => setMode('hud')}
                  title="Switch to Interactive Radar HUD"
                  className="absolute top-4 right-4 p-2.5 rounded-full liquid-glass-pill text-white transition-all shadow-lg cursor-pointer hover:scale-105"
                >
                  <Activity className="w-4 h-4 text-rose-500" />
                </button>
              </div>
            </div>

            {/* Middle: Track Title & Artist Subtitle */}
            <div className="text-center mb-6">
              <h2 className="text-2xl font-extrabold text-white tracking-tight leading-snug truncate">
                {currentTrack?.title || 'No Track Selected'}
              </h2>
              <p className="text-sm font-medium text-zinc-400 mt-1 truncate">
                {currentTrack?.rawArtist || currentTrack?.albumArtist || 'Select music to begin'}
              </p>
            </div>

            {/* Scrubber: Radix UI Slider with Timestamp Tooltip */}
            <div className="mb-6">
              <Tooltip.Provider delayDuration={100}>
                <Tooltip.Root>
                  <Tooltip.Trigger asChild>
                    <div className="w-full">
                      <Slider.Root
                        className="relative flex items-center select-none touch-none w-full h-5 cursor-pointer group"
                        value={[currentTime]}
                        max={duration > 0 ? duration : 100}
                        step={0.5}
                        onValueChange={([val]) => seek(val)}
                      >
                        <Slider.Track className="bg-white/[0.1] relative grow rounded-full h-1.5 overflow-hidden">
                          <Slider.Range className="absolute bg-white group-hover:bg-rose-500 rounded-full h-full transition-colors" />
                        </Slider.Track>
                        <Slider.Thumb className="block w-3.5 h-3.5 bg-white rounded-full shadow-lg border border-zinc-300 focus:outline-none opacity-0 group-hover:opacity-100 transition-opacity" />
                      </Slider.Root>
                    </div>
                  </Tooltip.Trigger>
                  <Tooltip.Portal>
                    <Tooltip.Content
                      side="top"
                      className="liquid-glass text-white text-[11px] font-hud px-2.5 py-1 rounded-md shadow-xl"
                    >
                      {formatTime(currentTime)} / {formatTime(duration)}
                      <Tooltip.Arrow className="fill-black" />
                    </Tooltip.Content>
                  </Tooltip.Portal>
                </Tooltip.Root>
              </Tooltip.Provider>

              {/* Time Indicators matching Image 2 */}
              <div className="flex items-center justify-between text-xs font-hud text-zinc-500 mt-1.5">
                <span>{formatTime(currentTime)}</span>
                <span>{formatTime(duration)}</span>
              </div>
            </div>

            {/* Controls Row matching Image 2: Shuffle, Prev, Big Play, Next, Repeat */}
            <div className="flex items-center justify-center gap-6 mb-6">
              <button
                onClick={toggleShuffle}
                title="Shuffle"
                className={`p-2 rounded-full transition-colors cursor-pointer ${
                  isShuffle ? 'text-white' : 'text-zinc-500 hover:text-white'
                }`}
              >
                <Shuffle className="w-4 h-4" />
              </button>

              <button
                onClick={previous}
                title="Previous"
                className="p-2 text-zinc-400 hover:text-white transition-colors cursor-pointer"
              >
                <SkipBack className="w-5 h-5 fill-current" />
              </button>

              {/* Big Circular Dark Play/Pause button with Liquid Glass border */}
              <motion.button
                whileHover={{ scale: 1.06 }}
                whileTap={{ scale: 0.94 }}
                onClick={togglePlay}
                title={isPlaying ? 'Pause' : 'Play'}
                className="w-14 h-14 rounded-full bg-white text-black flex items-center justify-center shadow-[0_10px_30px_rgba(255,255,255,0.15)] hover:bg-zinc-200 transition-all cursor-pointer"
              >
                {isPlaying ? (
                  <Pause className="w-6 h-6 fill-current" />
                ) : (
                  <Play className="w-6 h-6 fill-current ml-0.5" />
                )}
              </motion.button>

              <button
                onClick={next}
                title="Next"
                className="p-2 text-zinc-400 hover:text-white transition-colors cursor-pointer"
              >
                <SkipForward className="w-5 h-5 fill-current" />
              </button>

              <button
                onClick={toggleRepeat}
                title={`Repeat: ${repeatMode}`}
                className={`p-2 rounded-full transition-colors cursor-pointer ${
                  repeatMode !== 'off' ? 'text-white' : 'text-zinc-500 hover:text-white'
                }`}
              >
                {repeatMode === 'one' ? <Repeat1 className="w-4 h-4" /> : <Repeat className="w-4 h-4" />}
              </button>
            </div>

            {/* Bottom Actions Row: Heart, Queue, DSP, Volume, Options */}
            <div className="flex items-center justify-between pt-4 border-t border-white/[0.06]">
              <button
                onClick={toggleStarCurrentTrack}
                title={currentTrack?.isStarred ? 'Unstar' : 'Star'}
                className={`p-2 rounded-full transition-colors cursor-pointer ${
                  currentTrack?.isStarred ? 'text-rose-500' : 'text-zinc-500 hover:text-white'
                }`}
              >
                <Heart className={`w-4 h-4 ${currentTrack?.isStarred ? 'fill-current' : ''}`} />
              </button>

              <button
                onClick={() => setMode('queue')}
                title="View Queue"
                className="p-2 text-zinc-500 hover:text-white transition-colors cursor-pointer"
              >
                <ListMusic className="w-4 h-4" />
              </button>

              <button
                onClick={() => setShowDSP(true)}
                title="DSP Studio Audio Engine"
                className={`p-2 rounded-full transition-colors cursor-pointer ${
                  dspState.enabled ? 'text-rose-500' : 'text-zinc-500 hover:text-white'
                }`}
              >
                <Sliders className="w-4 h-4" />
              </button>

              {/* Volume Slider with Mute Button */}
              <div className="flex items-center gap-2">
                <button
                  onClick={toggleMute}
                  className="text-zinc-500 hover:text-white p-1 cursor-pointer"
                >
                  {isMuted || volume === 0 ? (
                    <VolumeX className="w-4 h-4" />
                  ) : (
                    <Volume2 className="w-4 h-4" />
                  )}
                </button>
                <Slider.Root
                  className="relative flex items-center select-none touch-none w-16 h-4 cursor-pointer"
                  value={[isMuted ? 0 : volume]}
                  max={1}
                  step={0.02}
                  onValueChange={([val]) => setVolume(val)}
                >
                  <Slider.Track className="bg-white/[0.1] relative grow rounded-full h-1">
                    <Slider.Range className="absolute bg-white rounded-full h-full" />
                  </Slider.Track>
                  <Slider.Thumb className="block w-2.5 h-2.5 bg-white rounded-full shadow focus:outline-none" />
                </Slider.Root>
              </div>

              {/* More Actions Dropdown */}
              <DropdownMenu.Root>
                <DropdownMenu.Trigger asChild>
                  <button className="p-2 text-zinc-500 hover:text-white rounded-full cursor-pointer">
                    <MoreHorizontal className="w-4 h-4" />
                  </button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Portal>
                  <DropdownMenu.Content className="min-w-[180px] liquid-glass rounded-xl p-1.5 shadow-2xl z-50 text-xs text-zinc-300 font-sans">
                    {currentTrack && (
                      <DropdownMenu.Item
                        onClick={() => window.open(api.getStreamUrl(currentTrack.id), '_blank')}
                        className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none"
                      >
                        <Download className="w-3.5 h-3.5" /> Download Current Song
                      </DropdownMenu.Item>
                    )}
                    <DropdownMenu.Item
                      onClick={() => setShowDSP(true)}
                      className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none"
                    >
                      <Sliders className="w-3.5 h-3.5 text-rose-500" /> Open DSP Master Engine
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Portal>
              </DropdownMenu.Root>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <AudioDSPModal isOpen={showDSP} onClose={() => setShowDSP(false)} />
    </aside>
  )
}
