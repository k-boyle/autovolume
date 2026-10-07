import { VolumeX } from 'lucide-react'
import { TvPower } from '../../gen/autovolume/v1/tv_pb'
import { useNowPlaying } from '../../hooks/useNowPlaying'
import type { StreamStatus } from '../../hooks/useNowPlayingStream'
import { Card, CardTitle } from '../Card'
import { DefaultVolumeForm } from '../DefaultVolumeForm'

const powerLabels: Record<TvPower, string> = {
  [TvPower.UNSPECIFIED]: 'Not set up',
  [TvPower.UNREACHABLE]: "Can't reach the TV",
  [TvPower.STANDBY]: 'In standby',
  [TvPower.ACTIVE]: 'Current volume',
}

export function TvCard() {
  const { snapshot, status } = useNowPlaying()
  const tv = snapshot?.tv
  const on = tv?.power === TvPower.ACTIVE
  const volume = on ? tv?.volume : undefined

  return (
    <Card className="flex flex-col">
      <CardTitle action={<StreamBadge status={status} />}>Your TV</CardTitle>
      <p className="text-sm text-muted">{tv ? powerLabels[tv.power] : 'Connecting…'}</p>
      <p className="mt-1 flex items-center gap-3 text-6xl font-semibold tracking-tight tabular-nums">
        {volume ?? '–'}
        {on && tv?.muted && (
          <span className="flex items-center gap-1 text-sm font-medium text-muted">
            <VolumeX aria-hidden className="size-4" /> Muted
          </span>
        )}
      </p>

      <div className="mt-6 border-t border-border pt-5">
        <h3 className="font-medium">Default volume</h3>
        <p className="mt-1 mb-4 text-sm text-muted">Used on pause and stop, and for anything without its own volume.</p>
        <DefaultVolumeForm tvVolume={volume} />
      </div>
    </Card>
  )
}

function StreamBadge({ status }: { status: StreamStatus }) {
  const live = status === 'live'
  return (
    <span className="flex items-center gap-1.5 text-xs text-muted">
      <span className={`size-2 rounded-full ${live ? 'bg-ok' : 'animate-pulse bg-accent'}`} aria-hidden />
      {live ? 'Live' : status === 'connecting' ? 'Connecting' : 'Reconnecting'}
    </span>
  )
}
