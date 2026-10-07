import { Code, ConnectError, createClient } from '@connectrpc/connect'
import { useTransport } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { allQueriesKey } from '../api/keys'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import {
  NowPlayingService,
  type WatchNowPlayingResponse,
} from '../gen/autovolume/v1/now_playing_pb'
import { SetupService } from '../gen/autovolume/v1/setup_pb'

export type StreamStatus = 'connecting' | 'live' | 'reconnecting'

export interface NowPlaying {
  snapshot?: WatchNowPlayingResponse
  status: StreamStatus
}

const maxBackoffMs = 15_000

export function useNowPlayingStream(enabled = true): NowPlaying {
  const transport = useTransport()
  const queryClient = useQueryClient()
  const [state, setState] = useState<NowPlaying>({ status: 'connecting' })

  useEffect(() => {
    if (!enabled) return
    const client = createClient(NowPlayingService, transport)
    const abort = new AbortController()
    const { signal } = abort

    void (async () => {
      let attempt = 0
      while (!signal.aborted) {
        try {
          for await (const snapshot of client.watchNowPlaying({}, { signal })) {
            attempt = 0
            setState({ snapshot, status: 'live' })
          }
        } catch (err) {
          if (signal.aborted) return
          const { code } = ConnectError.from(err)
          if (code === Code.Unauthenticated) {
            void queryClient.invalidateQueries({ queryKey: allQueriesKey(AuthService.method.getSession) })
          } else if (code === Code.FailedPrecondition) {
            void queryClient.invalidateQueries({ queryKey: allQueriesKey(SetupService.method.getSetupStatus) })
          }
        }
        if (signal.aborted) return
        setState((s) => ({ ...s, status: 'reconnecting' }))
        await sleep(Math.min(1000 * 2 ** attempt, maxBackoffMs), signal)
        attempt++
      }
    })()

    return () => abort.abort()
  }, [enabled, transport, queryClient])

  return state
}

function sleep(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    const timer = setTimeout(resolve, ms)
    signal.addEventListener('abort', () => {
      clearTimeout(timer)
      resolve()
    })
  })
}
