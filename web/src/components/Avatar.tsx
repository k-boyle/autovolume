import { useState } from 'react'

export function Avatar({ name, url, className = 'size-8' }: { name: string; url: string; className?: string }) {
  const [failed, setFailed] = useState(false)
  if (url && !failed) {
    return <img src={url} alt="" onError={() => setFailed(true)} className={`${className} rounded-full object-cover`} />
  }
  return (
    <span
      aria-hidden
      className={`${className} grid place-items-center rounded-full bg-surface-2 text-xs font-semibold text-muted uppercase`}
    >
      {name.slice(0, 1)}
    </span>
  )
}
