import React, { useEffect, useState, useRef } from 'react'
import { Search, Disc, Music, X, Loader2, Bell, Settings } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'
import type { SearchResult, Album, Track } from '../types'

interface HeaderProps {
  searchQuery: string
  onSearchChange: (q: string) => void
  onSelectAlbum?: (album: Album) => void
  onOpenSettings?: () => void
}

export const Header: React.FC<HeaderProps> = ({
  searchQuery,
  onSearchChange,
  onSelectAlbum,
  onOpenSettings,
}) => {
  const { user } = useAuthStore()
  const { playTrack } = usePlayerStore()
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
    }, 200)

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
    <header className="h-16 px-8 flex items-center justify-between gap-6 z-30 select-none">
      {/* Pill Search Input matching Image 2 */}
      <div ref={searchRef} className="relative max-w-md w-full">
        {isSearching ? (
          <Loader2 className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-rose-400 animate-spin" />
        ) : (
          <Search className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
        )}
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          onFocus={() => {
            if (searchResults) setShowResults(true)
          }}
          placeholder="Search for songs, artists..."
          className="w-full bg-[#24211e] border border-white/5 rounded-full pl-10 pr-8 py-2 text-sm text-zinc-200 placeholder-zinc-400 focus:outline-none focus:border-white/20 transition-all shadow-inner"
        />
        {searchQuery && (
          <button
            onClick={() => {
              onSearchChange('')
              setSearchResults(null)
              setShowResults(false)
            }}
            className="absolute right-3.5 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}

        {/* Live Search Results Dropdown */}
        {showResults && searchResults && (
          <div className="absolute left-0 right-0 top-full mt-2 bg-[#201d1b] border border-white/10 rounded-2xl shadow-2xl overflow-hidden backdrop-blur-2xl max-h-96 overflow-y-auto p-3 space-y-3 z-50">
            {searchResults.tracks.length > 0 && (
              <div>
                <p className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1 px-2 font-hud">
                  Tracks
                </p>
                <div className="space-y-0.5">
                  {searchResults.tracks.map((t) => (
                    <div
                      key={t.id}
                      onClick={() => handleTrackClick(t)}
                      className="flex items-center gap-2.5 px-3 py-2 rounded-xl hover:bg-white/5 cursor-pointer text-xs group"
                    >
                      <Music className="w-3.5 h-3.5 text-rose-400 shrink-0" />
                      <div className="min-w-0 flex-1">
                        <p className="font-semibold text-white truncate">{t.title}</p>
                        <p className="text-zinc-400 truncate text-[11px]">{t.rawArtist}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {searchResults.albums.length > 0 && (
              <div>
                <p className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1 px-2 font-hud">
                  Albums
                </p>
                <div className="space-y-0.5">
                  {searchResults.albums.map((a) => (
                    <div
                      key={a.id}
                      onClick={() => handleAlbumClick(a)}
                      className="flex items-center gap-2.5 px-3 py-2 rounded-xl hover:bg-white/5 cursor-pointer text-xs group"
                    >
                      <Disc className="w-3.5 h-3.5 text-amber-400 shrink-0" />
                      <div className="min-w-0 flex-1">
                        <p className="font-semibold text-white truncate">{a.title}</p>
                        <p className="text-zinc-400 truncate text-[11px]">{a.albumArtist}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Right Icons Row matching Image 2: Bell, Settings, User Avatar */}
      <div className="flex items-center gap-3">
        <button
          title="Notifications"
          className="p-2 text-zinc-400 hover:text-white rounded-full bg-[#24211e] hover:bg-[#2e2a27] border border-white/5 transition-colors cursor-pointer"
        >
          <Bell className="w-4 h-4" />
        </button>

        <button
          onClick={onOpenSettings}
          title="Settings"
          className="p-2 text-zinc-400 hover:text-white rounded-full bg-[#24211e] hover:bg-[#2e2a27] border border-white/5 transition-colors cursor-pointer"
        >
          <Settings className="w-4 h-4" />
        </button>

        {/* User Avatar Circle Pill matching Image 2 */}
        <div className="w-9 h-9 rounded-full bg-zinc-600 border border-white/10 flex items-center justify-center font-bold text-xs text-white shadow-inner cursor-pointer">
          {user?.username.charAt(0).toUpperCase() || 'N'}
        </div>
      </div>
    </header>
  )
}
