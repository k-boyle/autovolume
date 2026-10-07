import type { ReactNode } from 'react'
import { NowPlayingContext } from '../hooks/useNowPlaying'
import { useNowPlayingStream } from '../hooks/useNowPlayingStream'

export function NowPlayingProvider({ children }: { children: ReactNode }) {
  const value = useNowPlayingStream()
  return <NowPlayingContext value={value}>{children}</NowPlayingContext>
}
