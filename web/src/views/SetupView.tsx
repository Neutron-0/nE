import React, { useState } from 'react'
import { Lock, User, Mail, AlertCircle, ShieldCheck } from 'lucide-react'
import { motion } from 'motion/react'
import { useAuthStore } from '../store/authStore'

interface SetupViewProps {
  onBackToLogin?: () => void
}

export const SetupView: React.FC<SetupViewProps> = ({ onBackToLogin }) => {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [clientError, setClientError] = useState<string | null>(null)
  const { setup, isLoading, error, clearError } = useAuthStore()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setClientError(null)

    if (password !== confirmPassword) {
      setClientError('Passwords do not match')
      return
    }
    if (password.length < 6) {
      setClientError('Password must be at least 6 characters')
      return
    }

    try {
      await setup(username, email, password)
    } catch {
      // handled in store
    }
  }

  const activeError = clientError || error

  return (
    <div className="min-h-screen bg-black flex flex-col items-center justify-center p-6 text-zinc-200 select-none relative overflow-hidden">
      {/* Ambient Lighting */}
      <div className="absolute w-[600px] h-[600px] bg-white/[0.02] rounded-full blur-3xl pointer-events-none" />

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="w-full max-w-md liquid-glass rounded-[28px] p-8 shadow-2xl relative z-10"
      >
        {/* Monospace telemetry header */}
        <div className="flex items-center justify-between text-[11px] font-hud tracking-widest text-zinc-500 border-b border-white/[0.06] pb-4 mb-6">
          <span>SETUP_INIT // PROTOCOL_01</span>
          <div className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-white animate-pulse" />
            <span className="text-zinc-400">UNINITIALIZED</span>
          </div>
        </div>

        {/* Brand */}
        <div className="flex flex-col items-center text-center mb-8">
          <div className="w-14 h-14 rounded-2xl liquid-glass flex items-center justify-center shadow-2xl mb-3">
            <ShieldCheck className="w-7 h-7 text-white" />
          </div>
          <h2 className="text-2xl font-black text-white tracking-tight">System Provisioning</h2>
          <p className="text-xs text-zinc-500 mt-1 font-hud">SOLID BLACK • LIQUID GLASS</p>
        </div>

        {activeError && (
          <div className="mb-5 p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center gap-3 text-xs text-rose-400 font-medium">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{activeError}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-widest mb-1.5 font-hud">
              Admin Username
            </label>
            <div className="relative">
              <User className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                required
                value={username}
                onChange={(e) => {
                  clearError()
                  setClientError(null)
                  setUsername(e.target.value)
                }}
                placeholder="neo"
                className="w-full liquid-glass-input rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-widest mb-1.5 font-hud">
              Email (Optional)
            </label>
            <div className="relative">
              <Mail className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="neo@example.com"
                className="w-full liquid-glass-input rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-widest mb-1.5 font-hud">
              Root Password
            </label>
            <div className="relative">
              <Lock className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="password"
                required
                value={password}
                onChange={(e) => {
                  clearError()
                  setClientError(null)
                  setPassword(e.target.value)
                }}
                placeholder="••••••••"
                className="w-full liquid-glass-input rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-widest mb-1.5 font-hud">
              Confirm Password
            </label>
            <div className="relative">
              <Lock className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="password"
                required
                value={confirmPassword}
                onChange={(e) => {
                  setClientError(null)
                  setConfirmPassword(e.target.value)
                }}
                placeholder="••••••••"
                className="w-full liquid-glass-input rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-zinc-600 focus:outline-none"
              />
            </div>
          </div>

          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            type="submit"
            disabled={isLoading}
            className="w-full mt-2 py-3 px-4 rounded-xl bg-white text-black font-bold text-sm shadow-2xl hover:bg-zinc-200 transition-colors disabled:opacity-50 cursor-pointer"
          >
            {isLoading ? 'Configuring System...' : 'Initialize Server & Launch'}
          </motion.button>
        </form>

        {onBackToLogin && (
          <div className="mt-6 text-center pt-4 border-t border-white/[0.06]">
            <button
              onClick={onBackToLogin}
              className="text-xs text-zinc-500 hover:text-white transition-colors cursor-pointer"
            >
              ← Already configured? Back to sign in
            </button>
          </div>
        )}
      </motion.div>
    </div>
  )
}
