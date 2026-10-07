import { ChevronRight, type LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { Spinner } from '../Spinner'

interface Props {
  icon: LucideIcon
  title: string
  detail: ReactNode
  selected: boolean
  loading: boolean
  disabled?: boolean
  onSelect: () => void
}

export function OptionRow({ icon: Icon, title, detail, selected, loading, disabled, onSelect }: Props) {
  return (
    <button
      type="button"
      onClick={onSelect}
      disabled={disabled || loading}
      aria-pressed={selected}
      className={`flex w-full items-center gap-4 rounded-xl border p-4 text-left transition-colors disabled:cursor-not-allowed ${
        selected ? 'border-accent bg-surface-2' : 'border-border hover:border-muted'
      } ${disabled ? 'opacity-50' : ''}`}
    >
      <Icon aria-hidden className="size-6 shrink-0 text-muted" />
      <span className="min-w-0 flex-1">
        <span className="block truncate font-medium">{title}</span>
        <span className="mt-0.5 block text-sm text-muted">{detail}</span>
      </span>
      {loading ? (
        <Spinner />
      ) : selected ? (
        <span className="rounded-full bg-accent px-2 py-0.5 text-xs font-semibold text-on-accent">Selected</span>
      ) : (
        !disabled && <ChevronRight aria-hidden className="size-5 text-muted" />
      )}
    </button>
  )
}
