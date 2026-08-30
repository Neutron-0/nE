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
import { Sparkles } from 'lucide-react'
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts'

export const App: React.FC = () => {
  const { isAuthenticated, isLoading, checkAuth } = useAuthStore()
  const [currentView, setCurrentView] = useState<ViewType>('albums')
  const [selectedAlbum, setSelectedAlbum] = useState<Album | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showSetup, setShowSetup] = useState(false)

  useKeyboardShortcuts()

  useEffect(() => {
    checkAuth()
  }, [])

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#121110] flex flex-col items-center justify-center text-white font-hud">
        <div className="w-16 h-16 rounded-2xl bg-[#23201d] border border-white/10 flex items-center justify-center shadow-2xl animate-pulse mb-4">
          <span className="font-black text-2xl text-white">nE</span>
        </div>
        <p className="text-xs font-semibold tracking-widest text-zinc-400 flex items-center gap-2">
          ESTABLISHING NEURAL AUDIO LINK <Sparkles className="w-3.5 h-3.5 text-rose-500 animate-spin" />
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
        return <AlbumsView onSelectAlbum={handleSelectAlbum} searchQuery={searchQuery} />
      case 'album-detail':
        return selectedAlbum ? (
          <AlbumDetailView album={selectedAlbum} onBack={() => setCurrentView('albums')} />
        ) : (
          <AlbumsView onSelectAlbum={handleSelectAlbum} searchQuery={searchQuery} />
        )
      case 'artists':
        return <ArtistsView searchQuery={searchQuery} />
      case 'tracks':
        return <TracksView searchQuery={searchQuery} />
      case 'playlists':
        return <PlaylistsView />
      case 'favorites':
        return <FavoritesView />
      case 'history':
        return <HistoryView />
      case 'settings':
        return <SettingsView />
      default:
        return <AlbumsView onSelectAlbum={handleSelectAlbum} searchQuery={searchQuery} />
    }
  }

  return (
    <div className="h-screen w-screen bg-[#0e0d0c] p-2 md:p-3.5 flex items-center justify-center select-none overflow-hidden font-sans">
      {/* Outer Floating Hardware Bezel matching Image 2 */}
      <div className="w-full h-full bg-[#181615] rounded-[28px] md:rounded-[32px] border border-[#2f2b27] flex overflow-hidden shadow-[0_25px_70px_rgba(0,0,0,0.85)] relative">
        {/* Column 1: Left Navigation & Playlists matching Image 2 */}
        <Sidebar
          currentView={currentView}
          onViewChange={(view) => {
            setCurrentView(view)
            if (view !== 'album-detail') setSelectedAlbum(null)
          }}
        />

        {/* Column 2: Center Main Content Stream matching Image 2 */}
        <main className="flex-1 flex flex-col min-w-0 bg-[#1e1b19] overflow-hidden border-r border-[#2e2a26]">
          <Header
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            onSelectAlbum={handleSelectAlbum}
            onOpenSettings={() => setCurrentView('settings')}
          />
          <div className="flex-1 overflow-y-auto overflow-x-hidden">
            {renderMainView()}
          </div>
        </main>

        {/* Column 3: Dedicated Right Now Playing & Audio HUD Panel matching Image 2 */}
        <NowPlayingPanel />
      </div>
    </div>
  )
}

export default App
