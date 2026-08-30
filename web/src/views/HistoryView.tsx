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
      setHistory(res.items || [])
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
    <div className="p-6 md:p-8 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-5 pb-6 border-b border-white/5">
        <div className="w-16 h-16 rounded-[22px] bg-[#221f1c] border border-white/10 flex items-center justify-center text-amber-500 shadow-xl">
          <History className="w-8 h-8" />
        </div>
        <div>
          <span className="text-[11px] font-bold uppercase tracking-wider text-amber-400 font-hud">
            TELEMETRY LOG
          </span>
          <h2 className="text-3xl font-black text-white tracking-tight mt-0.5">Listening History</h2>
          <p className="text-xs text-zinc-400 mt-1 font-hud">{history.length} logged sessions</p>
        </div>
      </div>

      {isLoading ? (
        <div className="p-8 text-center text-zinc-500 font-hud text-sm animate-pulse">
          LOADING PLAYBACK LOGS...
        </div>
      ) : history.length === 0 ? (
        <div className="py-16 text-center text-zinc-500">
          <History className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
          <h3 className="text-lg font-bold text-zinc-400">No playback history</h3>
          <p className="text-xs text-zinc-600 mt-1">Tracks you play will be recorded here</p>
        </div>
      ) : (
        <div className="space-y-1.5">
          {history.map((item) => (
            <div
              key={item.id}
              onClick={() => handlePlay(item)}
              className="flex items-center justify-between p-3.5 rounded-2xl bg-[#211e1c] hover:bg-[#2b2724] border border-white/5 hover:border-white/15 cursor-pointer transition-all group shadow-sm hover:shadow-lg"
            >
              <div className="flex items-center gap-3.5 min-w-0">
                <div className="w-10 h-10 rounded-xl bg-[#171514] border border-white/5 flex items-center justify-center text-zinc-400 group-hover:text-white transition-colors shrink-0">
                  <Play className="w-4 h-4 fill-current hidden group-hover:inline-block ml-0.5" />
                  <Music className="w-4 h-4 group-hover:hidden" />
                </div>
                <div className="min-w-0">
                  <p className="font-bold text-sm text-white truncate group-hover:text-white transition-colors">
                    {item.trackTitle || 'Track'}
                  </p>
                  <p className="text-xs text-zinc-400 truncate mt-0.5">{item.artistName || 'Unknown Artist'}</p>
                </div>
              </div>

              <div className="flex items-center gap-4 text-xs font-hud text-zinc-400 shrink-0">
                {item.completed && (
                  <span className="flex items-center gap-1 text-emerald-400 text-[11px]">
                    <CheckCircle2 className="w-3.5 h-3.5" /> Scrobbled
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
