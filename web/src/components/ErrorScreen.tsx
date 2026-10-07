import { CircleAlert } from 'lucide-react'
import { errorMessage } from '../api/errors'
import { Button } from './Button'

export function ErrorScreen({ error, onRetry }: { error: unknown; onRetry: () => void }) {
  return (
    <div className="grid min-h-dvh place-items-center px-4">
      <div className="max-w-sm text-center">
        <CircleAlert aria-hidden className="mx-auto size-8 text-danger" />
        <h1 className="mt-4 text-lg font-semibold">Can't reach AutoVolume</h1>
        <p className="mt-2 text-sm text-muted">{errorMessage(error)}</p>
        <Button className="mt-6" onClick={onRetry}>
          Try again
        </Button>
      </div>
    </div>
  )
}
