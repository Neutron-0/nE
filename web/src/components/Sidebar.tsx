import React from 'react'
import { motion } from 'motion/react'
import {
  Home,
  Compass,
  Library,
  Settings,
  LogOut,
  Heart,
  History,
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

  // Primary navigation matching Image 2
  const menuItems = [
    { id: 'albums' as ViewType, label: 'Home', icon: Home },
    { id: 'tracks' as ViewType, label: 'Explore', icon: Compass },
    { id: 'artists' as ViewType, label: 'Library', icon: Library },
  ]

  // Playlists matching Image 2
  const defaultPlaylists = [
    { name: 'After Hours Mix', id: 'playlists' },
    { name: 'Productivity Flows', id: 'playlists' },
    { name: 'Late Night Jazz', id: 'playlists' },
    { name: 'Discovery Daily', id: 'playlists' },
  ]

  return (
    <aside className="w-56 md:w-60 bg-[#191716]/95 border-r border-[#302c28]/80 p-5 flex flex-col justify-between shrink-0 select-none h-full">
      <div className="overflow-y-auto">
        {/* Brand Logo matching nE */}
        <div className="flex items-center gap-3 px-2 py-3 mb-6">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-amber-500 to-rose-500 flex items-center justify-center shadow-lg shadow-rose-500/20">
            <span className="font-black text-lg text-black font-hud">nE</span>
          </div>
          <div>
            <h1 className="font-extrabold text-base tracking-tight text-white leading-none">
              nE Audio
            </h1>
            <span className="text-[11px] text-zinc-400 font-medium">Personal Streamer</span>
          </div>
        </div>

        {/* Section 1: MENU matching Image 2 */}
        <div className="mb-8">
          <p className="px-3 text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-2 font-hud">
            MENU
          </p>
          <nav className="space-y-1">
            {menuItems.map((item) => {
              const Icon = item.icon
              const isActive =
                currentView === item.id || (item.id === 'albums' && currentView === 'album-detail')

              return (
                <button
                  key={item.id}
                  onClick={() => onViewChange(item.id)}
                  className={`relative w-full flex items-center gap-3.5 px-3.5 py-2.5 rounded-xl text-sm font-semibold transition-colors cursor-pointer ${
                    isActive ? 'text-white' : 'text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  {/* Sliding active pill indicator (Image 2 signature) */}
                  {isActive && (
                    <motion.div
                      layoutId="active-pill"
                      className="absolute inset-0 bg-[#2d2824] rounded-xl border border-white/5 shadow-sm -z-0"
                      transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                    />
                  )}
                  <Icon className={`w-4 h-4 relative z-10 ${isActive ? 'text-white' : 'text-zinc-400'}`} />
                  <span className="relative z-10">{item.label}</span>
                </button>
              )
            })}
          </nav>
        </div>

        {/* Section 2: YOUR PLAYLISTS matching Image 2 */}
        <div className="mb-6">
          <p className="px-3 text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-2 font-hud">
            YOUR PLAYLISTS
          </p>
          <nav className="space-y-1">
            {defaultPlaylists.map((pl, idx) => (
              <button
                key={idx}
                onClick={() => onViewChange('playlists')}
                className={`w-full text-left px-3.5 py-2 rounded-xl text-sm transition-colors truncate cursor-pointer ${
                  currentView === 'playlists' && idx === 0
                    ? 'text-zinc-200 font-semibold'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#25221f]/50'
                }`}
              >
                {pl.name}
              </button>
            ))}
          </nav>
        </div>

        {/* Favorites & History shortcuts */}
        <div className="pt-2 border-t border-white/5">
          <button
            onClick={() => onViewChange('favorites')}
            className={`w-full flex items-center gap-3 px-3.5 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer ${
              currentView === 'favorites' ? 'text-rose-400 font-semibold' : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Heart className="w-3.5 h-3.5" />
            <span>Favorites</span>
          </button>
          <button
            onClick={() => onViewChange('history')}
            className={`w-full flex items-center gap-3 px-3.5 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer ${
              currentView === 'history' ? 'text-white font-semibold' : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <History className="w-3.5 h-3.5" />
            <span>History</span>
          </button>
        </div>
      </div>

      {/* User Session Footer */}
      <div className="pt-4 border-t border-[#302c28]/80 px-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="w-8 h-8 rounded-full bg-[#292522] border border-white/10 flex items-center justify-center text-xs font-bold text-zinc-300">
              {user?.username.charAt(0).toUpperCase() || 'N'}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-white truncate">{user?.username || 'neo'}</p>
              <p className="text-[11px] text-zinc-400 truncate">
                {user?.isAdmin ? 'Administrator' : 'Listener'}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => onViewChange('settings')}
              title="Settings"
              className="p-1.5 text-zinc-400 hover:text-white rounded-lg hover:bg-white/5 transition-colors cursor-pointer"
            >
              <Settings className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => logout()}
              title="Sign out"
              className="p-1.5 text-zinc-400 hover:text-rose-400 rounded-lg hover:bg-white/5 transition-colors cursor-pointer"
            >
              <LogOut className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </aside>
  )
}
