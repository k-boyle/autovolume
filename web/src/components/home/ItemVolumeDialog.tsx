import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { X } from 'lucide-react'
import { useEffect, useRef, useState, type MouseEvent } from 'react'
import { errorMessage } from '../../api/errors'
import { invalidateVolumeRuleQueries } from '../../api/keys'
import { MediaType, type MediaItem } from '../../gen/autovolume/v1/media_pb'
import { VolumeService } from '../../gen/autovolume/v1/volume_pb'
import { Button } from '../Button'
import { Notice } from '../Notice'
import { Poster } from '../Poster'
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
      aria-label={item ? `Volume for ${item.title}` : undefined}
      className="m-auto w-[min(30rem,calc(100vw-2rem))] rounded-2xl border border-border bg-surface p-0 text-fg"
    >
      {item && <Editor key={item.ratingKey} item={item} onDone={onClose} />}
    </dialog>
  )
}

function Editor({ item, onDone }: { item: MediaItem; onDone: () => void }) {
  const queryClient = useQueryClient()
  const defaults = useQuery(VolumeService.method.getDefaultVolume)
  const [draft, setDraft] = useState<number>()
  const value = draft ?? item.volume ?? defaults.data?.volume ?? 20
  const kind = item.type === MediaType.SHOW ? 'show' : 'movie'

  const onSuccess = () => {
    invalidateVolumeRuleQueries(queryClient)
    onDone()
  }
  const save = useMutation(VolumeService.method.setVolumeRule, { onSuccess })
  const clear = useMutation(VolumeService.method.clearVolumeRule, { onSuccess })
  const error = save.error ?? clear.error

  return (
    <div className="p-5">
      <div className="flex gap-4">
        <div className="w-16 shrink-0">
          <Poster title={item.title} thumbUrl={item.thumbUrl} displayWidth={64} />
        </div>
        <div className="min-w-0 flex-1">
          <h2 className="text-lg leading-snug font-semibold">{item.title}</h2>
          <p className="text-sm text-muted">
            {item.year > 0 && `${item.year} · `}
            {item.volume !== undefined ? `Saved volume ${item.volume}` : 'No saved volume'}
          </p>
        </div>
        <button
          type="button"
          onClick={onDone}
          aria-label="Close"
          className="grid size-8 shrink-0 place-items-center rounded-lg text-muted hover:bg-surface-2 hover:text-fg"
        >
          <X className="size-4" />
        </button>
      </div>

      <p className="mt-5 mb-3 text-sm text-muted">
        The TV switches to this whenever this {kind} starts or resumes{kind === 'show' ? ', for every episode' : ''}.
      </p>
      <VolumeControl value={value} onChange={setDraft} label={`Volume for ${item.title}`} />
      {error && (
        <div className="mt-3">
          <Notice tone="error">{errorMessage(error)}</Notice>
        </div>
      )}

      <div className="mt-6 flex flex-wrap items-center gap-2">
        {item.volume !== undefined && (
          <Button variant="danger" loading={clear.isPending} onClick={() => clear.mutate({ ratingKey: item.ratingKey })}>
            Remove
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
