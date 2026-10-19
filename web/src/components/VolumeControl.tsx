import { Minus, Plus } from 'lucide-react'

interface Props {
  value: number
  onChange: (value: number) => void
  label: string
  disabled?: boolean
}

const clamp = (v: number) => Math.min(100, Math.max(0, v))

export function VolumeControl({ value, onChange, label, disabled }: Props) {
  const stepButton = 'grid size-9 shrink-0 place-items-center rounded-lg text-muted hover:bg-surface-2 hover:text-fg disabled:opacity-40'
  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        className={stepButton}
        onClick={() => onChange(clamp(value - 1))}
        disabled={disabled || value <= 0}
        aria-label={`Decrease ${label}`}
      >
        <Minus className="size-4" />
      </button>
      <input
        type="range"
        min={0}
        max={100}
        step={1}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        disabled={disabled}
        aria-label={label}
        className="h-2 w-0 min-w-0 flex-1 cursor-pointer disabled:cursor-not-allowed"
      />
      <button
        type="button"
        className={stepButton}
        onClick={() => onChange(clamp(value + 1))}
        disabled={disabled || value >= 100}
        aria-label={`Increase ${label}`}
      >
        <Plus className="size-4" />
      </button>
      <output className="w-9 text-right text-2xl font-semibold tabular-nums" aria-live="polite">
        {value}
      </output>
    </div>
  )
}
