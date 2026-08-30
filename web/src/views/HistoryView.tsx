import React, { useEffect, useState } from 'react'
import { History, Play, Music, CheckCircle2 } from 'lucide-react'
import type { PlaybackRecord } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

export const HistoryView: React.FC = () => {
  const [history, setHistory] = useState<PlaybackRecord[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const { playTrack } = usePlayerStore()

  useEffect(() => {
    loadHistory()
  }, [])

  const loadHistory = async () => {
    try {
      setIsLoading(true)
      const res = await api.getRecentHistory(50)
      setHistory(res.items)
    } catch (err) {
      console.error('Failed to load history:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const formatPlayedAt = (isoStr: string) => {
    const d = new Date(isoStr)
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  const handlePlay = async (record: PlaybackRecord) => {
    try {
      const track = await api.getTrack(record.trackId)
      playTrack(track)
    } catch (err) {
      console.error('Play track failed:', err)
    }
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center gap-4 mb-8 pb-6 border-b border-zinc-800/60">
        <div className="w-16 h-16 rounded-2xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-rose-400 shadow-md">
          <History className="w-8 h-8" />
        </div>
        <div>
          <h2 className="text-2xl font-bold text-white tracking-tight">Listening History</h2>
          <p className="text-xs text-zinc-400 mt-1 font-mono">Recent playback sessions and scrobbles</p>
        </div>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-zinc-500 font-medium animate-pulse">
          Loading playback history...
        </div>
      ) : history.length === 0 ? (
        <div className="p-12 text-center text-zinc-500">
          <History className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
          <h3 className="text-lg font-semibold text-zinc-400">No playback history</h3>
          <p className="text-sm text-zinc-600 mt-1">Tracks you play will be recorded here</p>
        </div>
      ) : (
        <div className="space-y-1.5">
          {history.map((item) => (
            <div
              key={item.id}
              onClick={() => handlePlay(item)}
              className="flex items-center justify-between p-3 rounded-xl bg-zinc-900/40 hover:bg-zinc-900 border border-zinc-800/40 hover:border-zinc-700 cursor-pointer transition-all group"
            >
              <div className="flex items-center gap-3.5 min-w-0">
                <div className="w-10 h-10 rounded-lg bg-zinc-800 flex items-center justify-center text-zinc-400 group-hover:text-rose-400 transition-colors shrink-0">
                  <Play className="w-4 h-4 fill-current hidden group-hover:inline-block" />
                  <Music className="w-4 h-4 group-hover:hidden" />
                </div>
                <div className="min-w-0">
                  <p className="font-semibold text-sm text-white truncate group-hover:text-rose-400 transition-colors">
                    {item.trackTitle || 'Track'}
                  </p>
                  <p className="text-xs text-zinc-400 truncate">{item.artistName || 'Unknown Artist'}</p>
                </div>
              </div>

              <div className="flex items-center gap-4 text-xs font-mono text-zinc-500 shrink-0">
                {item.completed && (
                  <span className="flex items-center gap-1 text-emerald-400">
                    <CheckCircle2 className="w-3.5 h-3.5" /> Scrobble
                  </span>
                )}
                <span>{formatPlayedAt(item.playedAt)}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
