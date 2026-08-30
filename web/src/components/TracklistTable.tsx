import React from 'react'
import { motion } from 'motion/react'
import { Clock, Play, Heart, MoreHorizontal, Download } from 'lucide-react'
import * as DropdownMenu from '@radix-ui/react-dropdown-menu'
import type { Track } from '../types'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'

interface TracklistTableProps {
  tracks: Track[]
  onStarToggle?: (track: Track) => void
}

export const TracklistTable: React.FC<TracklistTableProps> = ({ tracks, onStarToggle }) => {
  const { playTrack, currentTrack, isPlaying } = usePlayerStore()

  const formatTime = (seconds: number) => {
    if (!seconds || isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`
  }

  if (tracks.length === 0) {
    return (
      <div className="py-16 text-center text-zinc-500 font-medium">
        No tracks found in this collection.
      </div>
    )
  }

  return (
    <div className="w-full select-none">
      {/* Header Row matching Image 2 */}
      <div className="grid grid-cols-12 text-[12px] font-semibold text-zinc-400 pb-3 border-b border-white/5 px-4 mb-2">
        <span className="col-span-1 text-center">#</span>
        <span className="col-span-6">Title</span>
        <span className="col-span-4">Artist</span>
        <span className="col-span-1 text-right flex items-center justify-end">
          <Clock className="w-3.5 h-3.5 text-zinc-400" />
        </span>
      </div>

      {/* Track Rows matching Image 2 */}
      <div className="space-y-1">
        {tracks.map((track, idx) => {
          const isCurrent = currentTrack?.id === track.id
          const trackNum = track.trackNumber || idx + 1

          return (
            <motion.div
              key={track.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: Math.min(idx * 0.02, 0.3) }}
              onClick={() => playTrack(track, tracks)}
              className={`grid grid-cols-12 items-center py-3.5 px-4 rounded-xl cursor-pointer group transition-colors ${
                isCurrent
                  ? 'bg-[#2b2724] text-white shadow-sm'
                  : 'hover:bg-[#23201e]/80 text-zinc-300'
              }`}
            >
              {/* Col 1: # or Play/Equalizer */}
              <div className="col-span-1 text-center text-xs font-hud text-zinc-400 group-hover:text-white flex items-center justify-center">
                {isCurrent && isPlaying ? (
                  <div className="flex items-end gap-0.5 h-3.5 w-3.5">
                    <span className="w-0.5 h-full bg-rose-500 rounded-full animate-pulse" />
                    <span className="w-0.5 h-2 bg-rose-400 rounded-full animate-bounce" />
                    <span className="w-0.5 h-3 bg-rose-500 rounded-full animate-pulse" />
                  </div>
                ) : (
                  <span className="group-hover:hidden">{trackNum}</span>
                )}
                {!isCurrent && (
                  <Play className="w-3.5 h-3.5 fill-current hidden group-hover:block ml-0.5" />
                )}
              </div>

              {/* Col 2: Title */}
              <div className="col-span-6 min-w-0 pr-4">
                <p
                  className={`text-sm font-semibold truncate ${
                    isCurrent ? 'text-white font-bold' : 'text-zinc-200'
                  }`}
                >
                  {track.title}
                </p>
              </div>

              {/* Col 3: Artist */}
              <div className="col-span-4 min-w-0 pr-4">
                <p className="text-xs text-zinc-400 truncate">
                  {track.rawArtist || track.albumArtist || 'Unknown Artist'}
                </p>
              </div>

              {/* Col 4: Actions & Duration */}
              <div className="col-span-1 flex items-center justify-end gap-2 text-xs font-hud text-zinc-400">
                {/* Star Button */}
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    onStarToggle?.(track)
                  }}
                  title={track.isStarred ? 'Unstar' : 'Star'}
                  className={`p-1 rounded transition-colors ${
                    track.isStarred
                      ? 'text-rose-500'
                      : 'opacity-0 group-hover:opacity-100 hover:text-zinc-200'
                  }`}
                >
                  <Heart className={`w-3.5 h-3.5 ${track.isStarred ? 'fill-current' : ''}`} />
                </button>

                {/* Radix Dropdown Context Menu */}
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger asChild>
                    <button
                      onClick={(e) => e.stopPropagation()}
                      className="p-1 text-zinc-400 hover:text-white rounded opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <MoreHorizontal className="w-3.5 h-3.5" />
                    </button>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Portal>
                    <DropdownMenu.Content
                      onClick={(e) => e.stopPropagation()}
                      className="min-w-[160px] bg-[#1f1d1a] border border-white/10 rounded-xl p-1.5 shadow-2xl z-50 text-xs text-zinc-300 font-sans"
                    >
                      <DropdownMenu.Item
                        onClick={() => playTrack(track, tracks)}
                        className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg hover:bg-white/10 cursor-pointer outline-none"
                      >
                        <Play className="w-3.5 h-3.5" /> Play Now
                      </DropdownMenu.Item>
                      <DropdownMenu.Item
                        onClick={() => window.open(api.getStreamUrl(track.id), '_blank')}
                        className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg hover:bg-white/10 cursor-pointer outline-none"
                      >
                        <Download className="w-3.5 h-3.5" /> Download Audio
                      </DropdownMenu.Item>
                    </DropdownMenu.Content>
                  </DropdownMenu.Portal>
                </DropdownMenu.Root>

                {/* Duration */}
                <span className="w-10 text-right">{formatTime(track.duration)}</span>
              </div>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
