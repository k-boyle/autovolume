import { LibraryBrowser } from '../components/home/LibraryBrowser'
import { NowPlayingCard } from '../components/home/NowPlayingCard'
import { TvCard } from '../components/home/TvCard'

export function HomePage() {
  return (
    <>
      <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
        <NowPlayingCard />
        <TvCard />
      </div>
      <div className="mt-12">
        <LibraryBrowser />
      </div>
    </>
  )
}
