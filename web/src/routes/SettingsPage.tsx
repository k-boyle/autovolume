import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { Link, useNavigate } from 'react-router'
import { Avatar } from '../components/Avatar'
import { Button } from '../components/Button'
import { Card, CardTitle } from '../components/Card'
import { DefaultVolumeForm } from '../components/DefaultVolumeForm'
import { TvConnectionForm } from '../components/TvConnectionForm'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import { SetupService } from '../gen/autovolume/v1/setup_pb'
import { TvPower } from '../gen/autovolume/v1/tv_pb'
import { useNowPlaying } from '../hooks/useNowPlaying'
import type { SetupStepId } from '../lib/setup'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const session = useQuery(AuthService.method.getSession)
  const setup = useQuery(SetupService.method.getSetupStatus)
  const { snapshot } = useNowPlaying()
  const status = setup.data?.status
  const user = session.data?.user
  const tvVolume = snapshot?.tv?.power === TvPower.ACTIVE ? snapshot.tv.volume : undefined

  const logout = useMutation(AuthService.method.logout, {
    onSuccess: () => {
      queryClient.clear()
      void navigate('/login', { replace: true })
    },
  })

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>

      <Card>
        <CardTitle>Plex</CardTitle>
        <dl className="divide-y divide-border">
          <SettingRow
            label="Server"
            value={status?.server?.name}
            detail={
              status?.plexConnected ? (
                <span className="text-ok">Connected</span>
              ) : (
                <span className="text-danger">Not connected. Check the server is online.</span>
              )
            }
            step="server"
          />
          <SettingRow label="Player" value={status?.player?.name} detail={status?.player?.product} step="player" />
        </dl>
      </Card>

      <Card>
        <CardTitle>TV connection</CardTitle>
        <TvConnectionForm submitLabel="Test and save" />
      </Card>

      <Card>
        <CardTitle>Default volume</CardTitle>
        <p className="-mt-2 mb-4 text-sm text-muted">Used on pause and stop, and for anything without its own volume.</p>
        <DefaultVolumeForm tvVolume={tvVolume} />
      </Card>

      <Card>
        <CardTitle>Account</CardTitle>
        <div className="flex items-center gap-3">
          {user && <Avatar name={user.username} url={user.avatarUrl} className="size-10" />}
          <span className="flex-1 font-medium">{user?.username}</span>
          <Button onClick={() => logout.mutate({})} loading={logout.isPending}>
            Sign out
          </Button>
        </div>
      </Card>
    </div>
  )
}

function SettingRow({ label, value, detail, step }: { label: string; value?: string; detail?: ReactNode; step: SetupStepId }) {
  return (
    <div className="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
      <div className="min-w-0 flex-1">
        <dt className="text-sm text-muted">{label}</dt>
        <dd className="truncate font-medium">{value ?? 'Not set'}</dd>
        {detail && <dd className="text-sm text-muted">{detail}</dd>}
      </div>
      <Link
        to={`/setup?step=${step}&return=/settings`}
        className="rounded-lg border border-border px-3 py-1.5 text-sm hover:border-muted"
      >
        Change
      </Link>
    </div>
  )
}
