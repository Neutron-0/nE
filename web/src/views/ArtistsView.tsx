import React, { useEffect, useState } from 'react'
import { User, Disc, X, Globe } from 'lucide-react'
import type { Artist, ArtistBiography, Album } from '../types'
import { api } from '../lib/api'

interface ArtistsViewProps {
  searchQuery: string
}

export const ArtistsView: React.FC<ArtistsViewProps> = ({ searchQuery }) => {
  const [artists, setArtists] = useState<Artist[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [selectedArtist, setSelectedArtist] = useState<Artist | null>(null)
  const [bio, setBio] = useState<ArtistBiography | null>(null)
  const [artistAlbums, setArtistAlbums] = useState<Album[]>([])
  const [loadingBio, setLoadingBio] = useState(false)

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

  const handleSelectArtist = async (artist: Artist) => {
    setSelectedArtist(artist)
    setLoadingBio(true)
    try {
      const [bioRes, albumsRes] = await Promise.all([
        api.getArtistBiography(artist.id),
        api.getAlbums(100, 0),
      ])
      setBio(bioRes)
      setArtistAlbums(albumsRes.items.filter((alb) => alb.albumArtistId === artist.id))
    } catch (err) {
      console.warn('Failed loading artist bio:', err)
    } finally {
      setLoadingBio(false)
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
            onClick={() => handleSelectArtist(artist)}
            className="group bg-zinc-900/40 hover:bg-zinc-900 border border-zinc-800/40 hover:border-zinc-700/60 rounded-xl p-4 transition-all duration-200 cursor-pointer flex flex-col items-center text-center shadow-sm hover:shadow-xl hover:-translate-y-0.5"
          >
            <div className="w-32 h-32 rounded-full bg-gradient-to-tr from-zinc-800 to-zinc-950 border border-zinc-700/50 flex items-center justify-center mb-4 shadow-md group-hover:scale-105 transition-transform duration-300 overflow-hidden">
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

      {/* Artist Detail Modal */}
      {selectedArtist && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50 backdrop-blur-md">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-6 max-w-2xl w-full shadow-2xl overflow-y-auto max-h-[85vh]">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-4">
                {bio?.imageUrl ? (
                  <img
                    src={bio.imageUrl}
                    alt={selectedArtist.name}
                    className="w-16 h-16 rounded-full object-cover border border-white/20 shadow-md"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-full bg-zinc-800 flex items-center justify-center text-zinc-400 border border-zinc-700">
                    <User className="w-8 h-8" />
                  </div>
                )}
                <div>
                  <h3 className="text-2xl font-bold text-white leading-tight flex items-center gap-2">
                    {selectedArtist.name}
                  </h3>
                  <p className="text-xs text-zinc-400 mt-1 flex items-center gap-1.5">
                    <Disc className="w-3 h-3 text-rose-400" /> {selectedArtist.albumCount} albums in library
                  </p>
                </div>
              </div>

              <button
                onClick={() => setSelectedArtist(null)}
                className="w-8 h-8 rounded-full bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white flex items-center justify-center cursor-pointer transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Biography Section */}
            <div className="bg-zinc-950/60 border border-zinc-800/80 rounded-xl p-4 mb-6">
              <div className="flex items-center gap-2 mb-2">
                <Globe className="w-4 h-4 text-amber-400" />
                <h4 className="text-xs font-bold text-zinc-300 uppercase tracking-wider">Biography & Editorial</h4>
              </div>
              {loadingBio ? (
                <div className="flex items-center gap-2 text-xs text-zinc-500 py-3">
                  <div className="w-3.5 h-3.5 border-2 border-rose-500 border-t-transparent rounded-full animate-spin" />
                  Fetching artist story...
                </div>
              ) : (
                <p className="text-sm text-zinc-300 leading-relaxed">
                  {bio?.biography || `${selectedArtist.name} is an artist featured in your catalog.`}
                </p>
              )}
            </div>

            {/* Albums Section */}
            <div>
              <h4 className="text-xs font-bold text-zinc-400 uppercase tracking-wider mb-3">Albums in Library</h4>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {artistAlbums.map((alb) => (
                  <div key={alb.id} className="p-3 bg-zinc-950 border border-zinc-800/60 rounded-lg">
                    <h5 className="font-semibold text-xs text-white truncate">{alb.title}</h5>
                    <p className="text-[11px] text-zinc-500 mt-0.5">{alb.year || 'Unknown Year'} • {alb.trackCount} tracks</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
