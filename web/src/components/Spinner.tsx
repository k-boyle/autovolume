import { LoaderCircle } from 'lucide-react'

export function Spinner({ className = 'size-5' }: { className?: string }) {
  return <LoaderCircle aria-hidden className={`animate-spin ${className}`} />
}

export function FullPageSpinner() {
  return (
    <div className="grid min-h-dvh place-items-center text-muted" role="status" aria-label="Loading">
      <Spinner className="size-6" />
    </div>
  )
}
