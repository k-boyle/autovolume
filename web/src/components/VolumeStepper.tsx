import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'

interface Props {
  value: number
  onChange: (value: number) => void
  label: string
  autoFocus?: boolean
}

const clamp = (v: number) => Math.min(100, Math.max(0, v))

export function VolumeStepper({ value, onChange, label, autoFocus }: Props) {
  const [text, setText] = useState<string>()
  const stepButton = 'grid size-8 shrink-0 place-items-center rounded-lg text-muted hover:bg-surface-2 hover:text-fg disabled:opacity-40'
  return (
    <div className="flex items-center gap-0.5">
      <button
        type="button"
        className={stepButton}
        onClick={() => onChange(clamp(value - 1))}
        disabled={value <= 0}
        aria-label={`Decrease ${label}`}
      >
        <Minus className="size-4" />
      </button>
      <input
        type="number"
        min={0}
        max={100}
        value={text ?? value}
        onChange={(e) => {
          setText(e.target.value)
          if (e.target.value !== '') onChange(clamp(Math.round(e.target.valueAsNumber)))
        }}
        onBlur={() => setText(undefined)}
        aria-label={label}
        autoFocus={autoFocus}
        className="h-8 w-11 rounded-md border border-border bg-bg text-center text-sm font-semibold tabular-nums [appearance:textfield] focus:border-accent focus:outline-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
      />
      <button
        type="button"
        className={stepButton}
        onClick={() => onChange(clamp(value + 1))}
        disabled={value >= 100}
        aria-label={`Increase ${label}`}
      >
        <Plus className="size-4" />
      </button>
    </div>
  )
}
