import React from 'react'
import {
  Disc,
  Music,
  User as UserIcon,
  Settings,
  LogOut,
  Sparkles,
  Heart,
  History,
  ListMusic,
} from 'lucide-react'
import { useAuthStore } from '../store/authStore'

export type ViewType =
  | 'albums'
  | 'artists'
  | 'tracks'
  | 'playlists'
  | 'favorites'
  | 'history'
  | 'settings'
  | 'album-detail'

interface SidebarProps {
  currentView: ViewType
  onViewChange: (view: ViewType) => void
}

export const Sidebar: React.FC<SidebarProps> = ({ currentView, onViewChange }) => {
  const { user, logout } = useAuthStore()

  const mainNav = [
    { id: 'albums' as ViewType, label: 'Albums', icon: Disc },
    { id: 'artists' as ViewType, label: 'Artists', icon: UserIcon },
    { id: 'tracks' as ViewType, label: 'Tracks', icon: Music },
  ]

  const libraryNav = [
    { id: 'playlists' as ViewType, label: 'Playlists', icon: ListMusic },
    { id: 'favorites' as ViewType, label: 'Favorites', icon: Heart },
    { id: 'history' as ViewType, label: 'History', icon: History },
  ]

  return (
    <aside className="w-64 bg-zinc-950/80 border-r border-zinc-800/60 flex flex-col justify-between p-4 shrink-0 backdrop-blur-md">
      <div className="overflow-y-auto">
        {/* Brand Logo */}
        <div className="flex items-center gap-3 px-3 py-4 mb-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-amber-500 to-rose-500 flex items-center justify-center shadow-lg shadow-rose-500/20">
            <span className="font-black text-xl text-black tracking-tight font-mono">nE</span>
          </div>
          <div>
            <h1 className="font-bold text-lg leading-none tracking-tight text-white flex items-center gap-1.5">
              nE Audio <Sparkles className="w-3.5 h-3.5 text-amber-400" />
            </h1>
            <span className="text-xs text-zinc-400 font-medium">Personal Music Server</span>
          </div>
        </div>

        {/* Discovery Menu */}
        <div className="mb-6">
          <p className="px-3 text-[11px] font-bold text-zinc-500 uppercase tracking-wider mb-2">
            Catalog
          </p>
          <nav className="space-y-1">
            {mainNav.map((item) => {
              const Icon = item.icon
              const isActive = currentView === item.id || (currentView === 'album-detail' && item.id === 'albums')
              return (
                <button
                  key={item.id}
                  onClick={() => onViewChange(item.id)}
                  className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all ${
                    isActive
                      ? 'bg-zinc-800 text-white shadow-sm font-semibold'
                      : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isActive ? 'text-rose-400' : 'text-zinc-400'}`} />
                  {item.label}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Personal Library Section */}
        <div className="mb-6">
          <p className="px-3 text-[11px] font-bold text-zinc-500 uppercase tracking-wider mb-2">
            My Collection
          </p>
          <nav className="space-y-1">
            {libraryNav.map((item) => {
              const Icon = item.icon
              const isActive = currentView === item.id
              return (
                <button
                  key={item.id}
                  onClick={() => onViewChange(item.id)}
                  className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all ${
                    isActive
                      ? 'bg-zinc-800 text-white shadow-sm font-semibold'
                      : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isActive ? 'text-rose-400' : 'text-zinc-400'}`} />
                  {item.label}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Settings */}
        <div>
          <nav className="space-y-1">
            <button
              onClick={() => onViewChange('settings')}
              className={`w-full flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all ${
                currentView === 'settings'
                  ? 'bg-zinc-800 text-white shadow-sm font-semibold'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
              }`}
            >
              <Settings className={`w-4 h-4 ${currentView === 'settings' ? 'text-rose-400' : 'text-zinc-400'}`} />
              Settings
            </button>
          </nav>
        </div>
      </div>

      {/* User Session Footer */}
      <div className="border-t border-zinc-800/60 pt-4 px-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="w-8 h-8 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-xs font-semibold text-zinc-300">
              {user?.username.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-white truncate">{user?.username}</p>
              <p className="text-xs text-zinc-400 truncate">{user?.isAdmin ? 'Administrator' : 'Listener'}</p>
            </div>
          </div>
          <button
            onClick={() => logout()}
            title="Sign out"
            className="p-1.5 text-zinc-400 hover:text-rose-400 hover:bg-zinc-800/80 rounded-md transition-colors"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>
  )
}
