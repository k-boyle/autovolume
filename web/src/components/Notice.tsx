import { CircleAlert, CircleCheck, Info } from 'lucide-react'
import type { ReactNode } from 'react'

const tones = {
  error: { icon: CircleAlert, color: 'text-danger' },
  success: { icon: CircleCheck, color: 'text-ok' },
  info: { icon: Info, color: 'text-muted' },
}

export function Notice({ tone, children }: { tone: keyof typeof tones; children: ReactNode }) {
  const { icon: Icon, color } = tones[tone]
  return (
    <p role={tone === 'error' ? 'alert' : 'status'} className={`flex items-start gap-2 text-sm ${color}`}>
      <Icon aria-hidden className="mt-0.5 size-4 shrink-0" />
      <span>{children}</span>
    </p>
  )
}
