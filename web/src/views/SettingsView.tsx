import React, { useEffect, useState } from 'react'
import { FolderPlus, RefreshCw, Server, HardDrive, Check, AlertCircle } from 'lucide-react'
import type { Library, Diagnostics } from '../types'
import { api } from '../lib/api'
import { useAuthStore } from '../store/authStore'

export const SettingsView: React.FC = () => {
  const { user } = useAuthStore()
  const [libraries, setLibraries] = useState<Library[]>([])
  const [diagnostics, setDiagnostics] = useState<Diagnostics | null>(null)
  const [newLibName, setNewLibName] = useState('')
  const [newLibPath, setNewLibPath] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [scanningMap, setScanningMap] = useState<Record<string, boolean>>({})
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [libRes, diagRes] = await Promise.all([
        api.getLibraries(),
        api.getDiagnostics(),
      ])
      setLibraries(libRes.items || [])
      setDiagnostics(diagRes)
    } catch (err: any) {
      console.error('Failed to load settings data:', err)
    }
  }

  const handleAddLibrary = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newLibName || !newLibPath) return

    setIsSubmitting(true)
    setError(null)
    setMessage(null)

    try {
      const created = await api.createLibrary(newLibName, newLibPath)
      setLibraries([...libraries, created])
      setNewLibName('')
      setNewLibPath('')
      setMessage(`Library "${created.name}" created successfully!`)
      await handleTriggerScan(created.id)
    } catch (err: any) {
      setError(err.message || 'Failed to add library')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleTriggerScan = async (libId: string) => {
    try {
      setScanningMap((prev) => ({ ...prev, [libId]: true }))
      await api.triggerScan(libId)
      setMessage('Scan telemetry initiated in background...')
      setTimeout(() => {
        setScanningMap((prev) => ({ ...prev, [libId]: false }))
        loadData()
      }, 4000)
    } catch (err: any) {
      setError(err.message || 'Failed to trigger scan')
      setScanningMap((prev) => ({ ...prev, [libId]: false }))
    }
  }

  return (
    <div className="p-6 md:p-8 max-w-4xl space-y-8 select-none">
      <div>
        <span className="text-[11px] font-bold uppercase tracking-wider text-rose-400 font-hud">
          SYSTEM // TELEMETRY
        </span>
        <h2 className="text-3xl font-black text-white tracking-tight mt-0.5">
          Settings & Administration
        </h2>
        <p className="text-xs text-zinc-400 mt-1">
          Manage music directories, scan libraries, and inspect server health
        </p>
      </div>

      {message && (
        <div className="p-3.5 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 flex items-center gap-3 text-xs text-emerald-400 font-medium">
          <Check className="w-4 h-4 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {error && (
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-xs text-rose-400 font-medium">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Libraries Section */}
      <div className="bg-[#211e1c] border border-white/5 rounded-[24px] p-6 shadow-xl space-y-6">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-[#171514] border border-white/5 flex items-center justify-center text-rose-400">
            <HardDrive className="w-4 h-4" />
          </div>
          <h3 className="text-lg font-bold text-white tracking-tight">Audio Libraries</h3>
        </div>

        <div className="space-y-3">
          {libraries.length === 0 ? (
            <p className="text-xs text-zinc-500 font-hud">NO MUSIC LIBRARIES CONFIGURED YET.</p>
          ) : (
            libraries.map((lib) => (
              <div
                key={lib.id}
                className="bg-[#171514] border border-white/5 rounded-2xl p-4 flex items-center justify-between gap-4"
              >
                <div>
                  <h4 className="font-bold text-sm text-white">{lib.name}</h4>
                  <p className="text-xs font-hud text-zinc-400 mt-0.5">{lib.path}</p>
                  <div className="flex items-center gap-2.5 mt-2 text-xs text-zinc-500 font-hud">
                    <span>{lib.trackCount} tracks</span>
                    <span>•</span>
                    <span>{lib.albumCount} albums</span>
                    <span>•</span>
                    <span className="capitalize text-zinc-400">Status: {lib.scanStatus}</span>
                  </div>
                </div>

                {user?.isAdmin && (
                  <button
                    onClick={() => handleTriggerScan(lib.id)}
                    disabled={scanningMap[lib.id]}
                    className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[#2a2622] hover:bg-[#35302b] text-xs font-bold text-white transition-all disabled:opacity-50 cursor-pointer border border-white/5"
                  >
                    <RefreshCw
                      className={`w-3.5 h-3.5 ${scanningMap[lib.id] ? 'animate-spin text-rose-400' : ''}`}
                    />
                    {scanningMap[lib.id] ? 'Scanning...' : 'Scan Now'}
                  </button>
                )}
              </div>
            ))
          )}
        </div>

        {user?.isAdmin && (
          <form onSubmit={handleAddLibrary} className="border-t border-white/5 pt-6 space-y-4">
            <h4 className="text-xs font-bold text-zinc-300 uppercase tracking-wider font-hud flex items-center gap-2">
              <FolderPlus className="w-4 h-4 text-amber-400" /> ADD AUDIO DIRECTORY
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
                  Library Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Master Library"
                  value={newLibName}
                  onChange={(e) => setNewLibName(e.target.value)}
                  className="w-full bg-[#141211] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-white/30"
                />
              </div>

              <div>
                <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
                  Directory Path
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. /music or D:\Music"
                  value={newLibPath}
                  onChange={(e) => setNewLibPath(e.target.value)}
                  className="w-full bg-[#141211] border border-white/10 rounded-xl px-3.5 py-2.5 text-sm text-white font-hud placeholder-zinc-500 focus:outline-none focus:border-white/30"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="px-5 py-2.5 rounded-full bg-white hover:bg-zinc-200 text-black font-bold text-xs transition-all shadow-md disabled:opacity-50 cursor-pointer"
            >
              {isSubmitting ? 'Mounting...' : 'Add Library & Scan'}
            </button>
          </form>
        )}
      </div>

      {/* System Diagnostics Section */}
      {diagnostics && (
        <div className="bg-[#211e1c] border border-white/5 rounded-[24px] p-6 shadow-xl">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-9 h-9 rounded-xl bg-[#171514] border border-white/5 flex items-center justify-center text-amber-400">
              <Server className="w-4 h-4" />
            </div>
            <h3 className="text-lg font-bold text-white tracking-tight">System Diagnostics</h3>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-hud">
            <div className="bg-[#171514] border border-white/5 rounded-2xl p-4">
              <span className="text-zinc-500 block mb-1 text-[10px] tracking-wider uppercase">Runtime</span>
              <span className="text-white font-bold">{diagnostics.go_version}</span>
            </div>

            <div className="bg-[#171514] border border-white/5 rounded-2xl p-4">
              <span className="text-zinc-500 block mb-1 text-[10px] tracking-wider uppercase">Platform</span>
              <span className="text-white font-bold">{diagnostics.platform}</span>
            </div>

            <div className="bg-[#171514] border border-white/5 rounded-2xl p-4">
              <span className="text-zinc-500 block mb-1 text-[10px] tracking-wider uppercase">Uptime</span>
              <span className="text-white font-bold">{diagnostics.uptime_seconds}s</span>
            </div>

            <div className="bg-[#171514] border border-white/5 rounded-2xl p-4">
              <span className="text-zinc-500 block mb-1 text-[10px] tracking-wider uppercase">Transcoder</span>
              <span
                className={
                  diagnostics.ffmpeg.available
                    ? 'text-emerald-400 font-bold'
                    : 'text-amber-400 font-bold'
                }
              >
                {diagnostics.ffmpeg.available ? 'FFmpeg Active' : 'Direct Play'}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
