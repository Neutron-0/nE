# 🎵 nE — Autonomous Personal Music Server & Streaming Platform

**nE** is a modern, lightweight, high-performance self-hosted personal music streaming server and web player designed for audiophiles. Built with a unified **Go backend** and an embedded **React 19 single-page web player**, nE provides zero-dependency deployment, real-time audio transcoding, synchronized karaoke lyrics, dynamic smart mixes, and native mobile streaming support via the OpenSubsonic protocol.

---

## ✨ Features

### 🎛️ 3-Tier Master Audio Engine (Normal, Spotify, Apple Studio)
nE includes a built-in 32-bit floating-point Web Audio DSP mastering pipeline featuring three distinct switchable engines:
* **Normal (Bit-Perfect Direct)**: 100% clean, uncolored, bit-perfect pass-through with flat frequency response for external DAC purists.
* **Spotify Master Engine**: Calibrated to EBU R128 **-14 dB LUFS** with a 10-band acoustic profile, sub-bass transient punch, and master bus soft-knee dynamic compression.
* **Apple Studio Hyperion Engine**: Our flagship audiophile processing chain engineered to surpass Apple Music's playback fidelity:
  * **24-Bit Dynamic Headroom (-16 dB LUFS)**: Preserves dynamic breathing room and natural transient attack.
  * **Aural Harmonic Synthesizer**: Uses non-linear polynomial wave-shaping ($f(x) = 1.5x - 0.5x^3$) to synthesize musical 2nd/3rd order overtones into the **16 kHz – 22 kHz** band, restoring the "air" and breath stripped by lossy compression.
  * **Sub-Harmonic Bass Extender**: Tightens fundamental sub-frequencies (< 75 Hz) without boomy mid-bass mud.
  * **Binaural Spatial Soundstage Matrix**: Employs psychoacoustic Haas micro-delays (12ms) to project instruments outside your headphones into an open 3D soundfield.
  * **-1.0 dBTP True-Peak Limiter**: Eliminates digital Inter-Sample Peaks (ISPs) and analog DAC clipping distortion.
* 📖 **[Read the Full Audio Engine Technical Architecture & Math Guide](docs/AUDIO_ENGINES.md)**

### 🎧 Playback & Streaming
* **Modern Web Player**: Built with React 19, Tailwind CSS, Lucide icons, and HTML5 Audio. Supports full queue control, shuffle, repeat, volume adjustment, and quick seeking.
* **RFC 7233 Byte-Range Streaming**: Ultra-low latency seeking (`206 Partial Content`) for instant scrubbing across large audio files.
* **On-the-Fly FFmpeg Transcoding**: Stream lossless FLAC or ALAC files seamlessly over cellular connections with automatic conversion to MP3/Opus with concurrency semaphores to protect low-RAM server instances.

### 🎤 Synchronized Real-Time Lyrics (Apple Music / Spotify Style)
* **Karaoke-Style Display**: Full-screen blurred artwork overlay with real-time timed lyrics that highlight and scroll automatically with song playback.
* **Dual-Source Auto-Fetching**:
  1. Checks for local `.lrc` files in your album directories (e.g. `Track 01.lrc`).
  2. Automatically falls back to the free, public [LRCLIB](https://lrclib.net/) database and caches lyrics locally.
* **Interactive Seeking**: Click any lyric line to jump directly to that timestamp in the song.

### ⚡ Dynamic Smart Playlists
* **Automated Music Mixes**:
  * **Recently Added**: Newest tracks ingested into your library.
  * **Most Played**: Your personal heavy-rotation songs.
  * **Recently Played**: Listening history.
  * **Top Rated**: Tracks rated 4 or 5 stars.
  * **Random Mix**: Instant shuffle mix across your entire collection.
* **Custom Playlists**: Create, rename, reorder, and curate custom playlists with continuous position sequencing.

### 📖 Artist Biographies & Editorial Metadata
* Automatically fetches artist background stories, summaries, and high-resolution photo portraits via Wikipedia / MusicBrainz without requiring paid third-party API keys.

### 📱 OpenSubsonic Mobile App Compatibility (`/rest/*`)
Connect your mobile device and stream on iOS, Android, or desktop using standard Subsonic client apps:
* **iOS**: [Amperfy](https://apps.apple.com/app/amperfy/id1500600109), [play:Sub](https://apps.apple.com/app/play-sub/id685326509), [Substreamer](https://apps.apple.com/app/substreamer/id1012991586)
* **Android**: [Symfonium](https://play.google.com/store/apps/details?id=app.symfonik.music.player), [DSub](https://play.google.com/store/apps/details?id=github.daneren200.dsub), [Substreamer](https://play.google.com/store/apps/details?id=com.substreamer)
* **CarPlay & Android Auto**: Fully supported through compatible Subsonic mobile apps.

### 🛡️ Resilient Architecture & Library Ingestion
* **11-Stage Ingestion Pipeline**: Extracts metadata (ID3v1, ID3v2, Vorbis Comments, MP4 tags) for FLAC, MP3, M4A, OGG, Opus, and WAV.
* **5-Tier Persistent ID (PID) Resolution**: If you rename files or reorganize album folders, your play history, favorites, and playlists remain completely intact without data loss.
* **Storage-Offline Safety**: If a USB drive or cloud volume disconnects, the scanner gracefully aborts without mass-pruning your tracks.
* **Hot SQLite Online Backup**: Atomic, non-blocking snapshots using SQLite's native `VACUUM INTO` engine.
* **High-Speed FTS5 Search**: Sub-millisecond full-text search across titles, artists, albums, and genres.

---

## 🚀 Quickstart (Local)

### Prerequisites
* **Go 1.24+**
* **FFmpeg** (in your system PATH)
* **Node.js 18+** (for building the frontend)

### 1. Build the Frontend
```bash
cd web
npm install
npm run build
cd ..
```

### 2. Run the Server
```bash
go run ./cmd/ne
```

Open your browser at `http://localhost:4533`.

---

## 🐳 Docker Deployment

You can run nE with Docker in a single command:

```bash
docker build -t ne:latest .
docker run -d \
  -p 4533:4533 \
  -v /path/to/your/music:/music:ro \
  -v ne-data:/data \
  --name ne-server \
  ne:latest
```

---

## ☁️ 100% Free Cloud Deployment on Render

### Step 1: Push Repository to GitHub
Ensure your repository is pushed to your private GitHub account.

### Step 2: Create Web Service on Render
1. Open [dashboard.render.com](https://dashboard.render.com).
2. Click **New +** $\rightarrow$ **Web Service**.
3. Connect your private repository (`Neutron-0/nE`).
4. Settings:
   * **Runtime**: `Docker`
   * **Plan**: `Free`
   * **Port**: `4533`
5. Click **Create Web Service**.

### Step 3: Prevent Render From Sleeping (24/7 Keep-Alive)
Render free web services sleep after 15 minutes of inactivity. To keep your server awake 24/7 for free:
1. Go to [UptimeRobot.com](https://uptimerobot.com) (Free).
2. Click **Add New Monitor**:
   * **Type**: `HTTP(s)`
   * **Name**: `nE Music Server`
   * **URL**: `https://<your-render-url>.onrender.com/api/v1/health/liveness`
   * **Interval**: `Every 5 minutes`
3. Click **Create Monitor**. Your server will stay active 24/7!

---

## 📱 Mobile App Connection Guide (Subsonic)

To listen to your music on iPhone or Android:
1. Download **Symfonium** (Android), **Amperfy** (iOS), or **Substreamer** (iOS/Android).
2. Choose **Add Server** $\rightarrow$ Select **Subsonic / OpenSubsonic**.
3. Fill in your server details:
   * **Server URL**: `https://<your-render-subdomain>.onrender.com`
   * **Username**: `neo`
   * **Password**: `neo03`
4. Click **Connect** and enjoy streaming with offline caching and CarPlay/Android Auto support!

---

## ⚙️ Environment Variables

| Variable | Default | Description |
|---|---|---|
| `NE_PORT` | `4533` | HTTP listening port (falls back to `$PORT` on cloud providers) |
| `NE_HOST` | `0.0.0.0` | Host IP address to bind |
| `NE_MUSIC_DIR` | `/music` | Default path for the music folder |
| `NE_DATA_DIR` | `/data` | Path for SQLite database storage |
| `NE_CACHE_DIR` | `/cache` | Path for artwork and transcode caches |
| `NE_CONFIG_DIR` | `/config` | Path for persistent cryptographic keys |
| `NE_JWT_SECRET` | *(auto-generated)* | 32-byte secret for JWT signing |
| `NE_STREAMING_MAX_CONCURRENT_TRANSCODES` | `1` | Maximum parallel FFmpeg transcode sessions |

---

## 📜 CLI Diagnostic Utilities

nE includes built-in maintenance tools:

```bash
# Check SQLite database integrity and foreign key constraints
ne check-db

# Rebuild SQLite FTS5 search index
ne rebuild-fts

# Recalculate duration and song counts across catalog
ne recalculate-stats

# Create an atomic hot backup of the database
ne backup ./backup.db
```

---

## 📄 License
MIT License. Built with ❤️ for independent, self-hosted music streaming.
