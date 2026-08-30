import React, { useEffect, useState, useRef } from 'react'
import { Search, RefreshCw, Disc, Music, User, X, Loader2 } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'
import type { SearchResult, Album, Track } from '../types'

interface HeaderProps {
  searchQuery: string
  onSearchChange: (q: string) => void
  onSelectAlbum?: (album: Album) => void
}

export const Header: React.FC<HeaderProps> = ({ searchQuery, onSearchChange, onSelectAlbum }) => {
  const { user } = useAuthStore()
  const { playTrack } = usePlayerStore()
  const [isScanning, setIsScanning] = useState(false)
  const [searchResults, setSearchResults] = useState<SearchResult | null>(null)
  const [isSearching, setIsSearching] = useState(false)
  const [showResults, setShowResults] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!searchQuery.trim()) {
      setSearchResults(null)
      setShowResults(false)
      return
    }

    const timer = setTimeout(async () => {
      try {
        setIsSearching(true)
        const res = await api.search(searchQuery.trim(), 5)
        setSearchResults(res)
        setShowResults(true)
      } catch (err) {
        console.error('Search error:', err)
      } finally {
        setIsSearching(false)
      }
    }, 250)

    return () => clearTimeout(timer)
  }, [searchQuery])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setShowResults(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleTriggerScan = async () => {
    try {
      setIsScanning(true)
      const libs = await api.getLibraries()
      if (libs.items.length > 0) {
        await api.triggerScan(libs.items[0].id)
      }
      setTimeout(() => setIsScanning(false), 3000)
    } catch (err) {
      console.error('Scan trigger error:', err)
      setIsScanning(false)
    }
  }

  const handleTrackClick = (track: Track) => {
    playTrack(track)
    setShowResults(false)
  }

  const handleAlbumClick = (album: Album) => {
    if (onSelectAlbum) {
      onSelectAlbum(album)
    }
    setShowResults(false)
  }

  return (
    <header className="h-16 border-b border-zinc-800/60 bg-zinc-950/40 px-6 flex items-center justify-between gap-4 backdrop-blur-md relative z-40">
      {/* Live FTS5 Search Input */}
      <div ref={searchRef} className="relative max-w-md w-full">
        {isSearching ? (
          <Loader2 className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-rose-500 animate-spin" />
        ) : (
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
        )}
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          onFocus={() => {
            if (searchResults) setShowResults(true)
          }}
          placeholder="Search tracks, albums, artists (FTS5)..."
          className="w-full bg-zinc-900/80 border border-zinc-800 rounded-full pl-9 pr-8 py-1.5 text-sm text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-rose-500/80 focus:ring-1 focus:ring-rose-500/50 transition-all"
        />
        {searchQuery && (
          <button
            onClick={() => {
              onSearchChange('')
              setSearchResults(null)
              setShowResults(false)
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}

        {/* Live Search Results Dropdown */}
        {showResults && searchResults && (
          <div className="absolute left-0 right-0 top-full mt-2 bg-zinc-900/95 border border-zinc-800 rounded-2xl shadow-2xl overflow-hidden backdrop-blur-xl max-h-96 overflow-y-auto p-3 space-y-3 z-50">
            {searchResults.tracks.length > 0 && (
              <div>
                <p className="text-[11px] font-bold text-zinc-500 uppercase tracking-wider mb-1.5 px-2">Tracks</p>
                <div className="space-y-0.5">
                  {searchResults.tracks.map((t) => (
                    <div
                      key={t.id}
                      onClick={() => handleTrackClick(t)}
                      className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg hover:bg-zinc-800 cursor-pointer text-xs group"
                    >
                      <Music className="w-3.5 h-3.5 text-rose-400 shrink-0" />
                      <div className="min-w-0 flex-1">
                        <p className="font-medium text-white truncate">{t.title}</p>
                        <p className="text-zinc-400 truncate text-[11px]">{t.rawArtist}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {searchResults.albums.length > 0 && (
              <div>
                <p className="text-[11px] font-bold text-zinc-500 uppercase tracking-wider mb-1.5 px-2">Albums</p>
                <div className="space-y-0.5">
                  {searchResults.albums.map((a) => (
                    <div
                      key={a.id}
                      onClick={() => handleAlbumClick(a)}
                      className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg hover:bg-zinc-800 cursor-pointer text-xs group"
                    >
                      <Disc className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                      <div className="min-w-0 flex-1">
                        <p className="font-medium text-white truncate">{a.title}</p>
                        <p className="text-zinc-400 truncate text-[11px]">{a.albumArtist}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {searchResults.artists.length > 0 && (
              <div>
                <p className="text-[11px] font-bold text-zinc-500 uppercase tracking-wider mb-1.5 px-2">Artists</p>
                <div className="space-y-0.5">
                  {searchResults.artists.map((art) => (
                    <div
                      key={art.id}
                      className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg hover:bg-zinc-800 cursor-pointer text-xs"
                    >
                      <User className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                      <p className="font-medium text-white truncate">{art.name}</p>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {searchResults.tracks.length === 0 && searchResults.albums.length === 0 && searchResults.artists.length === 0 && (
              <p className="text-xs text-zinc-500 text-center py-4">No results found for &ldquo;{searchQuery}&rdquo;</p>
            )}
          </div>
        )}
      </div>

      {/* Admin Actions */}
      <div className="flex items-center gap-3">
        {user?.isAdmin && (
          <button
            onClick={handleTriggerScan}
            disabled={isScanning}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-all disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isScanning ? 'animate-spin text-rose-400' : ''}`} />
            {isScanning ? 'Scanning...' : 'Scan Library'}
          </button>
        )}
      </div>
    </header>
  )
}
