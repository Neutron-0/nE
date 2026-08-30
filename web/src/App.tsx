import React, { useEffect, useState } from 'react'
import { useAuthStore } from './store/authStore'
import { Sidebar } from './components/Sidebar'
import type { ViewType } from './components/Sidebar'
import { Header } from './components/Header'
import { NowPlayingPanel } from './components/NowPlayingPanel'
import { LoginView } from './views/LoginView'
import { SetupView } from './views/SetupView'
import { AlbumsView } from './views/AlbumsView'
import { ArtistsView } from './views/ArtistsView'
import { TracksView } from './views/TracksView'
import { AlbumDetailView } from './views/AlbumDetailView'
import { PlaylistsView } from './views/PlaylistsView'
import { FavoritesView } from './views/FavoritesView'
import { HistoryView } from './views/HistoryView'
import { SettingsView } from './views/SettingsView'
import type { Album } from './types'
import { Sparkles, UploadCloud } from 'lucide-react'
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts'
import { api } from './lib/api'

export const App: React.FC = () => {
  const { isAuthenticated, isLoading, checkAuth } = useAuthStore()
  const [currentView, setCurrentView] = useState<ViewType>('albums')
  const [selectedAlbum, setSelectedAlbum] = useState<Album | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showSetup, setShowSetup] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  // Drag and drop state
  const [isDragOver, setIsDragOver] = useState(false)

  useKeyboardShortcuts()

  useEffect(() => {
    checkAuth()
  }, [])

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.dataTransfer.types.includes('Files')) {
      setIsDragOver(true)
    }
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)
  }

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)

    const files = e.dataTransfer.files
    if (!files || files.length === 0) return

    try {
      await api.uploadAudio(files)
      setRefreshKey((k) => k + 1)
    } catch (err) {
      console.error('Drag upload failed:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-black flex flex-col items-center justify-center text-white font-hud">
        <div className="w-16 h-16 rounded-2xl liquid-glass flex items-center justify-center shadow-2xl animate-pulse mb-4">
          <span className="font-black text-2xl text-white">nE</span>
        </div>
        <p className="text-xs font-semibold tracking-widest text-zinc-400 flex items-center gap-2">
          ESTABLISHING NEURAL AUDIO LINK <Sparkles className="w-3.5 h-3.5 text-white animate-spin" />
        </p>
      </div>
    )
  }

  if (!isAuthenticated) {
    if (showSetup) {
      return <SetupView onBackToLogin={() => setShowSetup(false)} />
    }
    return <LoginView onGoToSetup={() => setShowSetup(true)} />
  }

  const handleSelectAlbum = (album: Album) => {
    setSelectedAlbum(album)
    setCurrentView('album-detail')
  }

  const renderMainView = () => {
    switch (currentView) {
      case 'albums':
        return (
          <AlbumsView
            key={`albums-${refreshKey}`}
            onSelectAlbum={handleSelectAlbum}
            searchQuery={searchQuery}
          />
        )
      case 'album-detail':
        return selectedAlbum ? (
          <AlbumDetailView
            key={`album-${selectedAlbum.id}-${refreshKey}`}
            album={selectedAlbum}
            onBack={() => setCurrentView('albums')}
          />
        ) : (
          <AlbumsView
            key={`albums-${refreshKey}`}
            onSelectAlbum={handleSelectAlbum}
            searchQuery={searchQuery}
          />
        )
      case 'artists':
        return <ArtistsView key={`artists-${refreshKey}`} searchQuery={searchQuery} />
      case 'tracks':
        return <TracksView key={`tracks-${refreshKey}`} searchQuery={searchQuery} />
      case 'playlists':
        return <PlaylistsView key={`playlists-${refreshKey}`} />
      case 'favorites':
        return <FavoritesView key={`favs-${refreshKey}`} />
      case 'history':
        return <HistoryView key={`history-${refreshKey}`} />
      case 'settings':
        return <SettingsView key={`settings-${refreshKey}`} />
      default:
        return (
          <AlbumsView
            key={`albums-${refreshKey}`}
            onSelectAlbum={handleSelectAlbum}
            searchQuery={searchQuery}
          />
        )
    }
  }

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className="h-screen w-screen bg-black p-2 md:p-3 flex items-center justify-center select-none overflow-hidden font-sans relative"
    >
      {/* Global Drag & Drop Liquid Glass Overlay */}
      {isDragOver && (
        <div className="absolute inset-4 z-50 rounded-[28px] md:rounded-[32px] liquid-glass flex flex-col items-center justify-center text-center p-8 backdrop-blur-3xl border-2 border-dashed border-white/40 pointer-events-none">
          <UploadCloud className="w-16 h-16 text-white animate-bounce mb-4" />
          <h3 className="text-3xl font-black text-white tracking-tight">Drop Audio Files Here</h3>
          <p className="text-sm text-zinc-400 mt-2 font-hud">
            Files will be automatically saved, cataloged, and ingested into database
          </p>
        </div>
      )}

      {/* Outer Floating Solid Black Console with Liquid Glass */}
      <div className="w-full h-full bg-black/95 rounded-[28px] md:rounded-[32px] border border-white/[0.08] flex overflow-hidden shadow-[0_30px_100px_rgba(0,0,0,1)] relative backdrop-blur-3xl">
        {/* Column 1: Left Navigation & Playlists matching Image 2 */}
        <Sidebar
          currentView={currentView}
          onViewChange={(view) => {
            setCurrentView(view)
            if (view !== 'album-detail') setSelectedAlbum(null)
          }}
        />

        {/* Column 2: Center Main Content Stream matching Image 2 */}
        <main className="flex-1 flex flex-col min-w-0 bg-[#060606]/80 backdrop-blur-2xl overflow-hidden border-r border-white/[0.06]">
          <Header
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            onSelectAlbum={handleSelectAlbum}
            onOpenSettings={() => setCurrentView('settings')}
            onUploadSuccess={() => setRefreshKey((k) => k + 1)}
          />
          <div className="flex-1 overflow-y-auto overflow-x-hidden">
            {renderMainView()}
          </div>
        </main>

        {/* Column 3: Dedicated Right Now Playing & Interactive Audio HUD matching Image 2 */}
        <NowPlayingPanel />
      </div>
    </div>
  )
}

export default App
