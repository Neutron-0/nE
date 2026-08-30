import { useEffect } from 'react'
import { usePlayerStore } from '../store/playerStore'

export function useKeyboardShortcuts() {
  const { togglePlay, next, previous, seek, currentTime, duration, toggleMute } =
    usePlayerStore()

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable
      ) {
        return
      }

      switch (e.code) {
        case 'Space':
          e.preventDefault()
          togglePlay()
          break
        case 'ArrowRight':
          e.preventDefault()
          seek(Math.min(duration, currentTime + 5))
          break
        case 'ArrowLeft':
          e.preventDefault()
          seek(Math.max(0, currentTime - 5))
          break
        case 'KeyM':
          e.preventDefault()
          toggleMute()
          break
        case 'KeyJ':
          e.preventDefault()
          previous()
          break
        case 'KeyK':
          e.preventDefault()
          next()
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [togglePlay, next, previous, seek, currentTime, duration, toggleMute])
}
