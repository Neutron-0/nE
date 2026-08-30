import React, { useEffect, useState } from 'react'
import { User, Disc } from 'lucide-react'
import type { Artist } from '../types'
import { api } from '../lib/api'

interface ArtistsViewProps {
  searchQuery: string
}

export const ArtistsView: React.FC<ArtistsViewProps> = ({ searchQuery }) => {
  const [artists, setArtists] = useState<Artist[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadArtists()
  }, [])

  const loadArtists = async () => {
    try {
      setIsLoading(true)
      const res = await api.getArtists(100, 0)
      setArtists(res.items)
    } catch (err) {
      console.error('Error fetching artists:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredArtists = artists.filter((a) =>
    a.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-medium animate-pulse">
        Loading artists...
      </div>
    )
  }

  if (filteredArtists.length === 0) {
    return (
      <div className="p-12 text-center text-zinc-500">
        <User className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
        <h3 className="text-lg font-semibold text-zinc-400">No artists found</h3>
        <p className="text-sm text-zinc-600 mt-1">Add music to your library to populate artists</p>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-white tracking-tight">Artists</h2>
        <span className="text-xs text-zinc-500 font-mono">{filteredArtists.length} artists</span>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6">
        {filteredArtists.map((artist) => (
          <div
            key={artist.id}
            className="group bg-zinc-900/40 hover:bg-zinc-900 border border-zinc-800/40 hover:border-zinc-700/60 rounded-xl p-4 transition-all duration-200 cursor-pointer flex flex-col items-center text-center shadow-sm hover:shadow-xl hover:-translate-y-0.5"
          >
            <div className="w-32 h-32 rounded-full bg-gradient-to-tr from-zinc-800 to-zinc-950 border border-zinc-700/50 flex items-center justify-center mb-4 shadow-md group-hover:scale-105 transition-transform duration-300">
              <User className="w-12 h-12 text-zinc-600 stroke-1" />
            </div>

            <h4 className="font-semibold text-sm text-zinc-100 truncate w-full group-hover:text-rose-400 transition-colors">
              {artist.name}
            </h4>

            <div className="flex items-center gap-2 mt-1.5 text-xs text-zinc-500 font-mono">
              <span className="flex items-center gap-1">
                <Disc className="w-3 h-3" /> {artist.albumCount} albums
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
