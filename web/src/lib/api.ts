import type {
  AuthResult,
  CollectionResponse,
  User,
  Artist,
  Album,
  Track,
  Library,
  Diagnostics,
  Playlist,
  PlaybackRecord,
  SearchResult,
  AdminStats,
  LyricsResult,
  SmartPlaylistSummary,
  ArtistBiography,
} from '../types'

const BASE_URL = '/api/v1'

class ApiClient {
  private token: string | null = null

  constructor() {
    this.token = localStorage.getItem('ne_token')
  }

  setToken(token: string | null) {
    this.token = token
    if (token) {
      localStorage.setItem('ne_token', token)
    } else {
      localStorage.removeItem('ne_token')
    }
  }

  getToken(): string | null {
    return this.token
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers = new Headers(options.headers || {})
    headers.set('Content-Type', 'application/json')
    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    const response = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      if (response.status === 401) {
        this.setToken(null)
      }

      let errorMessage = `HTTP Error ${response.status}`
      try {
        const errorData = await response.json()
        if (errorData?.error?.message) {
          errorMessage = errorData.error.message
        }
      } catch {
        // use default error message
      }
      throw new Error(errorMessage)
    }

    return response.json()
  }

  // Auth Endpoints
  async setup(username: string, email: string, password: string): Promise<AuthResult> {
    const res = await this.request<AuthResult>('/auth/setup', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    })
    this.setToken(res.accessToken)
    return res
  }

  async login(username: string, password: string): Promise<AuthResult> {
    const res = await this.request<AuthResult>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    this.setToken(res.accessToken)
    return res
  }

  async logout(): Promise<void> {
    try {
      await this.request('/auth/logout', { method: 'POST' })
    } finally {
      this.setToken(null)
    }
  }

  async getMe(): Promise<User> {
    if (!this.token) {
      throw new Error('No authentication token')
    }
    return this.request<User>('/auth/me')
  }

  // Catalog Endpoints
  async getArtists(limit = 100, offset = 0): Promise<CollectionResponse<Artist>> {
    return this.request<CollectionResponse<Artist>>(`/artists?limit=${limit}&offset=${offset}`)
  }

  async getArtist(id: string): Promise<Artist> {
    return this.request<Artist>(`/artists/${id}`)
  }

  async getAlbums(limit = 100, offset = 0): Promise<CollectionResponse<Album>> {
    return this.request<CollectionResponse<Album>>(`/albums?limit=${limit}&offset=${offset}`)
  }

  async getAlbum(id: string): Promise<Album> {
    return this.request<Album>(`/albums/${id}`)
  }

  async getTracks(limit = 100, offset = 0): Promise<CollectionResponse<Track>> {
    return this.request<CollectionResponse<Track>>(`/tracks?limit=${limit}&offset=${offset}`)
  }

  async getTrack(id: string): Promise<Track> {
    return this.request<Track>(`/tracks/${id}`)
  }

  async search(query: string, limit = 20): Promise<SearchResult> {
    return this.request<SearchResult>(`/search?q=${encodeURIComponent(query)}&limit=${limit}`)
  }

  // Annotations & Scrobble
  async starItem(itemType: 'track' | 'album' | 'artist', itemId: string, isStarred: boolean): Promise<void> {
    await this.request('/annotations/star', {
      method: 'POST',
      body: JSON.stringify({ itemType, itemId, isStarred }),
    })
  }

  async rateItem(itemType: 'track' | 'album', itemId: string, rating: number): Promise<void> {
    await this.request('/annotations/rating', {
      method: 'POST',
      body: JSON.stringify({ itemType, itemId, rating }),
    })
  }

  async scrobble(trackId: string, durationPlayed: number, completed: boolean): Promise<void> {
    try {
      await this.request('/playback/scrobble', {
        method: 'POST',
        body: JSON.stringify({ trackId, durationPlayed, completed, playerName: 'nE Web Player' }),
      })
    } catch (err) {
      console.warn('Scrobble failed:', err)
    }
  }

  async getFavorites(): Promise<CollectionResponse<Track>> {
    return this.request<CollectionResponse<Track>>('/favorites')
  }

  async getRecentHistory(limit = 50): Promise<CollectionResponse<PlaybackRecord>> {
    return this.request<CollectionResponse<PlaybackRecord>>(`/history/recent?limit=${limit}`)
  }

  // Playlists
  async getPlaylists(): Promise<CollectionResponse<Playlist>> {
    return this.request<CollectionResponse<Playlist>>('/playlists')
  }

  async getPlaylist(id: string): Promise<Playlist> {
    return this.request<Playlist>(`/playlists/${id}`)
  }

  async createPlaylist(name: string, comment = '', isPublic = false): Promise<Playlist> {
    return this.request<Playlist>('/playlists', {
      method: 'POST',
      body: JSON.stringify({ name, comment, isPublic }),
    })
  }

  async setPlaylistTracks(id: string, trackIds: string[]): Promise<void> {
    await this.request(`/playlists/${id}/tracks`, {
      method: 'PUT',
      body: JSON.stringify({ trackIds }),
    })
  }

  async deletePlaylist(id: string): Promise<void> {
    await this.request(`/playlists/${id}`, { method: 'DELETE' })
  }

  // Admin Endpoints
  async getLibraries(): Promise<CollectionResponse<Library>> {
    return this.request<CollectionResponse<Library>>('/admin/libraries')
  }

  async createLibrary(name: string, path: string): Promise<Library> {
    return this.request<Library>('/admin/libraries', {
      method: 'POST',
      body: JSON.stringify({ name, path }),
    })
  }

  async triggerScan(libraryId: string): Promise<{ status: string }> {
    return this.request('/admin/scan', {
      method: 'POST',
      body: JSON.stringify({ libraryId }),
    })
  }

  async getAdminStats(): Promise<AdminStats> {
    return this.request<AdminStats>('/admin/stats')
  }

  async getDiagnostics(): Promise<Diagnostics> {
    return this.request<Diagnostics>('/health/diagnostics')
  }

  // Lyrics
  async getLyrics(trackId: string): Promise<LyricsResult> {
    return this.request<LyricsResult>(`/lyrics/${trackId}`)
  }

  // Smart Playlists
  async getSmartPlaylists(): Promise<SmartPlaylistSummary[]> {
    return this.request<SmartPlaylistSummary[]>('/smart-playlists')
  }

  async getSmartPlaylistTracks(id: string, limit = 50): Promise<CollectionResponse<Track>> {
    return this.request<CollectionResponse<Track>>(`/smart-playlists/${id}?limit=${limit}`)
  }

  // Artist Editorial Metadata
  async getArtistBiography(artistId: string): Promise<ArtistBiography> {
    return this.request<ArtistBiography>(`/artists/${artistId}/biography`)
  }

  // Direct Audio Upload
  async uploadAudio(files: FileList | File[]): Promise<{ success: boolean; message: string; files: string[]; libraryId: string }> {
    const formData = new FormData()
    Array.from(files).forEach((file) => {
      formData.append('files', file)
    })

    const headers = new Headers()
    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    const response = await fetch(`${BASE_URL}/upload`, {
      method: 'POST',
      headers,
      body: formData,
    })

    if (!response.ok) {
      let msg = `Upload failed (HTTP ${response.status})`
      try {
        const data = await response.json()
        if (data.message) msg = data.message
      } catch {}
      throw new Error(msg)
    }

    return response.json()
  }

  getStreamUrl(trackId: string): string {
    return `${BASE_URL}/stream/${trackId}?token=${this.token || ''}`
  }

  getArtworkUrl(type: 'album' | 'track' | 'artist', id: string, size = 300): string {
    return `${BASE_URL}/artwork/${type}/${id}?size=${size}&token=${this.token || ''}`
  }
}

export const api = new ApiClient()
