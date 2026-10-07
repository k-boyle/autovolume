import type { ReactNode } from 'react'

export function CenteredPanel({ children }: { children: ReactNode }) {
  return (
    <div className="grid min-h-dvh place-items-center px-4 py-12">
      <div className="w-full max-w-sm">{children}</div>
    </div>
  )
}
