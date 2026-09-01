import React, { useState } from 'react'
import { motion } from 'motion/react'
import { Clock, Play, Heart, MoreHorizontal, Download, ListPlus, Check } from 'lucide-react'
import * as DropdownMenu from '@radix-ui/react-dropdown-menu'
import type { Track, Playlist } from '../types'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'

interface TracklistTableProps {
  tracks: Track[]
  onStarToggle?: (track: Track) => void
}

export const TracklistTable: React.FC<TracklistTableProps> = ({ tracks, onStarToggle }) => {
  const { playTrack, currentTrack, isPlaying } = usePlayerStore()
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [addedFeedback, setAddedFeedback] = useState<{ [key: string]: boolean }>({})

  const loadPlaylists = async () => {
    try {
      const res = await api.getPlaylists()
      setPlaylists(res.items || [])
    } catch {}
  }

  const handleAddToPlaylist = async (playlistId: string, trackId: string) => {
    try {
      await api.addTrackToPlaylist(playlistId, trackId)
      const key = `${playlistId}-${trackId}`
      setAddedFeedback((prev) => ({ ...prev, [key]: true }))
      setTimeout(() => {
        setAddedFeedback((prev) => ({ ...prev, [key]: false }))
      }, 2000)
    } catch (err) {
      console.warn('Failed to add track to playlist:', err)
    }
  }

  const formatTime = (seconds: number) => {
    if (!seconds || isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`
  }

  if (tracks.length === 0) {
    return (
      <div className="py-16 text-center text-zinc-600 font-hud text-xs">
        NO AUDIO ENTITIES LOCATED IN SPECIFIED DATASET.
      </div>
    )
  }

  return (
    <div className="w-full select-none">
      {/* Header Row matching Image 2 */}
      <div className="grid grid-cols-12 text-[11px] font-bold text-zinc-500 uppercase tracking-widest pb-3 border-b border-white/[0.06] px-4 mb-2 font-hud">
        <span className="col-span-1 text-center">#</span>
        <span className="col-span-6">Title</span>
        <span className="col-span-4">Artist</span>
        <span className="col-span-1 text-right flex items-center justify-end">
          <Clock className="w-3.5 h-3.5 text-zinc-500" />
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
              className={`grid grid-cols-12 items-center py-3 px-4 rounded-xl cursor-pointer group transition-all ${
                isCurrent
                  ? 'liquid-glass-pill text-white shadow-lg'
                  : 'hover:bg-white/[0.035] text-zinc-400'
              }`}
            >
              {/* Col 1: # or Equalizer */}
              <div className="col-span-1 text-center text-xs font-hud text-zinc-500 group-hover:text-white flex items-center justify-center">
                {isCurrent && isPlaying ? (
                  <div className="flex items-end gap-0.5 h-3.5 w-3.5">
                    <span className="w-0.5 h-full bg-rose-500 rounded-full animate-pulse" />
                    <span className="w-0.5 h-2 bg-white rounded-full animate-bounce" />
                    <span className="w-0.5 h-3 bg-rose-500 rounded-full animate-pulse" />
                  </div>
                ) : (
                  <span className="group-hover:hidden">{trackNum}</span>
                )}
                {!isCurrent && (
                  <Play className="w-3.5 h-3.5 fill-current hidden group-hover:block ml-0.5 text-white" />
                )}
              </div>

              {/* Col 2: Title */}
              <div className="col-span-6 min-w-0 pr-4">
                <p
                  className={`text-sm font-semibold truncate ${
                    isCurrent ? 'text-white font-bold' : 'text-zinc-200 group-hover:text-white'
                  }`}
                >
                  {track.title}
                </p>
              </div>

              {/* Col 3: Artist */}
              <div className="col-span-4 min-w-0 pr-4">
                <p className="text-xs text-zinc-400 truncate group-hover:text-zinc-300">
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
                      : 'opacity-0 group-hover:opacity-100 hover:text-white'
                  }`}
                >
                  <Heart className={`w-3.5 h-3.5 ${track.isStarred ? 'fill-current' : ''}`} />
                </button>

                {/* Radix Dropdown Context Menu */}
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger asChild>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        loadPlaylists()
                      }}
                      className="p-1 text-zinc-500 hover:text-white rounded opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <MoreHorizontal className="w-3.5 h-3.5" />
                    </button>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Portal>
                    <DropdownMenu.Content
                      onClick={(e) => e.stopPropagation()}
                      className="min-w-[160px] liquid-glass rounded-xl p-1.5 shadow-2xl z-50 text-xs text-zinc-300 font-sans"
                    >
                      <DropdownMenu.Item
                        onClick={() => playTrack(track, tracks)}
                        className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none"
                      >
                        <Play className="w-3.5 h-3.5" /> Play Now
                      </DropdownMenu.Item>
                      <DropdownMenu.Sub>
                        <DropdownMenu.SubTrigger className="flex items-center justify-between px-2.5 py-1.5 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none w-full">
                          <span className="flex items-center gap-2">
                            <ListPlus className="w-3.5 h-3.5" /> Add to Playlist
                          </span>
                          <span className="text-[10px] text-zinc-500">▶</span>
                        </DropdownMenu.SubTrigger>
                        <DropdownMenu.Portal>
                          <DropdownMenu.SubContent className="min-w-[170px] liquid-glass rounded-xl p-1.5 shadow-2xl z-50 text-xs text-zinc-300 font-sans">
                            {playlists.length === 0 ? (
                              <div className="px-2.5 py-1.5 text-zinc-500 text-[11px]">No playlists found</div>
                            ) : (
                              playlists.map((pl) => {
                                const key = `${pl.id}-${track.id}`
                                const isAdded = addedFeedback[key]
                                return (
                                  <DropdownMenu.Item
                                    key={pl.id}
                                    onClick={() => handleAddToPlaylist(pl.id, track.id)}
                                    className="flex items-center justify-between px-2.5 py-1.5 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none"
                                  >
                                    <span className="truncate pr-2">{pl.name}</span>
                                    {isAdded && <Check className="w-3 h-3 text-emerald-400 shrink-0" />}
                                  </DropdownMenu.Item>
                                )
                              })
                            )}
                          </DropdownMenu.SubContent>
                        </DropdownMenu.Portal>
                      </DropdownMenu.Sub>
                      <DropdownMenu.Item
                        onClick={() => window.open(api.getStreamUrl(track.id), '_blank')}
                        className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg hover:bg-white/[0.08] cursor-pointer outline-none"
                      >
                        <Download className="w-3.5 h-3.5" /> Download Audio
                      </DropdownMenu.Item>
                    </DropdownMenu.Content>
                  </DropdownMenu.Portal>
                </DropdownMenu.Root>

                {/* Duration */}
                <span className="w-10 text-right text-zinc-500 group-hover:text-zinc-300">
                  {formatTime(track.duration)}
                </span>
              </div>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
