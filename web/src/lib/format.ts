import type { Episode } from '../gen/autovolume/v1/now_playing_pb'

const relative = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' })

export function timeAgo(date: Date): string {
  const seconds = Math.round((date.getTime() - Date.now()) / 1000)
  const units: [Intl.RelativeTimeFormatUnit, number][] = [
    ['day', 86_400],
    ['hour', 3_600],
    ['minute', 60],
  ]
  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size) return relative.format(Math.round(seconds / size), unit)
  }
  return 'just now'
}

export function episodeLabel(episode: Episode): string {
  return `S${episode.seasonNumber} · E${episode.episodeNumber} — ${episode.title}`
}
