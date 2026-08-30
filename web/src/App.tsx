import React, { useEffect, useState } from 'react'
import { useAuthStore } from './store/authStore'
import { Sidebar } from './components/Sidebar'
import type { ViewType } from './components/Sidebar'
import { Header } from './components/Header'
import { PlayerBar } from './components/PlayerBar'
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

export const App: React.FC = () => {
  const { isAuthenticated, isLoading, checkAuth } = useAuthStore()
  const [currentView, setCurrentView] = useState<ViewType>('albums')
  const [selectedAlbum, setSelectedAlbum] = useState<Album | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showSetup, setShowSetup] = useState(false)

  useEffect(() => {
    checkAuth()
  }, [])

  if (isLoading) {
    return (
      <div className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center text-white">
        <div className="w-16 h-16 rounded-2xl bg-gradient-to-tr from-amber-500 to-rose-500 flex items-center justify-center shadow-2xl shadow-rose-500/30 animate-pulse mb-4">
          <span className="font-black text-2xl text-black font-mono">nE</span>
        </div>
        <p className="text-sm font-medium text-zinc-400 flex items-center gap-1.5">
          Initializing nE Server <Sparkles className="w-3.5 h-3.5 text-amber-400 animate-spin" />
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
    <div className="h-screen w-screen flex flex-col bg-zinc-950 text-zinc-100 overflow-hidden font-sans">
      {/* Upper Area: Sidebar + Main Content */}
      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          currentView={currentView}
          onViewChange={(view) => {
            setCurrentView(view)
            if (view !== 'album-detail') setSelectedAlbum(null)
          }}
        />

        <main className="flex-1 flex flex-col min-w-0 bg-gradient-to-b from-zinc-900/30 to-zinc-950 overflow-hidden">
          <Header
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            onSelectAlbum={handleSelectAlbum}
          />
          <div className="flex-1 overflow-y-auto overflow-x-hidden">
            {renderMainView()}
          </div>
        </main>
      </div>

      {/* Bottom Fixed Audio Player */}
      <PlayerBar />
    </div>
  )
}

export default App
