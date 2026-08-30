import React, { useEffect, useState } from 'react'
import { ListMusic, Plus, Play, Trash2, Sparkles, Clock, Flame, History as HistoryIcon, Star, Shuffle } from 'lucide-react'
import type { Playlist, Track, SmartPlaylistSummary } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'
import { TracklistTable } from '../components/TracklistTable'

export const PlaylistsView: React.FC = () => {
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [smartPlaylists, setSmartPlaylists] = useState<SmartPlaylistSummary[]>([])
  const [selectedPlaylist, setSelectedPlaylist] = useState<{ id: string; name: string; comment?: string; tracks: Track[] } | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newPlaylistName, setNewPlaylistName] = useState('')
  const [newPlaylistComment, setNewPlaylistComment] = useState('')
  const { playTrack } = usePlayerStore()

  useEffect(() => {
    loadAll()
  }, [])

  const loadAll = async () => {
    try {
      setIsLoading(true)
      const [customRes, smartRes] = await Promise.all([
        api.getPlaylists(),
        api.getSmartPlaylists(),
      ])
      setPlaylists(customRes.items || [])
      setSmartPlaylists(smartRes || [])
    } catch (err) {
      console.error('Failed to load playlists:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newPlaylistName.trim()) return
    try {
      const created = await api.createPlaylist(newPlaylistName.trim(), newPlaylistComment.trim())
      setPlaylists([created, ...playlists])
      setNewPlaylistName('')
      setNewPlaylistComment('')
      setShowCreateModal(false)
    } catch (err) {
      console.error('Create playlist failed:', err)
    }
  }

  const handleSelectCustomPlaylist = async (p: Playlist) => {
    try {
      const full = await api.getPlaylist(p.id)
      setSelectedPlaylist({
        id: full.id,
        name: full.name,
        comment: full.comment,
        tracks: full.tracks || [],
      })
    } catch (err) {
      console.error('Load playlist failed:', err)
    }
  }

  const handleSelectSmartPlaylist = async (sp: SmartPlaylistSummary) => {
    try {
      const tracksRes = await api.getSmartPlaylistTracks(sp.id, 50)
      setSelectedPlaylist({
        id: sp.id,
        name: sp.name,
        comment: sp.description,
        tracks: tracksRes.items || [],
      })
    } catch (err) {
      console.error('Load smart playlist failed:', err)
    }
  }

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation()
    if (!confirm('Are you sure you want to delete this playlist?')) return
    try {
      await api.deletePlaylist(id)
      setPlaylists(playlists.filter((p) => p.id !== id))
      if (selectedPlaylist?.id === id) {
        setSelectedPlaylist(null)
      }
    } catch (err) {
      console.error('Delete playlist failed:', err)
    }
  }

  const getSmartIcon = (iconName: string) => {
    switch (iconName) {
      case 'Clock': return <Clock className="w-5 h-5 text-amber-400" />
      case 'Flame': return <Flame className="w-5 h-5 text-rose-500" />
      case 'History': return <HistoryIcon className="w-5 h-5 text-blue-400" />
      case 'Star': return <Star className="w-5 h-5 text-yellow-400 fill-yellow-400/30" />
      case 'Shuffle': return <Shuffle className="w-5 h-5 text-emerald-400" />
      default: return <Sparkles className="w-5 h-5 text-rose-400" />
    }
  }

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-hud text-sm">
        <p className="animate-pulse">LOADING SMART PLAYLISTS...</p>
      </div>
    )
  }

  return (
    <div className="p-6 md:p-8 space-y-8">
      {/* Smart Playlists Section */}
      <div>
        <div className="flex items-center gap-2 mb-4 px-1">
          <Sparkles className="w-4 h-4 text-rose-400" />
          <h3 className="text-xl font-bold text-white tracking-tight">Dynamic Smart Mixes</h3>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3.5">
          {smartPlaylists.map((sp) => (
            <div
              key={sp.id}
              onClick={() => handleSelectSmartPlaylist(sp)}
              className={`p-4 rounded-2xl border transition-all cursor-pointer flex flex-col justify-between group ${
                selectedPlaylist?.id === sp.id
                  ? 'bg-[#2f2a26] border-rose-500/50 shadow-lg'
                  : 'bg-[#211e1c] hover:bg-[#2a2623] border-white/5 hover:border-white/15'
              }`}
            >
              <div className="flex items-center gap-3 mb-2">
                <div className="w-10 h-10 rounded-xl bg-[#181615] flex items-center justify-center shrink-0 border border-white/5">
                  {getSmartIcon(sp.icon)}
                </div>
                <h4 className="font-bold text-sm text-white truncate">{sp.name}</h4>
              </div>
              <p className="text-xs text-zinc-400 line-clamp-2">{sp.description}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Custom Playlists Header */}
      <div>
        <div className="flex items-center justify-between mb-4 px-1">
          <div>
            <h3 className="text-xl font-bold text-white tracking-tight">My Playlists</h3>
            <p className="text-xs text-zinc-400 mt-0.5">Create and organize custom music sets</p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-full bg-white hover:bg-zinc-200 text-black font-bold text-xs transition-all shadow-md cursor-pointer"
          >
            <Plus className="w-4 h-4" /> New Playlist
          </button>
        </div>

        {/* Playlist Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {playlists.length === 0 ? (
            <div className="col-span-full py-8 text-center text-xs text-zinc-500 border border-dashed border-white/10 rounded-2xl font-hud">
              NO CUSTOM PLAYLISTS CREATED YET. CLICK "+ NEW PLAYLIST" TO BEGIN.
            </div>
          ) : (
            playlists.map((pl) => (
              <div
                key={pl.id}
                onClick={() => handleSelectCustomPlaylist(pl)}
                className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-center justify-between group ${
                  selectedPlaylist?.id === pl.id
                    ? 'bg-[#2f2a26] border-rose-500/50 shadow-lg'
                    : 'bg-[#211e1c] hover:bg-[#2a2623] border-white/5 hover:border-white/15'
                }`}
              >
                <div className="flex items-center gap-3.5 min-w-0">
                  <div className="w-12 h-12 rounded-xl bg-[#181615] border border-white/5 flex items-center justify-center text-rose-400 shrink-0">
                    <ListMusic className="w-6 h-6" />
                  </div>
                  <div className="min-w-0">
                    <h4 className="font-bold text-sm text-white truncate">{pl.name}</h4>
                    <p className="text-xs text-zinc-400 mt-0.5 font-hud">{pl.trackCount} tracks</p>
                  </div>
                </div>

                <button
                  onClick={(e) => handleDelete(e, pl.id)}
                  className="p-2 text-zinc-500 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Selected Playlist Tracks Section */}
      {selectedPlaylist && (
        <div className="bg-[#211e1c] border border-white/5 rounded-[24px] p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between mb-2">
            <div>
              <h3 className="text-2xl font-black text-white">{selectedPlaylist.name}</h3>
              {selectedPlaylist.comment && (
                <p className="text-xs text-zinc-400 mt-1">{selectedPlaylist.comment}</p>
              )}
            </div>
            {selectedPlaylist.tracks && selectedPlaylist.tracks.length > 0 && (
              <button
                onClick={() => playTrack(selectedPlaylist.tracks[0], selectedPlaylist.tracks)}
                className="flex items-center gap-2 px-5 py-2.5 rounded-full bg-white hover:bg-zinc-200 text-black text-xs font-bold transition-all shadow-md cursor-pointer"
              >
                <Play className="w-4 h-4 fill-current ml-0.5" /> Play All
              </button>
            )}
          </div>

          <TracklistTable tracks={selectedPlaylist.tracks} />
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-4 z-50 backdrop-blur-xl">
          <div className="bg-[#1f1c1a] border border-white/10 rounded-[28px] p-6 max-w-md w-full shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Create New Playlist</h3>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
                  Playlist Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Late Night Synthwave"
                  value={newPlaylistName}
                  onChange={(e) => setNewPlaylistName(e.target.value)}
                  className="w-full bg-[#141211] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-white/30"
                />
              </div>

              <div>
                <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
                  Description (Optional)
                </label>
                <input
                  type="text"
                  placeholder="e.g. Chill tracks for focus"
                  value={newPlaylistComment}
                  onChange={(e) => setNewPlaylistComment(e.target.value)}
                  className="w-full bg-[#141211] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-white/30"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-xl text-xs font-semibold text-zinc-400 hover:text-white cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-5 py-2.5 rounded-xl bg-white hover:bg-zinc-200 text-black text-xs font-bold transition-all shadow-md cursor-pointer"
                >
                  Create Playlist
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
