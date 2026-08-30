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
      setArtists(res.items || [])
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
      setArtistAlbums((albumsRes.items || []).filter((alb) => alb.albumArtistId === artist.id))
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
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-hud text-sm">
        <p className="animate-pulse">LOADING ARTIST ROSTER...</p>
      </div>
    )
  }

  if (filteredArtists.length === 0) {
    return (
      <div className="p-12 text-center text-zinc-500">
        <User className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
        <h3 className="text-lg font-bold text-zinc-400">No artists found</h3>
        <p className="text-xs text-zinc-600 mt-1">Add music to your library to populate artists</p>
      </div>
    )
  }

  return (
    <div className="p-6 md:p-8 space-y-6">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-2xl font-extrabold text-white tracking-tight">Artists Library</h2>
        <span className="text-xs text-zinc-400 font-hud">{filteredArtists.length} artists</span>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-5">
        {filteredArtists.map((artist) => (
          <div
            key={artist.id}
            onClick={() => handleSelectArtist(artist)}
            className="group bg-[#211e1c] hover:bg-[#2b2724] border border-white/5 hover:border-white/15 rounded-2xl p-4 transition-all cursor-pointer flex flex-col items-center text-center shadow-lg hover:shadow-2xl hover:-translate-y-1"
          >
            <div className="w-28 h-28 rounded-full bg-[#181615] border border-white/10 flex items-center justify-center mb-3.5 shadow-inner group-hover:scale-105 transition-transform overflow-hidden relative">
              <User className="w-12 h-12 text-zinc-600 stroke-1" />
              {/* Subtle vinyl circular gloss */}
              <div className="absolute inset-0 rounded-full border border-white/5 pointer-events-none" />
            </div>

            <h4 className="font-bold text-sm text-zinc-100 truncate w-full group-hover:text-white transition-colors">
              {artist.name}
            </h4>

            <div className="flex items-center gap-1.5 mt-1 text-xs text-zinc-400 font-hud">
              <Disc className="w-3 h-3 text-amber-500" />
              <span>{artist.albumCount} {artist.albumCount === 1 ? 'album' : 'albums'}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Artist Detail Modal */}
      {selectedArtist && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50 backdrop-blur-xl">
          <div className="bg-[#1f1c1a] border border-white/10 rounded-[28px] p-6 max-w-2xl w-full shadow-2xl overflow-y-auto max-h-[85vh]">
            <div className="flex items-start justify-between mb-5">
              <div className="flex items-center gap-4">
                {bio?.imageUrl ? (
                  <img
                    src={bio.imageUrl}
                    alt={selectedArtist.name}
                    className="w-16 h-16 rounded-full object-cover border border-white/20 shadow-md"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-full bg-[#2a2622] flex items-center justify-center text-zinc-400 border border-white/10">
                    <User className="w-8 h-8" />
                  </div>
                )}
                <div>
                  <h3 className="text-2xl font-black text-white leading-tight">
                    {selectedArtist.name}
                  </h3>
                  <p className="text-xs text-zinc-400 mt-1 flex items-center gap-1.5 font-hud">
                    <Disc className="w-3.5 h-3.5 text-rose-400" /> {selectedArtist.albumCount} albums in library
                  </p>
                </div>
              </div>

              <button
                onClick={() => setSelectedArtist(null)}
                className="w-8 h-8 rounded-full bg-white/5 hover:bg-white/10 text-zinc-400 hover:text-white flex items-center justify-center cursor-pointer transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Editorial Biography */}
            <div className="bg-[#171514] border border-white/5 rounded-2xl p-5 mb-6">
              <div className="flex items-center gap-2 mb-2.5">
                <Globe className="w-4 h-4 text-amber-400" />
                <h4 className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider font-hud">
                  BIOGRAPHY & EDITORIAL
                </h4>
              </div>
              {loadingBio ? (
                <div className="flex items-center gap-2 text-xs text-zinc-400 py-3 font-hud">
                  <div className="w-3.5 h-3.5 border-2 border-rose-500 border-t-transparent rounded-full animate-spin" />
                  FETCHING ARTIST TELEMETRY...
                </div>
              ) : (
                <p className="text-xs md:text-sm text-zinc-300 leading-relaxed font-sans">
                  {bio?.biography || `${selectedArtist.name} is an artist featured in your catalog.`}
                </p>
              )}
            </div>

            {/* Albums in Library */}
            <div>
              <h4 className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-3 font-hud">
                ALBUMS IN COLLECTION
              </h4>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {artistAlbums.map((alb) => (
                  <div key={alb.id} className="p-3 bg-[#171514] border border-white/5 rounded-xl">
                    <h5 className="font-bold text-xs text-white truncate">{alb.title}</h5>
                    <p className="text-[11px] text-zinc-400 mt-0.5 font-hud">
                      {alb.year || 'Unknown'} • {alb.trackCount} tracks
                    </p>
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
