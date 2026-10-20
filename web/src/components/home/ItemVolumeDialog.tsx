import { createClient, type ConnectError } from '@connectrpc/connect'
import { useMutation as useRpcMutation, useQuery, useTransport } from '@connectrpc/connect-query'
import { useMutation, useQueryClient, type UseQueryResult } from '@tanstack/react-query'
import { X } from 'lucide-react'
import { useEffect, useId, useRef, useState, type MouseEvent } from 'react'
import { errorMessage } from '../../api/errors'
import { invalidateVolumeRuleQueries } from '../../api/keys'
import { LibraryService, type ListSeasonsResponse } from '../../gen/autovolume/v1/library_pb'
import { MediaType, type MediaItem } from '../../gen/autovolume/v1/media_pb'
import { VolumeService } from '../../gen/autovolume/v1/volume_pb'
import { itemName } from '../../lib/format'
import { Button } from '../Button'
import { Notice } from '../Notice'
import { Poster } from '../Poster'
import { Spinner } from '../Spinner'
import { VolumeControl } from '../VolumeControl'
import { VolumeStepper } from '../VolumeStepper'

export function ItemVolumeDialog({ item, onClose }: { item: MediaItem | null; onClose: () => void }) {
  const ref = useRef<HTMLDialogElement>(null)
  const closeOnBackdropClick = (e: MouseEvent<HTMLDialogElement>) => e.target === e.currentTarget && onClose()

  useEffect(() => {
    const dialog = ref.current
    if (!dialog) return
    if (item && !dialog.open) dialog.showModal()
    if (!item && dialog.open) dialog.close()
  }, [item])

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      onClick={closeOnBackdropClick}
      aria-label={item ? `Volume for ${itemName(item)}` : undefined}
      className="m-auto max-h-[calc(100dvh-2rem)] w-[min(30rem,calc(100vw-2rem))] overflow-y-auto rounded-2xl border border-border bg-surface p-0 text-fg"
    >
      {item && <Editor key={item.ratingKey} item={item} onClose={onClose} />}
    </dialog>
  )
}

interface VolumeChange {
  ratingKey: string
  volume?: number
}

function Editor({ item, onClose }: { item: MediaItem; onClose: () => void }) {
  const queryClient = useQueryClient()
  const transport = useTransport()
  const isSeason = item.type === MediaType.SEASON
  const isShow = item.type === MediaType.SHOW
  const defaults = useQuery(VolumeService.method.getDefaultVolume)
  const rules = useQuery(VolumeService.method.listVolumeRules, {}, { enabled: isSeason })
  const seasons = useQuery(LibraryService.method.listSeasons, { showRatingKey: item.ratingKey }, { enabled: isShow })
  const showVolume = isSeason ? rules.data?.items.find((r) => r.ratingKey === item.showRatingKey)?.volume : undefined
  const [draft, setDraft] = useState<number>()
  const [overrides, setOverrides] = useState<Record<string, number | null>>({})
  const value = draft ?? item.volume ?? showVolume ?? defaults.data?.volume ?? 20

  const seasonVolume = (season: MediaItem) => {
    const override = overrides[season.ratingKey]
    return override === undefined ? season.volume : (override ?? undefined)
  }
  const seasonChanges: VolumeChange[] = (seasons.data?.seasons ?? [])
    .filter((season) => seasonVolume(season) !== season.volume)
    .map((season) => ({ ratingKey: season.ratingKey, volume: seasonVolume(season) }))
  const saveItem = value !== item.volume && (draft !== undefined || seasonChanges.length === 0)
  const changes = saveItem ? [{ ratingKey: item.ratingKey, volume: value }, ...seasonChanges] : seasonChanges

  const save = useMutation({
    mutationFn: (changes: VolumeChange[]) => {
      const client = createClient(VolumeService, transport)
      return Promise.all(
        changes.map(({ ratingKey, volume }) =>
          volume === undefined ? client.clearVolumeRule({ ratingKey }) : client.setVolumeRule({ ratingKey, volume }),
        ),
      )
    },
    onSettled: () => invalidateVolumeRuleQueries(queryClient),
    onSuccess: onClose,
  })
  const clear = useRpcMutation(VolumeService.method.clearVolumeRule, {
    onSuccess: () => {
      invalidateVolumeRuleQueries(queryClient)
      onClose()
    },
  })
  const error = save.error ?? clear.error

  let status: string
  if (item.volume !== undefined) status = `Saved volume ${item.volume}`
  else if (showVolume !== undefined) status = `Uses the show's volume (${showVolume})`
  else status = 'No saved volume'

  let explanation: string
  if (isSeason) explanation = `Overrides the show's volume for every episode in ${item.title}.`
  else if (isShow) {
    explanation = 'The TV switches to this whenever an episode starts or resumes, unless its season has its own volume.'
  } else explanation = 'The TV switches to this whenever this movie starts or resumes.'

  return (
    <>
      <div className="p-5">
        <div className="flex gap-4">
          <div className="w-16 shrink-0">
            <Poster title={itemName(item)} thumbUrl={item.thumbUrl} displayWidth={64} />
          </div>
          <div className="min-w-0 flex-1">
            <h2 className="text-lg leading-snug font-semibold">{isSeason ? item.showTitle : item.title}</h2>
            <p className="text-sm text-muted">
              {isSeason && `${item.title} · `}
              {!isSeason && item.year > 0 && `${item.year} · `}
              {status}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="grid size-8 shrink-0 place-items-center rounded-lg text-muted hover:bg-surface-2 hover:text-fg"
          >
            <X className="size-4" />
          </button>
        </div>

        <p className="mt-5 mb-3 text-sm text-muted">{explanation}</p>
        <VolumeControl value={value} onChange={setDraft} label={`Volume for ${itemName(item)}`} />
      </div>

      {isShow && (
        <SeasonOverrides
          seasons={seasons}
          volumeOf={seasonVolume}
          showVolume={value}
          onChange={(season, volume) => setOverrides((o) => ({ ...o, [season.ratingKey]: volume }))}
        />
      )}

      <div className="sticky bottom-0 border-t border-border bg-surface px-5 py-4">
        {error && (
          <div className="mb-3">
            <Notice tone="error">{errorMessage(error)}</Notice>
          </div>
        )}
        <div className="flex flex-wrap items-center gap-2">
          {item.volume !== undefined && (
            <Button variant="danger" loading={clear.isPending} onClick={() => clear.mutate({ ratingKey: item.ratingKey })}>
              {isSeason ? 'Remove override' : 'Remove'}
            </Button>
          )}
          <div className="ml-auto flex gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button
              variant="primary"
              loading={save.isPending}
              disabled={changes.length === 0}
              onClick={() => save.mutate(changes)}
            >
              {seasonChanges.length > 0 ? 'Save changes' : `Save ${value}`}
            </Button>
          </div>
        </div>
      </div>
    </>
  )
}

interface SeasonOverridesProps {
  seasons: UseQueryResult<ListSeasonsResponse, ConnectError>
  volumeOf: (season: MediaItem) => number | undefined
  showVolume: number
  onChange: (season: MediaItem, volume: number | null) => void
}

function SeasonOverrides({ seasons, volumeOf, showVolume, onChange }: SeasonOverridesProps) {
  const headingId = useId()

  return (
    <section aria-labelledby={headingId} className="border-t border-border px-5 pt-4 pb-2">
      <h3 id={headingId} className="text-sm font-semibold">
        Seasons
      </h3>
      <p className="mt-1 text-sm text-muted">Seasons use the show's volume unless you override them.</p>
      {seasons.isPending && (
        <div className="my-3">
          <Spinner />
        </div>
      )}
      {seasons.isError && (
        <div className="my-3">
          <Notice tone="error">{errorMessage(seasons.error)}</Notice>
        </div>
      )}
      {seasons.data && (
        <ul className="mt-2 divide-y divide-border">
          {seasons.data.seasons.map((season) => (
            <SeasonRow
              key={season.ratingKey}
              season={season}
              volume={volumeOf(season)}
              onOverride={() => onChange(season, showVolume)}
              onChange={(volume) => onChange(season, volume)}
            />
          ))}
        </ul>
      )}
    </section>
  )
}

interface SeasonRowProps {
  season: MediaItem
  volume: number | undefined
  onOverride: () => void
  onChange: (volume: number | null) => void
}

function SeasonRow({ season, volume, onOverride, onChange }: SeasonRowProps) {
  const [focus, setFocus] = useState<'stepper' | 'override'>()

  return (
    <li className="flex min-h-12 items-center gap-2 py-1.5">
      <span className="min-w-0 flex-1 truncate text-sm font-medium">{season.title}</span>
      {volume === undefined ? (
        <>
          <span className="text-sm text-muted">Same as show</span>
          <Button
            size="sm"
            autoFocus={focus === 'override'}
            aria-label={`Override volume for ${season.title}`}
            onClick={() => {
              setFocus('stepper')
              onOverride()
            }}
          >
            Override
          </Button>
        </>
      ) : (
        <>
          <VolumeStepper
            value={volume}
            onChange={onChange}
            label={`Volume for ${season.title}`}
            autoFocus={focus === 'stepper'}
          />
          <button
            type="button"
            title="Use the show's volume"
            aria-label={`Use the show's volume for ${season.title}`}
            onClick={() => {
              setFocus('override')
              onChange(null)
            }}
            className="grid size-8 shrink-0 place-items-center rounded-lg text-muted hover:bg-surface-2 hover:text-fg"
          >
            <X className="size-4" />
          </button>
        </>
      )}
    </li>
  )
}
