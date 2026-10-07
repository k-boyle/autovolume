import { Code, createClient } from '@connectrpc/connect'
import { useTransport } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { errorCode, errorMessage } from '../api/errors'
import { allQueriesKey } from '../api/keys'
import { Button } from '../components/Button'
import { CenteredPanel } from '../components/CenteredPanel'
import { Logo } from '../components/Logo'
import { Notice } from '../components/Notice'
import { Spinner } from '../components/Spinner'
import { AuthService } from '../gen/autovolume/v1/auth_pb'

const pollMs = 1000
const giveUpAfterMs = 3 * 60_000

export function LoginCallbackPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const transport = useTransport()
  const [problem, setProblem] = useState<string | null>(null)

  useEffect(() => {
    const client = createClient(AuthService, transport)
    const abort = new AbortController()
    const { signal } = abort

    const forgetSignedOutSession = () => queryClient.removeQueries({ queryKey: allQueriesKey(AuthService.method.getSession) })
    const finish = () => {
      forgetSignedOutSession()
      void navigate('/', { replace: true })
    }

    // CompleteLogin consumes the PIN, so an earlier call whose response we never saw may have signed us in.
    const signedIn = () => client.getSession({}, { signal }).then(() => true, () => false)

    void (async () => {
      const deadline = Date.now() + giveUpAfterMs
      while (!signal.aborted) {
        try {
          const { complete } = await client.completeLogin({}, { signal })
          if (complete) return finish()
        } catch (err) {
          if (signal.aborted) return
          if (errorCode(err) === Code.NotFound) {
            if (await signedIn()) return finish()
            return setProblem('This sign-in expired or was started in a different browser.')
          }
          return setProblem(errorMessage(err))
        }
        if (Date.now() > deadline) return setProblem('Sign-in took too long.')
        await new Promise((resolve) => setTimeout(resolve, pollMs))
      }
    })()

    return () => abort.abort()
  }, [transport, queryClient, navigate])

  return (
    <CenteredPanel>
      <Logo />
      {problem ? (
        <div className="mt-8 space-y-6">
          <Notice tone="error">{problem}</Notice>
          <Button variant="primary" className="w-full" onClick={() => void navigate('/login', { replace: true })}>
            Try again
          </Button>
        </div>
      ) : (
        <p className="mt-8 flex items-center gap-3 text-muted" role="status">
          <Spinner />
          Finishing sign-in with Plex…
        </p>
      )}
    </CenteredPanel>
  )
}
