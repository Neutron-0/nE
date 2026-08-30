import { create } from 'zustand'
import type { User } from '../types'
import { api } from '../lib/api'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  isNeedsSetup: boolean
  error: string | null

  checkAuth: () => Promise<void>
  login: (username: string, pass: string) => Promise<void>
  setup: (username: string, email: string, pass: string) => Promise<void>
  logout: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,
  isNeedsSetup: false,
  error: null,

  checkAuth: async () => {
    set({ isLoading: true, error: null })
    try {
      const user = await api.getMe()
      set({ user, isAuthenticated: true, isLoading: false, isNeedsSetup: false })
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false })
    }
  },

  login: async (username, password) => {
    set({ isLoading: true, error: null })
    try {
      const res = await api.login(username, password)
      set({ user: res.user, isAuthenticated: true, isLoading: false, isNeedsSetup: false })
    } catch (err: any) {
      set({ error: err.message || 'Login failed', isLoading: false })
      throw err
    }
  },

  setup: async (username, email, password) => {
    set({ isLoading: true, error: null })
    try {
      const res = await api.setup(username, email, password)
      set({ user: res.user, isAuthenticated: true, isLoading: false, isNeedsSetup: false })
    } catch (err: any) {
      set({ error: err.message || 'Setup failed', isLoading: false })
      throw err
    }
  },

  logout: async () => {
    set({ isLoading: true })
    await api.logout()
    set({ user: null, isAuthenticated: false, isLoading: false })
  },

  clearError: () => set({ error: null }),
}))
