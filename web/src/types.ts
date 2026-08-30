export interface User {
  id: string
  username: string
  email?: string
  isAdmin: boolean
  canTranscode: boolean
  createdAt: string
}

export interface AuthResult {
  accessToken: string
  expiresAt: string
  user: User
}

export interface Artist {
  id: string
  name: string
  sortName: string
  biography?: string
  albumCount: number
  trackCount: number
  isStarred?: boolean
  rating?: number
}

export interface Album {
  id: string
  title: string
  sortTitle: string
  albumArtistId: string
  albumArtist: string
  year: number
  originalYear?: number
  releaseDate?: string
  discCount: number
  trackCount: number
  duration: number
  sizeBytes: number
  isCompilation: boolean
  tracks?: Track[]
  genres?: string[]
  isStarred?: boolean
  rating?: number
}

export interface Track {
  id: string
  pid: string
  libraryId: string
  path: string
  folderPath: string
  filename: string
  title: string
  sortTitle: string
  rawArtist: string
  albumId: string
  albumTitle?: string
  albumArtist?: string
  trackNumber: number
  discNumber: number
  discSubtitle?: string
  year: number
  duration: number
  bitRate: number
  sampleRate: number
  bitDepth?: number
  channels: number
  format: string
  codec: string
  fileSize: number
  rgTrackGain?: number
  rgTrackPeak?: number
  rgAlbumGain?: number
  rgAlbumPeak?: number
  hasEmbeddedCover: boolean
  isStarred?: boolean
  rating?: number
  playCount?: number
}

export interface Genre {
  id: string
  name: string
  trackCount: number
  albumCount: number
}

export interface Playlist {
  id: string
  name: string
  comment?: string
  ownerId: string
  isPublic: boolean
  trackCount: number
  duration: number
  tracks?: Track[]
  createdAt: string
  updatedAt: string
}

export interface PlaybackRecord {
  id: string
  userId: string
  trackId: string
  playerName?: string
  playedAt: string
  durationPlayed: number
  completed: boolean
  trackTitle?: string
  artistName?: string
  albumTitle?: string
}

export interface SearchResult {
  query: string
  artists: Artist[]
  albums: Album[]
  tracks: Track[]
}

export interface Library {
  id: string
  name: string
  path: string
  lastScannedAt?: string
  scanStatus: 'idle' | 'scanning' | 'failed'
  trackCount: number
  albumCount: number
  totalBytes: number
}

export interface CollectionResponse<T> {
  items: T[]
  nextCursor?: string
  total: number
}

export interface AdminStats {
  totalTracks: number
  totalAlbums: number
  totalArtists: number
  totalGenres: number
  totalPlaylists: number
  totalDuration: number
  totalBytes: number
}

export interface Diagnostics {
  status: string
  uptime_seconds: number
  version: string
  go_version: string
  platform: string
  num_cpu: number
  ffmpeg: {
    available: boolean
    path?: string
    version?: string
  }
  storage: {
    data_dir: string
    cache_dir: string
    config_dir: string
    music_dir: string
  }
}

export interface LyricLine {
  time: number
  text: string
}

export interface LyricsResult {
  trackId: string
  isSynced: boolean
  lines: LyricLine[]
  plainLyrics?: string
  source: string
}

export interface SmartPlaylistSummary {
  id: string
  name: string
  description: string
  icon: string
}

export interface ArtistBiography {
  artistId: string
  name: string
  biography: string
  imageUrl?: string
  source: string
}

export interface GitHubSyncStatus {
  configured: boolean
  repo: string
  branch: string
  autoSync: boolean
  lastSyncAt?: string
  lastError?: string
  tokenHint?: string
}
