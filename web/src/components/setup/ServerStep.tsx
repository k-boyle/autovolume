import { useMutation, useQuery } from '@connectrpc/connect-query'
import { Server as ServerIcon } from 'lucide-react'
import { errorMessage } from '../../api/errors'
import { SetupService } from '../../gen/autovolume/v1/setup_pb'
import { Notice } from '../Notice'
import { Spinner } from '../Spinner'
import { OptionRow } from './OptionRow'
import { StepHeading } from './StepHeading'

export function ServerStep({ selectedId, onDone }: { selectedId?: string; onDone: () => void }) {
  const servers = useQuery(SetupService.method.listServers)
  const select = useMutation(SetupService.method.selectServer, { onSuccess: onDone })

  return (
    <>
      <StepHeading
        title="Choose your Plex server"
        description="AutoVolume listens to this server for playback. Plex only sends playback events to a server's owner, so shared servers can't be used."
      />
      {servers.isPending && <Spinner />}
      {servers.isError && <Notice tone="error">{errorMessage(servers.error)}</Notice>}
      {servers.isSuccess && servers.data.servers.length === 0 && (
        <Notice tone="info">No Plex servers found on this account.</Notice>
      )}
      <div className="space-y-2">
        {servers.data?.servers.map((server) => (
          <OptionRow
            key={server.id}
            icon={ServerIcon}
            title={server.name}
            detail={
              <>
                {server.owned ? 'Owned by you' : 'Shared with you'} ·{' '}
                <span className={server.online ? 'text-ok' : ''}>{server.online ? 'Online' : 'Offline'}</span>
              </>
            }
            selected={server.id === selectedId}
            disabled={!server.owned}
            loading={select.isPending && select.variables.serverId === server.id}
            onSelect={() => select.mutate({ serverId: server.id })}
          />
        ))}
      </div>
      {select.isError && (
        <div className="mt-4">
          <Notice tone="error">{errorMessage(select.error)}</Notice>
        </div>
      )}
    </>
  )
}
