import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight, Volume2, X } from 'lucide-react'
import { useEffect, useId, useRef, useState, type MouseEvent } from 'react'
import { errorMessage } from '../../api/errors'
import { invalidateVolumeRuleQueries } from '../../api/keys'
import { LibraryService } from '../../gen/autovolume/v1/library_pb'
import { MediaType, type MediaItem } from '../../gen/autovolume/v1/media_pb'
import { VolumeService } from '../../gen/autovolume/v1/volume_pb'
import { itemName } from '../../lib/format'
import { Button } from '../Button'
import { Notice } from '../Notice'
import { Poster } from '../Poster'
import { Spinner } from '../Spinner'
import { VolumeControl } from '../VolumeControl'

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
      {item && <DialogContent key={item.ratingKey} item={item} onClose={onClose} />}
    </dialog>
  )
}

function DialogContent({ item, onClose }: { item: MediaItem; onClose: () => void }) {
  const [season, setSeason] = useState<MediaItem | null>(null)

  if (season) {
    const back = () => setSeason(null)
    return <Editor key={season.ratingKey} item={season} onDone={back} onClose={onClose} backTo={item.title} />
  }
  return (
    <>
      <Editor item={item} onDone={onClose} onClose={onClose} />
      {item.type === MediaType.SHOW && <SeasonList show={item} onPick={setSeason} />}
    </>
  )
}

interface EditorProps {
  item: MediaItem
  onDone: () => void
  onClose: () => void
  backTo?: string
}

function Editor({ item, onDone, onClose, backTo }: EditorProps) {
  const queryClient = useQueryClient()
  const isSeason = item.type === MediaType.SEASON
  const defaults = useQuery(VolumeService.method.getDefaultVolume)
  const rules = useQuery(VolumeService.method.listVolumeRules, {}, { enabled: isSeason })
  const showVolume = isSeason ? rules.data?.items.find((r) => r.ratingKey === item.showRatingKey)?.volume : undefined
  const [draft, setDraft] = useState<number>()
  const value = draft ?? item.volume ?? showVolume ?? defaults.data?.volume ?? 20

  const onSuccess = () => {
    invalidateVolumeRuleQueries(queryClient)
    onDone()
  }
  const save = useMutation(VolumeService.method.setVolumeRule, { onSuccess })
  const clear = useMutation(VolumeService.method.clearVolumeRule, { onSuccess })
  const error = save.error ?? clear.error

  let status: string
  if (item.volume !== undefined) status = `Saved volume ${item.volume}`
  else if (showVolume !== undefined) status = `Uses the show's volume (${showVolume})`
  else status = 'No saved volume'

  let explanation: string
  if (isSeason) explanation = `Overrides the show's volume for every episode in ${item.title}.`
  else if (item.type === MediaType.SHOW) {
    explanation = 'The TV switches to this whenever an episode starts or resumes, unless its season has its own volume.'
  } else explanation = 'The TV switches to this whenever this movie starts or resumes.'

  return (
    <div className="p-5">
      {backTo && (
        <button
          type="button"
          onClick={onDone}
          className="mb-3 -ml-1 inline-flex items-center gap-1 rounded text-sm text-muted hover:text-fg focus-visible:outline-2 focus-visible:outline-accent"
        >
          <ChevronLeft aria-hidden className="size-4" />
          {backTo}
        </button>
      )}
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
      {error && (
        <div className="mt-3">
          <Notice tone="error">{errorMessage(error)}</Notice>
        </div>
      )}

      <div className="mt-6 flex flex-wrap items-center gap-2">
        {item.volume !== undefined && (
          <Button variant="danger" loading={clear.isPending} onClick={() => clear.mutate({ ratingKey: item.ratingKey })}>
            {isSeason ? 'Remove override' : 'Remove'}
          </Button>
        )}
        <div className="ml-auto flex gap-2">
          <Button variant="ghost" onClick={onDone}>
            Cancel
          </Button>
          <Button
            variant="primary"
            loading={save.isPending}
            disabled={value === item.volume}
            onClick={() => save.mutate({ ratingKey: item.ratingKey, volume: value })}
          >
            Save {value}
          </Button>
        </div>
      </div>
    </div>
  )
}

function SeasonList({ show, onPick }: { show: MediaItem; onPick: (season: MediaItem) => void }) {
  const seasons = useQuery(LibraryService.method.listSeasons, { showRatingKey: show.ratingKey })
  const headingId = useId()

  return (
    <section aria-labelledby={headingId} className="border-t border-border px-5 pt-4 pb-5">
      <h3 id={headingId} className="text-sm font-semibold">
        Seasons
      </h3>
      <p className="mt-1 text-sm text-muted">Give a season its own volume to override the show's.</p>
      {seasons.isPending && (
        <div className="mt-3">
          <Spinner />
        </div>
      )}
      {seasons.isError && (
        <div className="mt-3">
          <Notice tone="error">{errorMessage(seasons.error)}</Notice>
        </div>
      )}
      {seasons.data && (
        <ul className="mt-2 divide-y divide-border">
          {seasons.data.seasons.map((season) => (
            <li key={season.ratingKey}>
              <button
                type="button"
                onClick={() => onPick(season)}
                className="flex w-full items-center gap-3 rounded py-2.5 text-left text-sm hover:text-accent-fg focus-visible:outline-2 focus-visible:outline-accent"
              >
                <span className="flex-1 truncate font-medium">{season.title}</span>
                {season.volume !== undefined ? (
                  <span className="flex items-center gap-1 rounded-full bg-accent px-2 py-0.5 text-xs font-semibold text-on-accent">
                    <Volume2 aria-hidden className="size-3" />
                    <span className="sr-only">Own volume</span>
                    {season.volume}
                  </span>
                ) : (
                  <span className="text-muted">Same as show</span>
                )}
                <ChevronRight aria-hidden className="size-4 text-muted" />
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
