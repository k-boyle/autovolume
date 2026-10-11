import { Code, createClient } from '@connectrpc/connect'
import { useTransport } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router'
import { errorCode, errorMessage } from '../api/errors'
import { allQueriesKey } from '../api/keys'
import { AuthService } from '../gen/autovolume/v1/auth_pb'

const pollMs = 1000
const giveUpAfterMs = 3 * 60_000

export type SignInPhase =
  | { step: 'idle' }
  | { step: 'starting' }
  | { step: 'waiting'; blockedAuthUrl?: string }
  | { step: 'failed'; message: string }

interface Attempt {
  abort: AbortController
  popup: Window | null
}

export function usePlexSignIn() {
  const transport = useTransport()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [phase, setPhase] = useState<SignInPhase>({ step: 'idle' })
  const current = useRef<Attempt | null>(null)

  useEffect(() => () => endAttempt(current), [])

  async function start() {
    endAttempt(current)
    const attempt: Attempt = { abort: new AbortController(), popup: openPlexPopup() }
    current.current = attempt
    const { signal } = attempt.abort
    const { popup } = attempt
    const client = createClient(AuthService, transport)

    const finishWith = (next: SignInPhase) => {
      endAttempt(current)
      setPhase(next)
    }

    setPhase({ step: 'starting' })
    try {
      const { authUrl, loginId } = await client.startLogin({}, { signal })
      if (popup?.closed) return finishWith({ step: 'idle' })
      if (popup) popup.location.href = authUrl
      setPhase({ step: 'waiting', blockedAuthUrl: popup ? undefined : authUrl })

      const deadline = Date.now() + giveUpAfterMs
      let popupClosedBeforeLastPoll = false
      for (;;) {
        const { complete } = await client.completeLogin({ loginId }, { signal })
        if (complete) {
          endAttempt(current)
          queryClient.removeQueries({ queryKey: allQueriesKey(AuthService.method.getSession) })
          return void navigate('/', { replace: true })
        }
        if (popupClosedBeforeLastPoll) return finishWith({ step: 'idle' })
        popupClosedBeforeLastPoll = popup?.closed ?? false
        if (Date.now() > deadline) return finishWith({ step: 'failed', message: 'Sign-in took too long. Try again.' })
        await new Promise((resolve) => setTimeout(resolve, pollMs))
        if (signal.aborted) return
      }
    } catch (err) {
      if (signal.aborted) return
      const message = errorCode(err) === Code.NotFound ? 'This sign-in expired. Try again.' : errorMessage(err)
      finishWith({ step: 'failed', message })
    }
  }

  function cancel() {
    endAttempt(current)
    setPhase({ step: 'idle' })
  }

  return { phase, start, cancel }
}

function endAttempt(current: { current: Attempt | null }) {
  current.current?.abort.abort()
  current.current?.popup?.close()
  current.current = null
}

function openPlexPopup() {
  const width = 600
  const height = 720
  const left = Math.round(window.screenX + (window.outerWidth - width) / 2)
  const top = Math.round(window.screenY + (window.outerHeight - height) / 2)
  return window.open('', 'plex-sign-in', `popup,width=${width},height=${height},left=${left},top=${top}`)
}
