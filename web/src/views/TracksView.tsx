import React, { useEffect, useState } from 'react'
import { Music } from 'lucide-react'
import type { Track } from '../types'
import { api } from '../lib/api'
import { TracklistTable } from '../components/TracklistTable'

interface TracksViewProps {
  searchQuery: string
}

export const TracksView: React.FC<TracksViewProps> = ({ searchQuery }) => {
  const [tracks, setTracks] = useState<Track[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadTracks()
  }, [])

  const loadTracks = async () => {
    try {
      setIsLoading(true)
      const res = await api.getTracks(200, 0)
      setTracks(res.items || [])
    } catch (err) {
      console.error('Error fetching tracks:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredTracks = tracks.filter((t) => {
    const q = searchQuery.toLowerCase()
    return (
      t.title.toLowerCase().includes(q) ||
      (t.rawArtist && t.rawArtist.toLowerCase().includes(q)) ||
      (t.albumTitle && t.albumTitle.toLowerCase().includes(q))
    )
  })

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-hud text-sm">
        <p className="animate-pulse">LOADING TRACKS...</p>
      </div>
    )
  }

  if (filteredTracks.length === 0) {
    return (
      <div className="p-12 text-center text-zinc-500">
        <Music className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
        <h3 className="text-lg font-semibold text-zinc-400">No tracks found</h3>
        <p className="text-sm text-zinc-600 mt-1">Add audio files to your library and scan</p>
      </div>
    )
  }

  return (
    <div className="p-6 md:p-8 space-y-6">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-2xl font-extrabold text-white tracking-tight">Explore Songs</h2>
        <span className="text-xs text-zinc-400 font-hud">{filteredTracks.length} tracks</span>
      </div>

      <TracklistTable tracks={filteredTracks} />
    </div>
  )
}
