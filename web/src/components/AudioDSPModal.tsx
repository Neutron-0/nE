import React from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { X, Sliders, Volume2, Sparkles, Zap, Radio } from 'lucide-react'
import { usePlayerStore } from '../store/playerStore'
import { PRESETS, EQ_FREQUENCIES, type PresetName } from '../lib/audioEngine'

interface AudioDSPModalProps {
  isOpen: boolean
  onClose: () => void
}

export const AudioDSPModal: React.FC<AudioDSPModalProps> = ({ isOpen, onClose }) => {
  const {
    dspState,
    setDspEnabled,
    setDspPreset,
    setPreAmpGain,
    setEQBand,
    setCompressorEnabled,
  } = usePlayerStore()

  if (!isOpen) return null

  const formatFreq = (freq: number) => {
    if (freq >= 1000) return `${freq / 1000}k`
    return `${freq}`
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-2xl flex items-center justify-center p-4 select-none">
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 16 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 16 }}
          transition={{ duration: 0.25 }}
          className="w-full max-w-2xl liquid-glass rounded-[28px] p-6 md:p-8 shadow-2xl overflow-hidden relative border border-white/[0.12]"
        >
          {/* Header */}
          <div className="flex items-center justify-between border-b border-white/[0.08] pb-4 mb-6">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl liquid-glass flex items-center justify-center text-white">
                <Sliders className="w-5 h-5 text-rose-500" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-xl font-black text-white tracking-tight">
                    Studio DSP Master Engine
                  </h3>
                  <span className="text-[10px] font-hud font-bold px-2 py-0.5 rounded-full bg-rose-500/20 text-rose-400 border border-rose-500/30">
                    24-BIT MASTER
                  </span>
                </div>
                <p className="text-xs text-zinc-400 mt-0.5">
                  Spotify-grade loudness matching, dynamic range compression & 10-band equalizer
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              {/* Master Bypass Toggle */}
              <button
                onClick={() => setDspEnabled(!dspState.enabled)}
                className={`flex items-center gap-2 px-3.5 py-1.5 rounded-full text-xs font-bold transition-all cursor-pointer ${
                  dspState.enabled
                    ? 'bg-rose-500 text-white shadow-lg shadow-rose-500/30'
                    : 'liquid-glass text-zinc-400 hover:text-white'
                }`}
              >
                <Radio className={`w-3.5 h-3.5 ${dspState.enabled ? 'animate-pulse' : ''}`} />
                <span>{dspState.enabled ? 'DSP ACTIVE' : 'BYPASSED'}</span>
              </button>

              <button
                onClick={onClose}
                className="p-1.5 text-zinc-400 hover:text-white rounded-full liquid-glass-pill transition-colors cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="space-y-6 max-h-[70vh] overflow-y-auto pr-1">
            {/* Presets Row */}
            <div>
              <label className="block text-[11px] font-bold text-zinc-400 uppercase tracking-widest mb-2.5 font-hud">
                Acoustic Mastering Presets
              </label>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                {(Object.keys(PRESETS) as PresetName[]).map((key) => {
                  const p = PRESETS[key]
                  const isActive = dspState.preset === key && dspState.enabled
                  return (
                    <button
                      key={key}
                      onClick={() => {
                        if (!dspState.enabled) setDspEnabled(true)
                        setDspPreset(key)
                      }}
                      className={`p-3 rounded-2xl text-left transition-all cursor-pointer ${
                        isActive
                          ? 'liquid-glass-pill border-rose-500/50 bg-rose-500/10 text-white shadow-lg'
                          : 'liquid-glass-card hover:bg-white/[0.04] text-zinc-400'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className={`text-xs font-bold ${isActive ? 'text-white' : 'text-zinc-300'}`}>
                          {p.name}
                        </span>
                        {key === 'spotify' && <Sparkles className="w-3 h-3 text-emerald-400" />}
                        {key === 'flat' && <Zap className="w-3 h-3 text-zinc-400" />}
                      </div>
                      <p className="text-[10px] text-zinc-500 line-clamp-2 leading-snug">
                        {p.description}
                      </p>
                    </button>
                  )
                })}
              </div>
            </div>

            {/* 10-Band Graphic Equalizer */}
            <div className="liquid-glass rounded-2xl p-5 border border-white/[0.08]">
              <div className="flex items-center justify-between mb-4">
                <span className="text-[11px] font-bold text-white uppercase tracking-widest font-hud">
                  10-Band Precision Equalizer
                </span>
                <span className="text-[11px] font-hud text-zinc-400">
                  {dspState.enabled ? 'ACTIVE // REALTIME BIQUAD' : 'BYPASSED'}
                </span>
              </div>

              <div className="grid grid-cols-10 gap-2 items-end h-40 pt-2 pb-1">
                {EQ_FREQUENCIES.map((freq, idx) => {
                  const gain = dspState.eqBands[idx] || 0
                  return (
                    <div key={freq} className="flex flex-col items-center gap-2 h-full">
                      {/* Gain dB Readout */}
                      <span className="text-[9px] font-hud text-zinc-400">
                        {gain > 0 ? `+${gain.toFixed(0)}` : gain.toFixed(0)}
                      </span>

                      {/* Vertical Slider Track */}
                      <div className="relative w-2 flex-1 bg-white/[0.08] rounded-full overflow-hidden flex items-end">
                        <div
                          className="w-full bg-rose-500 rounded-full transition-all"
                          style={{
                            height: `${((gain + 12) / 24) * 100}%`,
                          }}
                        />
                      </div>

                      {/* Input Range */}
                      <input
                        type="range"
                        min="-12"
                        max="12"
                        step="0.5"
                        value={gain}
                        disabled={!dspState.enabled}
                        onChange={(e) => setEQBand(idx, parseFloat(e.target.value))}
                        className="w-12 h-1 accent-rose-500 -rotate-90 origin-center absolute opacity-0 cursor-pointer"
                      />

                      {/* Frequency Label */}
                      <span className="text-[9px] font-hud text-zinc-500 mt-1">
                        {formatFreq(freq)}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>

            {/* Pre-Amp Loudness & Studio Dynamics Controls */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {/* Pre-Amp Loudness Stage */}
              <div className="liquid-glass rounded-2xl p-4 border border-white/[0.08] flex flex-col justify-between">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Volume2 className="w-4 h-4 text-rose-500" />
                    <span className="text-xs font-bold text-white">Loudness Pre-Amp</span>
                  </div>
                  <span className="text-xs font-hud font-bold text-white">
                    {dspState.preAmpGain > 0 ? `+${dspState.preAmpGain.toFixed(1)} dB` : `${dspState.preAmpGain.toFixed(1)} dB`}
                  </span>
                </div>
                <p className="text-[11px] text-zinc-400 mb-3">
                  Calibrated gain boost to achieve Spotify's mastered -14 dB LUFS standard.
                </p>
                <input
                  type="range"
                  min="-6"
                  max="6"
                  step="0.2"
                  disabled={!dspState.enabled}
                  value={dspState.preAmpGain}
                  onChange={(e) => setPreAmpGain(parseFloat(e.target.value))}
                  className="w-full accent-rose-500 cursor-pointer"
                />
              </div>

              {/* Studio Compressor & Limiter */}
              <div className="liquid-glass rounded-2xl p-4 border border-white/[0.08] flex flex-col justify-between">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Sliders className="w-4 h-4 text-rose-500" />
                    <span className="text-xs font-bold text-white">Dynamic Compressor</span>
                  </div>
                  <button
                    onClick={() => setCompressorEnabled(!dspState.compressorEnabled)}
                    disabled={!dspState.enabled}
                    className={`text-[10px] font-bold px-2 py-0.5 rounded cursor-pointer ${
                      dspState.compressorEnabled && dspState.enabled
                        ? 'bg-rose-500 text-white'
                        : 'bg-white/10 text-zinc-400'
                    }`}
                  >
                    {dspState.compressorEnabled && dspState.enabled ? 'ON' : 'OFF'}
                  </button>
                </div>
                <p className="text-[11px] text-zinc-400 mb-2">
                  Soft-knee master bus glue. Adds punch to kick drums & brings vocals forward.
                </p>
                <div className="grid grid-cols-3 gap-2 text-[10px] font-hud text-zinc-400 pt-1 border-t border-white/[0.06]">
                  <div>Ratio: 3.5:1</div>
                  <div>Attack: 3ms</div>
                  <div>Knee: 12dB</div>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  )
}
