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
      setLibraries(libRes.items)
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
      setMessage('Scan started in background...')
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
    <div className="p-8 max-w-4xl">
      <h2 className="text-2xl font-bold text-white tracking-tight mb-2">Settings & Administration</h2>
      <p className="text-sm text-zinc-400 mb-8">Manage music directories, scan libraries, and inspect server health.</p>

      {message && (
        <div className="mb-6 p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center gap-3 text-sm text-emerald-400">
          <Check className="w-5 h-5 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {error && (
        <div className="mb-6 p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-sm text-rose-400">
          <AlertCircle className="w-5 h-5 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Libraries Section */}
      <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-2xl p-6 mb-8 backdrop-blur-md">
        <div className="flex items-center gap-3 mb-4">
          <HardDrive className="w-5 h-5 text-rose-400" />
          <h3 className="text-lg font-bold text-white">Music Libraries</h3>
        </div>

        <div className="space-y-3 mb-6">
          {libraries.length === 0 ? (
            <p className="text-sm text-zinc-500 italic">No music libraries configured yet.</p>
          ) : (
            libraries.map((lib) => (
              <div
                key={lib.id}
                className="bg-zinc-950/60 border border-zinc-800 rounded-xl p-4 flex items-center justify-between gap-4"
              >
                <div>
                  <h4 className="font-semibold text-sm text-white">{lib.name}</h4>
                  <p className="text-xs font-mono text-zinc-400 mt-0.5">{lib.path}</p>
                  <div className="flex items-center gap-3 mt-2 text-xs text-zinc-500 font-mono">
                    <span>{lib.trackCount} tracks</span>
                    <span>•</span>
                    <span>{lib.albumCount} albums</span>
                    <span>•</span>
                    <span className="capitalize">Status: {lib.scanStatus}</span>
                  </div>
                </div>

                {user?.isAdmin && (
                  <button
                    onClick={() => handleTriggerScan(lib.id)}
                    disabled={scanningMap[lib.id]}
                    className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-semibold text-white transition-all disabled:opacity-50"
                  >
                    <RefreshCw className={`w-3.5 h-3.5 ${scanningMap[lib.id] ? 'animate-spin text-rose-400' : ''}`} />
                    {scanningMap[lib.id] ? 'Scanning...' : 'Scan Now'}
                  </button>
                )}
              </div>
            ))
          )}
        </div>

        {user?.isAdmin && (
          <form onSubmit={handleAddLibrary} className="border-t border-zinc-800/60 pt-6">
            <h4 className="text-sm font-semibold text-zinc-200 mb-4 flex items-center gap-2">
              <FolderPlus className="w-4 h-4 text-amber-400" /> Add Music Folder
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1.5">Library Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Master FLAC Library"
                  value={newLibName}
                  onChange={(e) => setNewLibName(e.target.value)}
                  className="w-full bg-zinc-950/80 border border-zinc-800 rounded-xl px-3.5 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-rose-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-zinc-400 mb-1.5">Absolute Folder Path</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. D:\Music or /music"
                  value={newLibPath}
                  onChange={(e) => setNewLibPath(e.target.value)}
                  className="w-full bg-zinc-950/80 border border-zinc-800 rounded-xl px-3.5 py-2 text-sm text-white font-mono placeholder-zinc-500 focus:outline-none focus:border-rose-500"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-2.5 rounded-xl bg-rose-500 hover:bg-rose-600 text-white font-semibold text-xs transition-all shadow-md disabled:opacity-50"
            >
              {isSubmitting ? 'Adding...' : 'Add Library & Scan'}
            </button>
          </form>
        )}
      </div>

      {/* System Diagnostics Section */}
      {diagnostics && (
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-2xl p-6 backdrop-blur-md">
          <div className="flex items-center gap-3 mb-4">
            <Server className="w-5 h-5 text-amber-400" />
            <h3 className="text-lg font-bold text-white">System Diagnostics</h3>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-mono">
            <div className="bg-zinc-950/60 border border-zinc-800/60 rounded-xl p-3">
              <span className="text-zinc-500 block mb-1">Runtime</span>
              <span className="text-white font-semibold">{diagnostics.go_version}</span>
            </div>

            <div className="bg-zinc-950/60 border border-zinc-800/60 rounded-xl p-3">
              <span className="text-zinc-500 block mb-1">Platform</span>
              <span className="text-white font-semibold">{diagnostics.platform}</span>
            </div>

            <div className="bg-zinc-950/60 border border-zinc-800/60 rounded-xl p-3">
              <span className="text-zinc-500 block mb-1">Uptime</span>
              <span className="text-white font-semibold">{diagnostics.uptime_seconds}s</span>
            </div>

            <div className="bg-zinc-950/60 border border-zinc-800/60 rounded-xl p-3">
              <span className="text-zinc-500 block mb-1">FFmpeg Transcoder</span>
              <span className={diagnostics.ffmpeg.available ? 'text-emerald-400 font-semibold' : 'text-amber-400 font-semibold'}>
                {diagnostics.ffmpeg.available ? 'Available' : 'Disabled (Direct Play Active)'}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
