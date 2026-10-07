import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { errorMessage } from '../api/errors'
import { allQueriesKey } from '../api/keys'
import { VolumeService } from '../gen/autovolume/v1/volume_pb'
import { Button } from './Button'
import { Notice } from './Notice'
import { VolumeControl } from './VolumeControl'

interface Props {
  tvVolume?: number
  onSaved?: () => void
  saveLabel?: string
}

export function DefaultVolumeForm({ tvVolume, onSaved, saveLabel = 'Save default' }: Props) {
  const queryClient = useQueryClient()
  const current = useQuery(VolumeService.method.getDefaultVolume)
  const saved = current.data?.volume
  const [draft, setDraft] = useState<number>()
  const value = draft ?? saved ?? tvVolume ?? 20

  const save = useMutation(VolumeService.method.setDefaultVolume, {
    onSuccess: () => {
      setDraft(undefined)
      void queryClient.invalidateQueries({ queryKey: allQueriesKey(VolumeService.method.getDefaultVolume) })
      onSaved?.()
    },
  })

  return (
    <div className="space-y-4">
      <VolumeControl value={value} onChange={setDraft} label="Default volume" disabled={current.isPending} />
      {saved === undefined && current.isSuccess && (
        <Notice tone="info">
          No default yet. Until you set one, AutoVolume won't change the volume on pause or stop, or for anything without
          its own volume.
        </Notice>
      )}
      {save.isError && <Notice tone="error">{errorMessage(save.error)}</Notice>}
      <div className="flex flex-wrap items-center gap-2">
        <Button
          variant="primary"
          loading={save.isPending}
          disabled={saved === value}
          onClick={() => save.mutate({ volume: value })}
        >
          {saveLabel}
        </Button>
        {tvVolume !== undefined && tvVolume !== value && (
          <Button variant="ghost" onClick={() => setDraft(tvVolume)}>
            Use the TV's current volume ({tvVolume})
          </Button>
        )}
      </div>
    </div>
  )
}
