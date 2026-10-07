import { useQuery } from '@connectrpc/connect-query'
import { Settings, Volume2, VolumeX } from 'lucide-react'
import { Link, NavLink, Outlet } from 'react-router'
import { Avatar } from '../components/Avatar'
import { Logo } from '../components/Logo'
import { NowPlayingProvider } from '../components/NowPlayingProvider'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import { TvPower } from '../gen/autovolume/v1/tv_pb'
import { useNowPlaying } from '../hooks/useNowPlaying'

export function AppShell() {
  const session = useQuery(AuthService.method.getSession)
  const user = session.data?.user

  return (
    <NowPlayingProvider>
      <div className="min-h-dvh">
        <header className="sticky top-0 z-10 border-b border-border bg-bg/85 backdrop-blur">
          <div className="mx-auto flex h-16 max-w-6xl items-center gap-3 px-4">
            <Link to="/" aria-label="AutoVolume home">
              <Logo />
            </Link>
            <div className="ml-auto flex items-center gap-1 sm:gap-3">
              <TvPill />
              <NavLink
                to="/settings"
                aria-label="Settings"
                className={({ isActive }) =>
                  `grid size-9 place-items-center rounded-lg hover:bg-surface-2 ${isActive ? 'text-fg' : 'text-muted'}`
                }
              >
                <Settings className="size-5" />
              </NavLink>
              {user && <Avatar name={user.username} url={user.avatarUrl} />}
            </div>
          </div>
        </header>
        <main className="mx-auto max-w-6xl px-4 pt-6 pb-16">
          <Outlet />
        </main>
      </div>
    </NowPlayingProvider>
  )
}

function TvPill() {
  const { snapshot } = useNowPlaying()
  const tv = snapshot?.tv
  if (!tv) return null
  const on = tv.power === TvPower.ACTIVE && tv.volume !== undefined
  const Icon = on && !tv.muted ? Volume2 : VolumeX
  return (
    <span className="flex h-8 items-center gap-1.5 rounded-full border border-border px-3 text-sm text-muted">
      <Icon aria-hidden className="size-4" />
      <span className="sr-only">TV volume</span>
      {on ? <span className="font-medium text-fg tabular-nums">{tv.volume}</span> : 'TV off'}
    </span>
  )
}
