import React, { useState } from 'react'
import { Lock, User, AlertCircle } from 'lucide-react'
import { motion } from 'motion/react'
import { useAuthStore } from '../store/authStore'

interface LoginViewProps {
  onGoToSetup?: () => void
}

export const LoginView: React.FC<LoginViewProps> = ({ onGoToSetup }) => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const { login, isLoading, error, clearError } = useAuthStore()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) return
    try {
      await login(username, password)
    } catch {
      // handled in store
    }
  }

  return (
    <div className="min-h-screen bg-[#121110] flex flex-col items-center justify-center p-6 text-zinc-200 select-none">
      {/* Subtle ambient lighting */}
      <div className="absolute w-[500px] h-[500px] bg-rose-600/5 rounded-full blur-3xl pointer-events-none" />

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="w-full max-w-md bg-[#191716] border border-[#2d2824] rounded-[28px] p-8 shadow-2xl relative z-10"
      >
        {/* Monospace telemetry header matching Image 1 */}
        <div className="flex items-center justify-between text-[11px] font-hud tracking-widest text-zinc-400 border-b border-white/5 pb-4 mb-6">
          <span>SECURE_AUTH // W-01</span>
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse" />
            <span className="text-zinc-400">READY</span>
          </div>
        </div>

        {/* Brand */}
        <div className="flex flex-col items-center text-center mb-8">
          <div className="w-14 h-14 rounded-2xl bg-[#23201d] border border-white/10 flex items-center justify-center shadow-2xl mb-3">
            <span className="font-black text-2xl text-white font-hud tracking-tighter">nE</span>
          </div>
          <h2 className="text-2xl font-black text-white tracking-tight">nE Audio</h2>
          <p className="text-xs text-zinc-400 mt-1">Autonomous Personal Music Server</p>
        </div>

        {error && (
          <div className="mb-5 p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-xs text-rose-400 font-medium">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
              Username
            </label>
            <div className="relative">
              <User className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
              <input
                type="text"
                required
                value={username}
                onChange={(e) => {
                  clearError()
                  setUsername(e.target.value)
                }}
                placeholder="neo"
                className="w-full bg-[#131211] border border-[#2d2824] rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-400 focus:outline-none focus:border-white/30 transition-colors font-sans shadow-inner"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-wider mb-1.5 font-hud">
              Password
            </label>
            <div className="relative">
              <Lock className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
              <input
                type="password"
                required
                value={password}
                onChange={(e) => {
                  clearError()
                  setPassword(e.target.value)
                }}
                placeholder="••••••••"
                className="w-full bg-[#131211] border border-[#2d2824] rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-400 focus:outline-none focus:border-white/30 transition-colors font-sans shadow-inner"
              />
            </div>
          </div>

          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            type="submit"
            disabled={isLoading}
            className="w-full mt-2 py-3 px-4 rounded-xl bg-white text-black font-bold text-sm shadow-xl hover:bg-zinc-100 transition-colors disabled:opacity-50 cursor-pointer"
          >
            {isLoading ? 'Establishing Link...' : 'Enter Session'}
          </motion.button>
        </form>

        {onGoToSetup && (
          <div className="mt-6 text-center pt-4 border-t border-white/5">
            <button
              onClick={onGoToSetup}
              className="text-xs text-zinc-400 hover:text-white transition-colors cursor-pointer"
            >
              First time setup? Create admin account →
            </button>
          </div>
        )}
      </motion.div>
    </div>
  )
}
