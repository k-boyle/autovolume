import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { useState, type FormEvent } from 'react'
import { errorMessage } from '../api/errors'
import { allQueriesKey } from '../api/keys'
import { SetupService } from '../gen/autovolume/v1/setup_pb'
import { TvPower, TvService } from '../gen/autovolume/v1/tv_pb'
import { Button } from './Button'
import { Notice } from './Notice'

const inputClass =
  'mt-1.5 h-10 w-full rounded-lg border border-border bg-bg px-3 text-sm placeholder:text-muted focus:border-accent focus:outline-none'

export function TvConnectionForm({ onSaved, submitLabel = 'Test and save' }: { onSaved?: () => void; submitLabel?: string }) {
  const queryClient = useQueryClient()
  const config = useQuery(TvService.method.getTvConfig)
  const saved = config.data?.config
  const [address, setAddress] = useState<string>()
  const [psk, setPsk] = useState('')

  const test = useMutation(TvService.method.testTvConnection)
  const save = useMutation(TvService.method.updateTvConfig, {
    onSuccess: () => {
      setPsk('')
      void queryClient.invalidateQueries({ queryKey: allQueriesKey(TvService.method.getTvConfig) })
      void queryClient.invalidateQueries({ queryKey: allQueriesKey(SetupService.method.getSetupStatus) })
      onSaved?.()
    },
  })

  const addressValue = address ?? saved?.address ?? ''
  const input = { address: addressValue.trim(), psk: psk || undefined }
  const busy = test.isPending || save.isPending

  function submit(e: FormEvent) {
    e.preventDefault()
    save.reset()
    test.mutate(input, { onSuccess: () => save.mutate(input) })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <label className="block text-sm font-medium">
        TV IP address
        <input
          className={inputClass}
          value={addressValue}
          onChange={(e) => setAddress(e.target.value)}
          placeholder="192.168.1.50"
          autoComplete="off"
          inputMode="url"
          required
        />
      </label>
      <label className="block text-sm font-medium">
        Pre-shared key
        <input
          className={inputClass}
          type="password"
          value={psk}
          onChange={(e) => setPsk(e.target.value)}
          placeholder={saved?.pskSet ? 'Saved. Leave blank to keep it.' : ''}
          autoComplete="off"
          required={!saved?.pskSet}
        />
      </label>

      <details className="text-sm text-muted">
        <summary className="cursor-pointer select-none hover:text-fg">Where do I find these?</summary>
        <ol className="mt-2 list-decimal space-y-1 pl-5">
          <li>
            On the TV, open <b>Settings → Network &amp; Internet → Home network → IP control</b>.
          </li>
          <li>
            Set <b>Authentication</b> to <b>Normal and Pre-Shared Key</b>, then set a <b>Pre-Shared Key</b>.
          </li>
          <li>The IP address is on the TV's network status screen, or in your router's device list.</li>
        </ol>
        <p className="mt-2">Menu names vary a little between Bravia models.</p>
      </details>

      {test.isSuccess && test.data.state && (
        <Notice tone="success">
          Connected to {test.data.model || 'your TV'}
          {test.data.state.power === TvPower.ACTIVE && test.data.state.volume !== undefined
            ? `, volume ${test.data.state.volume}`
            : ', currently in standby'}
          .
        </Notice>
      )}
      {test.isError && <Notice tone="error">{errorMessage(test.error)}</Notice>}
      {save.isError && <Notice tone="error">{errorMessage(save.error)}</Notice>}

      <div className="flex flex-wrap gap-2">
        <Button type="submit" variant="primary" loading={busy}>
          {submitLabel}
        </Button>
        {test.isError && (
          <Button onClick={() => save.mutate(input)} disabled={busy || !input.address}>
            Save anyway
          </Button>
        )}
      </div>
    </form>
  )
}
