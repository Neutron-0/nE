import React, { useEffect, useState } from 'react'
import { Heart } from 'lucide-react'
import type { Track } from '../types'
import { api } from '../lib/api'
import { TracklistTable } from '../components/TracklistTable'

export const FavoritesView: React.FC = () => {
  const [tracks, setTracks] = useState<Track[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadFavorites()
  }, [])

  const loadFavorites = async () => {
    try {
      setIsLoading(true)
      const res = await api.getFavorites()
      setTracks(res.items || [])
    } catch (err) {
      console.error('Failed to load favorites:', err)
    } finally {
      setIsLoading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-hud text-sm">
        <p className="animate-pulse">LOADING FAVORITES...</p>
      </div>
    )
  }

  return (
    <div className="p-6 md:p-8 space-y-6">
      {/* Header with Liquid Glass */}
      <div className="flex items-center gap-5 pb-6 border-b border-white/[0.06]">
        <div className="w-16 h-16 rounded-[22px] liquid-glass flex items-center justify-center text-rose-500 shadow-xl">
          <Heart className="w-8 h-8 fill-current" />
        </div>
        <div>
          <span className="text-[11px] font-bold uppercase tracking-widest text-zinc-400 font-hud">
            COLLECTION
          </span>
          <h2 className="text-3xl font-black text-white tracking-tight mt-0.5">Starred Songs</h2>
          <p className="text-xs text-zinc-500 mt-1 font-hud">{tracks.length} favorites</p>
        </div>
      </div>

      {tracks.length === 0 ? (
        <div className="py-16 text-center text-zinc-500">
          <Heart className="w-12 h-12 mx-auto mb-3 text-zinc-800 stroke-1" />
          <h3 className="text-lg font-bold text-zinc-400">No starred tracks</h3>
          <p className="text-xs text-zinc-600 mt-1">Star tracks while listening to save them here</p>
        </div>
      ) : (
        <TracklistTable tracks={tracks} />
      )}
    </div>
  )
}
