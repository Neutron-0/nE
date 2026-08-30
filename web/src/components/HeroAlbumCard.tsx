import React, { useState } from 'react'
import { motion } from 'motion/react'
import { Play, Shuffle, Disc } from 'lucide-react'
import type { Album, Track } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

interface HeroAlbumCardProps {
  album: Album | null
  tracks: Track[]
  onSelectAlbum?: (album: Album) => void
}

export const HeroAlbumCard: React.FC<HeroAlbumCardProps> = ({ album, tracks, onSelectAlbum }) => {
  const { playTrack, isPlaying, currentTrack } = usePlayerStore()
  const [isHovered, setIsHovered] = useState(false)

  const title = album?.title || 'Digital Ethereal'
  const artist = album?.albumArtist || 'Lumina Collective'
  const year = album?.year || 2024
  const songCount = tracks.length || album?.trackCount || 12
  const totalDurationMin = album?.duration ? Math.round(album.duration / 60) : 48
  const artworkUrl = album?.id ? api.getArtworkUrl('album', album.id, 500) : ''

  const isCurrentAlbum = currentTrack?.albumId === album?.id

  const handlePlayAlbum = () => {
    if (tracks.length > 0) {
      playTrack(tracks[0], tracks)
    }
  }

  const handleShuffleAlbum = () => {
    if (tracks.length > 0) {
      const randomIdx = Math.floor(Math.random() * tracks.length)
      playTrack(tracks[randomIdx], tracks)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="relative w-full rounded-[28px] bg-[#23201d] border border-[#38332e] p-7 md:p-9 shadow-2xl overflow-hidden mb-8 group"
    >
      {/* Ambient Backlight Glow */}
      <div className="absolute -top-12 -left-12 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl pointer-events-none -z-0" />
      <div className="absolute -bottom-12 -right-12 w-96 h-96 bg-rose-600/10 rounded-full blur-3xl pointer-events-none -z-0" />

      <div className="flex flex-col md:flex-row items-center gap-8 md:gap-10 relative z-10">
        {/* Left: Interactive Vinyl Record & Sleeve Jacket (Matching Image 2) */}
        <div
          className="relative w-48 h-48 md:w-56 md:h-56 shrink-0 cursor-pointer"
          onClick={() => album && onSelectAlbum?.(album)}
        >
          {/* Vinyl Record sliding out to the right */}
          <motion.div
            animate={{
              x: isHovered || (isPlaying && isCurrentAlbum) ? 68 : 16,
              rotate: isPlaying && isCurrentAlbum ? 360 : 0,
            }}
            transition={{
              x: { type: 'spring', stiffness: 240, damping: 22 },
              rotate: { repeat: Infinity, duration: 16, ease: 'linear' },
            }}
            className="absolute top-2 right-0 w-44 h-44 md:w-52 md:h-52 rounded-full vinyl-grooves border-2 border-[#2b2724] flex items-center justify-center shadow-2xl z-0"
          >
            {/* Realistic Physical Specular Light Sheen */}
            <div
              className="absolute inset-0 rounded-full pointer-events-none opacity-40"
              style={{
                background:
                  'conic-gradient(from 45deg, transparent 0deg, rgba(255,255,255,0.18) 45deg, transparent 90deg, rgba(255,255,255,0.12) 225deg, transparent 270deg)',
              }}
            />

            {/* Center Vinyl Label with miniature typography */}
            <div className="w-18 h-18 md:w-22 md:h-22 rounded-full bg-gradient-to-tr from-amber-700 via-rose-700 to-amber-600 border-2 border-black flex flex-col items-center justify-center overflow-hidden shadow-inner p-1">
              <span className="text-[7px] text-white/80 font-bold uppercase tracking-wider truncate max-w-[50px]">
                {artist}
              </span>
              <div className="w-4 h-4 rounded-full bg-black border border-zinc-700 my-0.5" />
              <span className="text-[6px] text-amber-200/90 font-mono tracking-tighter truncate max-w-[50px]">
                33 RPM
              </span>
            </div>
          </motion.div>

          {/* Sleeve Jacket (Front Cover Card) */}
          <div className="relative z-10 w-44 h-44 md:w-52 md:h-52 rounded-[22px] bg-[#1a1816] border border-white/10 shadow-2xl overflow-hidden flex items-center justify-center">
            {artworkUrl ? (
              <img
                src={artworkUrl}
                alt={title}
                className="w-full h-full object-cover"
                onError={(e) => {
                  ;(e.target as HTMLElement).style.display = 'none'
                }}
              />
            ) : (
              <div className="w-full h-full bg-gradient-to-tr from-zinc-800 to-zinc-900 flex items-center justify-center">
                <Disc className="w-20 h-20 text-zinc-700" />
              </div>
            )}

            {/* Vintage Spine Shadow & Edge Sheen */}
            <div className="absolute inset-y-0 left-0 w-3.5 bg-gradient-to-r from-black/70 to-transparent pointer-events-none" />
            <div className="absolute inset-x-0 top-0 h-1/2 bg-gradient-to-b from-white/10 to-transparent pointer-events-none" />
          </div>
        </div>

        {/* Right: Album Details matching Image 2 */}
        <div className="flex-1 min-w-0 text-center md:text-left">
          {/* Metadata tag */}
          <div className="flex items-center justify-center md:justify-start gap-2 text-xs font-semibold text-zinc-400 mb-2">
            <span>Album</span>
            <span>•</span>
            <span>{year}</span>
          </div>

          {/* Huge Bold Title */}
          <h1 className="text-3xl sm:text-4xl md:text-5xl font-black tracking-tight text-white mb-2 leading-tight">
            {title}
          </h1>

          {/* Artist Subtitle & Song Stats */}
          <div className="flex flex-wrap items-center justify-center md:justify-start gap-2 text-sm text-zinc-300 mb-6">
            <span className="font-semibold text-zinc-100">{artist}</span>
            <span className="text-zinc-600">•</span>
            <span className="text-zinc-400">
              {songCount} Songs, {totalDurationMin} min
            </span>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-center md:justify-start gap-3.5">
            <motion.button
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
              onClick={handlePlayAlbum}
              className="inline-flex items-center gap-2.5 px-6 py-3 rounded-full bg-white text-black font-bold text-sm shadow-xl hover:bg-zinc-100 transition-colors cursor-pointer"
            >
              <Play className="w-4 h-4 fill-current ml-0.5" />
              Play
            </motion.button>

            <motion.button
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
              onClick={handleShuffleAlbum}
              className="inline-flex items-center gap-2 px-5 py-3 rounded-full bg-[#2e2a26] hover:bg-[#3b3530] border border-white/5 text-zinc-200 font-semibold text-sm transition-colors cursor-pointer"
            >
              <Shuffle className="w-4 h-4" />
              Shuffle
            </motion.button>
          </div>
        </div>
      </div>
    </motion.div>
  )
}
