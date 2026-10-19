import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { Pause, Play, Tv } from 'lucide-react'
import { useState } from 'react'
import { errorMessage } from '../../api/errors'
import { invalidateVolumeRuleQueries } from '../../api/keys'
import { MediaType } from '../../gen/autovolume/v1/media_pb'
import { PlaybackState, type Playback } from '../../gen/autovolume/v1/now_playing_pb'
import { SetupService } from '../../gen/autovolume/v1/setup_pb'
import { TvPower, TvService, type TvState } from '../../gen/autovolume/v1/tv_pb'
import { VolumeService } from '../../gen/autovolume/v1/volume_pb'
import { useNowPlaying } from '../../hooks/useNowPlaying'
import { episodeLabel } from '../../lib/format'
import { Button } from '../Button'
import { Card } from '../Card'
import { Notice } from '../Notice'
import { Poster } from '../Poster'
import { Spinner } from '../Spinner'
import { VolumeControl } from '../VolumeControl'

export function NowPlayingCard() {
  const { snapshot } = useNowPlaying()
  const setup = useQuery(SetupService.method.getSetupStatus)
  const playerName = setup.data?.status?.player?.name ?? 'your TV'
  const playback = snapshot?.playback

  if (!snapshot) {
    return (
      <Card className="grid min-h-64 place-items-center text-muted">
        <Spinner />
      </Card>
    )
  }

  if (!playback?.item) {
    return (
      <Card className="grid min-h-64 place-items-center text-center">
        <div>
          <Tv aria-hidden className="mx-auto size-10 text-muted" />
          <h2 className="mt-4 text-lg font-semibold">Nothing playing on {playerName}</h2>
          <p className="mt-1 text-sm text-muted">Start something on Plex and it'll show up here.</p>
        </div>
      </Card>
    )
  }

  return (
    <PlayingCard
      key={playback.season?.ratingKey ?? playback.item.ratingKey}
      playback={playback}
      tv={snapshot.tv}
      defaultVolume={snapshot.defaultVolume}
      playerName={playerName}
    />
  )
}

interface PlayingCardProps {
  playback: Playback
  tv?: TvState
  defaultVolume?: number
  playerName: string
}

function PlayingCard({ playback, tv, defaultVolume, playerName }: PlayingCardProps) {
  const item = playback.item!
  const season = playback.season
  const rule = season?.volume ?? item.volume
  const kind = item.type === MediaType.SHOW ? 'show' : 'movie'
  const [scope, setScope] = useState<'item' | 'season'>(season?.volume !== undefined ? 'season' : 'item')
  const target = scope === 'season' && season ? season : item
  const targetName = target === season ? season.title : `this ${kind}`
  const playing = playback.state === PlaybackState.PLAYING
  const tvOn = tv?.power === TvPower.ACTIVE
  const tvVolume = tvOn ? tv?.volume : undefined

  const queryClient = useQueryClient()
  const [draft, setDraft] = useState<number>()
  const value = draft ?? target.volume ?? rule ?? defaultVolume ?? tvVolume ?? 20

  const save = useMutation(VolumeService.method.setVolumeRule, { onSuccess: () => invalidateVolumeRuleQueries(queryClient) })
  const clear = useMutation(VolumeService.method.clearVolumeRule, { onSuccess: () => invalidateVolumeRuleQueries(queryClient) })
  const apply = useMutation(TvService.method.setTvVolume)
  const error = save.error ?? clear.error ?? apply.error

  let ruleLine: string
  if (season?.volume !== undefined) {
    ruleLine = `Saved volume for ${season.title}: ${season.volume}`
    if (item.volume !== undefined) ruleLine += ` (show: ${item.volume})`
  } else if (item.volume !== undefined) ruleLine = `Saved volume for this ${kind}: ${item.volume}`
  else if (defaultVolume !== undefined) ruleLine = `No saved volume yet, so it plays at your default (${defaultVolume}).`
  else ruleLine = `No saved volume or default, so AutoVolume leaves the TV alone.`

  return (
    <Card className="flex flex-col gap-6 sm:flex-row">
      <div className="w-28 shrink-0 sm:w-44">
        <Poster title={item.title} thumbUrl={item.thumbUrl} displayWidth={176} />
      </div>
      <div className="flex min-w-0 flex-1 flex-col">
        <p className="flex items-center gap-2 text-xs font-semibold tracking-wide text-accent-fg uppercase">
          {playing ? <Play aria-hidden className="size-3.5 fill-current" /> : <Pause aria-hidden className="size-3.5 fill-current" />}
          {playing ? 'Now playing' : 'Paused'} on {playerName}
        </p>
        <h2 className="mt-2 text-2xl font-semibold tracking-tight text-balance sm:text-3xl">
          {item.title}
          {item.year > 0 && <span className="ml-2 text-lg font-normal text-muted">{item.year}</span>}
        </h2>
        {playback.episode && <p className="mt-1 text-muted">{episodeLabel(playback.episode)}</p>}

        <div className="mt-auto pt-6">
          <div className="mb-2 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 text-sm">
            <span className="text-muted">{ruleLine}</span>
            {tvVolume !== undefined && <span className="text-muted tabular-nums">TV is at {tvVolume}</span>}
          </div>
          <VolumeControl value={value} onChange={setDraft} label={`Volume for ${item.title}`} />
          {season && (
            <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
              <span className="text-muted">Save for</span>
              <div role="radiogroup" aria-label="Save for" className="inline-flex rounded-lg border border-border p-0.5">
                {(['item', 'season'] as const).map((option) => (
                  <button
                    key={option}
                    type="button"
                    role="radio"
                    aria-checked={scope === option}
                    onClick={() => setScope(option)}
                    className={`rounded-md px-3 py-1.5 font-medium transition-colors focus-visible:outline-2 focus-visible:outline-accent ${
                      scope === option ? 'bg-surface-2 text-fg' : 'text-muted hover:text-fg'
                    }`}
                  >
                    {option === 'season' ? season.title : 'Whole show'}
                  </button>
                ))}
              </div>
            </div>
          )}
          <div className="mt-4 flex flex-wrap gap-2">
            <Button
              variant="primary"
              loading={save.isPending}
              disabled={value === target.volume}
              onClick={() => save.mutate({ ratingKey: target.ratingKey, volume: value })}
            >
              Save {value} for {targetName}
            </Button>
            <Button
              loading={apply.isPending}
              disabled={!tvOn || value === tvVolume}
              onClick={() => apply.mutate({ volume: value })}
            >
              Set TV to {value} now
            </Button>
            {target.volume !== undefined && (
              <Button
                variant="ghost"
                loading={clear.isPending}
                onClick={() => clear.mutate({ ratingKey: target.ratingKey })}
              >
                Clear volume for {targetName}
              </Button>
            )}
          </div>
          {error && (
            <div className="mt-3">
              <Notice tone="error">{errorMessage(error)}</Notice>
            </div>
          )}
        </div>
      </div>
    </Card>
  )
}
