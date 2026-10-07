import { useQuery } from '@connectrpc/connect-query'
import { Navigate, Outlet } from 'react-router'
import { ErrorScreen } from '../components/ErrorScreen'
import { FullPageSpinner } from '../components/Spinner'
import { SetupService } from '../gen/autovolume/v1/setup_pb'
import { hasRequiredSetup } from '../lib/setup'

export function RequireSetup() {
  const setup = useQuery(SetupService.method.getSetupStatus)
  if (setup.isPending) return <FullPageSpinner />
  if (setup.isError) return <ErrorScreen error={setup.error} onRetry={() => void setup.refetch()} />
  if (!hasRequiredSetup(setup.data.status)) return <Navigate to="/setup" replace />
  return <Outlet />
}
