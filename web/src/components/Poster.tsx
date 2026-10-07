import { useState } from 'react'

interface Props {
  title: string
  thumbUrl: string
  displayWidth: number
  className?: string
}

export function Poster({ title, thumbUrl, displayWidth, className = '' }: Props) {
  const [failed, setFailed] = useState(false)
  const base = `aspect-[2/3] w-full overflow-hidden rounded-lg bg-surface-2 ${className}`

  if (!thumbUrl || failed) {
    return (
      <div
        className={`${base} flex p-2.5 ${displayWidth < 100 ? 'items-center justify-center' : 'items-end'}`}
        style={{ backgroundImage: placeholderGradient(title) }}
        aria-hidden
      >
        <span className="line-clamp-3 text-sm leading-tight font-semibold text-white/90">
          {displayWidth < 100 ? title.slice(0, 1) : title}
        </span>
      </div>
    )
  }

  return (
    <div className={base}>
      <img
        src={sizedUrl(thumbUrl, displayWidth * 2, displayWidth * 3)}
        alt=""
        loading="lazy"
        decoding="async"
        onError={() => setFailed(true)}
        className="size-full object-cover"
      />
    </div>
  )
}

function sizedUrl(thumbUrl: string, w: number, h: number): string {
  const url = new URL(thumbUrl, window.location.origin)
  url.searchParams.set('w', String(w))
  url.searchParams.set('h', String(h))
  return url.pathname + url.search
}

function placeholderGradient(title: string): string {
  let hash = 0
  for (const ch of title) hash = (hash * 31 + ch.charCodeAt(0)) | 0
  const hue = Math.abs(hash) % 360
  return `linear-gradient(160deg, hsl(${hue} 45% 38%), hsl(${(hue + 40) % 360} 50% 18%))`
}
