import { useQuery } from '@connectrpc/connect-query'
import { Navigate } from 'react-router'
import { Button } from '../components/Button'
import { CenteredPanel } from '../components/CenteredPanel'
import { Logo } from '../components/Logo'
import { Notice } from '../components/Notice'
import { Spinner } from '../components/Spinner'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import { usePlexSignIn } from '../hooks/usePlexSignIn'

export function LoginPage() {
  const session = useQuery(AuthService.method.getSession, {}, { retry: false })
  const { phase, start, cancel } = usePlexSignIn()

  if (session.isSuccess) return <Navigate to="/" replace />

  return (
    <CenteredPanel>
      <Logo />
      <h1 className="mt-8 text-2xl font-semibold tracking-tight">The right volume for whatever's on</h1>
      <p className="mt-3 text-muted">
        Save a volume for each show and movie once. When it starts playing on Plex, AutoVolume sets your TV to it.
      </p>
      {phase.step === 'waiting' ? (
        <div className="mt-8 space-y-4">
          {phase.blockedAuthUrl ? (
            <>
              <Notice tone="info">Your browser blocked the Plex window. Open it yourself, then come back to this tab.</Notice>
              <a
                href={phase.blockedAuthUrl}
                target="_blank"
                rel="noreferrer"
                className="inline-flex h-10 w-full items-center justify-center rounded-lg bg-accent px-4 text-sm font-medium text-on-accent hover:bg-accent-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              >
                Open Plex sign-in
              </a>
            </>
          ) : (
            <p className="flex items-center gap-3 text-muted" role="status">
              <Spinner />
              Waiting for you to sign in to Plex in the other window…
            </p>
          )}
          <Button variant="ghost" className="w-full" onClick={cancel}>
            Cancel
          </Button>
        </div>
      ) : (
        <>
          <Button
            variant="primary"
            className="mt-8 w-full"
            loading={phase.step === 'starting'}
            onClick={() => void start()}
          >
            Sign in with Plex
          </Button>
          {phase.step === 'failed' && (
            <div className="mt-4">
              <Notice tone="error">{phase.message}</Notice>
            </div>
          )}
        </>
      )}
    </CenteredPanel>
  )
}
