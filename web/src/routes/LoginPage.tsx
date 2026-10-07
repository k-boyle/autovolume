import { useMutation, useQuery } from '@connectrpc/connect-query'
import { Navigate } from 'react-router'
import { errorMessage } from '../api/errors'
import { Button } from '../components/Button'
import { CenteredPanel } from '../components/CenteredPanel'
import { Logo } from '../components/Logo'
import { Notice } from '../components/Notice'
import { AuthService } from '../gen/autovolume/v1/auth_pb'

export function LoginPage() {
  const session = useQuery(AuthService.method.getSession, {}, { retry: false })
  const start = useMutation(AuthService.method.startLogin, {
    onSuccess: ({ authUrl }) => window.location.assign(authUrl),
  })

  if (session.isSuccess) return <Navigate to="/" replace />

  return (
    <CenteredPanel>
      <Logo />
      <h1 className="mt-8 text-2xl font-semibold tracking-tight">The right volume for whatever's on</h1>
      <p className="mt-3 text-muted">
        Save a volume for each show and movie once. When it starts playing on Plex, AutoVolume sets your TV to it.
      </p>
      <Button
        variant="primary"
        className="mt-8 w-full"
        loading={start.isPending || start.isSuccess}
        onClick={() => start.mutate({ forwardUrl: `${window.location.origin}/login/callback` })}
      >
        Sign in with Plex
      </Button>
      {start.isError && (
        <div className="mt-4">
          <Notice tone="error">{errorMessage(start.error)}</Notice>
        </div>
      )}
    </CenteredPanel>
  )
}
