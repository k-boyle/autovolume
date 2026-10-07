import { createContext, useContext } from 'react'
import type { NowPlaying } from './useNowPlayingStream'

export const NowPlayingContext = createContext<NowPlaying | null>(null)

export function useNowPlaying(): NowPlaying {
  const value = useContext(NowPlayingContext)
  if (!value) throw new Error('useNowPlaying must be used inside <NowPlayingProvider>')
  return value
}
