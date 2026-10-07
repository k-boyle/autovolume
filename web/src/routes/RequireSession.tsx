import { Code } from '@connectrpc/connect'
import { useQuery } from '@connectrpc/connect-query'
import { Navigate, Outlet } from 'react-router'
import { errorCode } from '../api/errors'
import { ErrorScreen } from '../components/ErrorScreen'
import { FullPageSpinner } from '../components/Spinner'
import { AuthService } from '../gen/autovolume/v1/auth_pb'

export function RequireSession() {
  const session = useQuery(AuthService.method.getSession)
  if (session.isPending) return <FullPageSpinner />
  if (session.isError) {
    if (errorCode(session.error) === Code.Unauthenticated) return <Navigate to="/login" replace />
    return <ErrorScreen error={session.error} onRetry={() => void session.refetch()} />
  }
  return <Outlet />
}
