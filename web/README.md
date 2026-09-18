# nE Web Client

The official single-page web client for the nE Autonomous Personal Music Server.

Built with React 19, TypeScript, Tailwind CSS v4, Zustand, and the Web Audio API.

## Features

* **3-Tier Master Audio Engine**: Integrated Web Audio DSP mastering pipeline (Direct, Broadcast Master, and Hyperion Studio Pro).
* **Synchronized Lyrics**: Interactive karaoke-style LRC lyric presentation with line-by-line seeking.
* **Spectrum Telemetry HUD**: Real-time 256-bin Fast Fourier Transform (FFT) visualizer.
* **Smart & Custom Playlists**: Playlist curation, reordering, and dynamic mix playback.
* **Autonomous Token Refresh**: Transparent 401 handling with HTTP-only refresh tokens.
* **Keyboard Shortcuts**: Space (play/pause), arrow keys (seek/volume), M (mute), L (lyrics toggle).

## Development

```bash
# Install dependencies
npm install

# Start development server with Hot Module Replacement
npm run dev

# Build production bundle (embedded into the Go binary)
npm run build

# Lint source files
npm run lint
```

## Architecture

* `src/lib/audioEngine.ts`: Web Audio DSP signal chain, harmonic wave-shaping, and binaural cross-feed.
* `src/lib/api.ts`: Typed REST client with automatic token rotation.
* `src/store/playerStore.ts`: Global playback state, queue sequencing, and volume management.
* `src/components/`: Modular UI elements (NowPlayingPanel, AudioHUD, LyricsOverlay, PlayerBar).
* `src/views/`: Primary views (Tracks, Albums, Artists, Playlists, History, Settings).

