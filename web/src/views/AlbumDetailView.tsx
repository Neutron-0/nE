import React, { useEffect, useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import type { Album, Track } from '../types'
import { api } from '../lib/api'
import { HeroAlbumCard } from '../components/HeroAlbumCard'
import { TracklistTable } from '../components/TracklistTable'

interface AlbumDetailViewProps {
  album: Album
  onBack: () => void
}

export const AlbumDetailView: React.FC<AlbumDetailViewProps> = ({ album, onBack }) => {
  const [tracks, setTracks] = useState<Track[]>(album.tracks || [])
  const [isLoading, setIsLoading] = useState(!album.tracks)

  useEffect(() => {
    if (!album.tracks) {
      loadFullAlbum()
    }
  }, [album.id])

  const loadFullAlbum = async () => {
    try {
      setIsLoading(true)
      const res = await api.getAlbum(album.id)
      setTracks(res.tracks || [])
    } catch (err) {
      console.error('Error fetching album details:', err)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="p-6 md:p-8 space-y-6">
      {/* Back Button */}
      <button
        onClick={onBack}
        className="inline-flex items-center gap-2 text-xs font-semibold text-zinc-400 hover:text-white transition-colors cursor-pointer px-2 py-1 rounded-lg hover:bg-white/5"
      >
        <ArrowLeft className="w-4 h-4" /> Back to Albums
      </button>

      {/* Hero Showcase Card matching Image 2 */}
      <HeroAlbumCard album={album} tracks={tracks} />

      {/* Tracklist matching Image 2 */}
      {isLoading ? (
        <div className="p-8 text-center text-zinc-500 font-hud text-sm animate-pulse">
          LOADING TRACKLIST TELEMETRY...
        </div>
      ) : (
        <TracklistTable tracks={tracks} />
      )}
    </div>
  )
}
