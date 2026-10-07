import { timestampDate } from '@bufbuild/protobuf/wkt'
import { useMutation, useQuery } from '@connectrpc/connect-query'
import { Tv } from 'lucide-react'
import { errorMessage } from '../../api/errors'
import { SetupService } from '../../gen/autovolume/v1/setup_pb'
import { timeAgo } from '../../lib/format'
import { Notice } from '../Notice'
import { Spinner } from '../Spinner'
import { OptionRow } from './OptionRow'
import { StepHeading } from './StepHeading'

export function PlayerStep({ selectedId, onDone }: { selectedId?: string; onDone: () => void }) {
  const players = useQuery(SetupService.method.listPlayers, {}, { refetchInterval: 3000 })
  const select = useMutation(SetupService.method.selectPlayer, { onSuccess: onDone })

  return (
    <>
      <StepHeading
        title="Which Plex player is on your TV?"
        description="Only playback on this player changes the TV volume, so watching on your phone won't."
      />
      {players.isPending && <Spinner />}
      {players.isError && <Notice tone="error">{errorMessage(players.error)}</Notice>}
      <div className="space-y-2">
        {players.data?.players.map((player) => (
          <OptionRow
            key={player.id}
            icon={Tv}
            title={player.name}
            detail={
              <>
                {player.product}
                {player.lastSeen && <> · played {timeAgo(timestampDate(player.lastSeen))}</>}
              </>
            }
            selected={player.id === selectedId}
            loading={select.isPending && select.variables.playerId === player.id}
            onSelect={() => select.mutate({ playerId: player.id })}
          />
        ))}
      </div>
      {select.isError && (
        <div className="mt-4">
          <Notice tone="error">{errorMessage(select.error)}</Notice>
        </div>
      )}
      <p className="mt-6 flex items-center gap-2 text-sm text-muted">
        <Spinner className="size-4" />
        Don't see your TV? Start playing something on it and it'll appear here.
      </p>
    </>
  )
}
