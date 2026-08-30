import React, { useEffect, useState } from 'react'
import { ListMusic, Plus, Play, Trash2 } from 'lucide-react'
import type { Playlist } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

export const PlaylistsView: React.FC = () => {
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  const [selectedPlaylist, setSelectedPlaylist] = useState<Playlist | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newPlaylistName, setNewPlaylistName] = useState('')
  const [newPlaylistComment, setNewPlaylistComment] = useState('')
  const { playTrack } = usePlayerStore()

  useEffect(() => {
    loadPlaylists()
  }, [])

  const loadPlaylists = async () => {
    try {
      setIsLoading(true)
      const res = await api.getPlaylists()
      setPlaylists(res.items)
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

  const handleSelectPlaylist = async (p: Playlist) => {
    try {
      const full = await api.getPlaylist(p.id)
      setSelectedPlaylist(full)
    } catch (err) {
      console.error('Load playlist failed:', err)
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

  const formatDuration = (secs: number) => {
    if (!secs || isNaN(secs)) return '0:00'
    const m = Math.floor(secs / 60)
    const s = Math.floor(secs % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
  }

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-medium animate-pulse">
        Loading playlists...
      </div>
    )
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-white tracking-tight">Playlists</h2>
          <p className="text-sm text-zinc-400 mt-0.5">Create custom playlists and organize tracks</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-rose-500 hover:bg-rose-600 text-white font-semibold text-xs transition-all shadow-md"
        >
          <Plus className="w-4 h-4" /> New Playlist
        </button>
      </div>

      {/* Playlist Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 mb-8">
        {playlists.map((pl) => (
          <div
            key={pl.id}
            onClick={() => handleSelectPlaylist(pl)}
            className={`p-4 rounded-xl border transition-all cursor-pointer flex items-center justify-between group ${
              selectedPlaylist?.id === pl.id
                ? 'bg-zinc-800/90 border-rose-500/50 shadow-lg shadow-rose-500/10'
                : 'bg-zinc-900/40 hover:bg-zinc-900 border-zinc-800/50 hover:border-zinc-700'
            }`}
          >
            <div className="flex items-center gap-3.5 min-w-0">
              <div className="w-12 h-12 rounded-lg bg-zinc-800 flex items-center justify-center text-rose-400 shrink-0">
                <ListMusic className="w-6 h-6" />
              </div>
              <div className="min-w-0">
                <h4 className="font-semibold text-sm text-white truncate">{pl.name}</h4>
                <p className="text-xs text-zinc-500 mt-0.5 font-mono">{pl.trackCount} tracks</p>
              </div>
            </div>

            <button
              onClick={(e) => handleDelete(e, pl.id)}
              className="p-1.5 text-zinc-600 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        ))}
      </div>

      {/* Selected Playlist Tracks */}
      {selectedPlaylist && (
        <div className="bg-zinc-900/40 border border-zinc-800/60 rounded-2xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-xl font-bold text-white">{selectedPlaylist.name}</h3>
              {selectedPlaylist.comment && (
                <p className="text-xs text-zinc-400 mt-1">{selectedPlaylist.comment}</p>
              )}
            </div>
            {selectedPlaylist.tracks && selectedPlaylist.tracks.length > 0 && (
              <button
                onClick={() => playTrack(selectedPlaylist.tracks![0], selectedPlaylist.tracks)}
                className="flex items-center gap-2 px-4 py-2 rounded-full bg-rose-500 text-white text-xs font-semibold hover:scale-105 transition-all shadow-md"
              >
                <Play className="w-3.5 h-3.5 fill-current" /> Play All
              </button>
            )}
          </div>

          <div className="divide-y divide-zinc-900/60">
            {(!selectedPlaylist.tracks || selectedPlaylist.tracks.length === 0) ? (
              <p className="text-xs text-zinc-500 py-6 text-center">No tracks in this playlist yet.</p>
            ) : (
              selectedPlaylist.tracks.map((track, idx) => (
                <div
                  key={track.id}
                  onClick={() => playTrack(track, selectedPlaylist.tracks)}
                  className="grid grid-cols-12 items-center py-2.5 px-3 rounded-lg hover:bg-zinc-800/60 cursor-pointer text-xs group"
                >
                  <span className="col-span-1 text-zinc-500 text-center">{idx + 1}</span>
                  <div className="col-span-6 min-w-0 pr-4">
                    <p className="font-medium text-white truncate">{track.title}</p>
                    <p className="text-zinc-500 truncate text-[11px]">{track.rawArtist}</p>
                  </div>
                  <span className="col-span-4 text-zinc-500 truncate">{track.albumTitle}</span>
                  <span className="col-span-1 text-right text-zinc-500 font-mono">
                    {formatDuration(track.duration)}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50 backdrop-blur-sm">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-6 max-w-md w-full shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Create New Playlist</h3>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1.5">Playlist Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Late Night Synthwave"
                  value={newPlaylistName}
                  onChange={(e) => setNewPlaylistName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3.5 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-rose-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1.5">Description (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. Favorites for coding"
                  value={newPlaylistComment}
                  onChange={(e) => setNewPlaylistComment(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3.5 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-rose-500"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-xl text-xs font-semibold text-zinc-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-rose-500 hover:bg-rose-600 text-white text-xs font-semibold transition-all shadow-md"
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
