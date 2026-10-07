import { useInfiniteQuery, useQuery } from '@connectrpc/connect-query'
import { Search, Volume2 } from 'lucide-react'
import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { errorMessage } from '../../api/errors'
import { LibraryService } from '../../gen/autovolume/v1/library_pb'
import type { MediaItem } from '../../gen/autovolume/v1/media_pb'
import { VolumeService } from '../../gen/autovolume/v1/volume_pb'
import { useDebounced } from '../../hooks/useDebounced'
import { Button } from '../Button'
import { Notice } from '../Notice'
import { Poster } from '../Poster'
import { ItemVolumeDialog } from './ItemVolumeDialog'

const pageSize = 48
const savedTab = 'saved'

export function LibraryBrowser() {
  const libraries = useQuery(LibraryService.method.listLibraries)
  const rules = useQuery(VolumeService.method.listVolumeRules)
  const [params, setParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const query = useDebounced(search.trim(), 250)
  const [editing, setEditing] = useState<MediaItem | null>(null)

  const tabs = [
    ...(libraries.data?.libraries.map((l) => ({ id: l.id, label: l.title })) ?? []),
    { id: savedTab, label: `Saved volumes${rules.data ? ` (${rules.data.items.length})` : ''}` },
  ]
  const active = params.get('library') ?? libraries.data?.libraries[0]?.id ?? (libraries.isPending ? undefined : savedTab)

  return (
    <section aria-labelledby="library-heading">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <h2 id="library-heading" className="text-xl font-semibold tracking-tight">
          Library
        </h2>
        <label className="relative w-full sm:w-72">
          <span className="sr-only">Search titles</span>
          <Search aria-hidden className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted" />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search titles"
            className="h-10 w-full rounded-lg border border-border bg-surface pr-3 pl-9 text-sm placeholder:text-muted focus:border-accent focus:outline-none"
          />
        </label>
      </div>

      <div role="tablist" aria-label="Libraries" className="mt-4 flex gap-1 overflow-x-auto shadow-[inset_0_-1px_0_var(--color-border)] [scrollbar-width:none]">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={tab.id === active}
            onClick={() => setParams({ library: tab.id }, { replace: true })}
            className={`border-b-2 px-3 py-2.5 text-sm font-medium whitespace-nowrap transition-colors ${
              tab.id === active ? 'border-accent text-fg' : 'border-transparent text-muted hover:text-fg'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="mt-6" role="tabpanel">
        {libraries.isError && <Notice tone="error">{errorMessage(libraries.error)}</Notice>}
        {active === undefined ? (
          <GridSkeleton />
        ) : active === savedTab ? (
          <SavedGrid query={query} onEdit={setEditing} />
        ) : (
          <LibraryGrid key={active} libraryId={active} query={query} onEdit={setEditing} />
        )}
      </div>

      <ItemVolumeDialog item={editing} onClose={() => setEditing(null)} />
    </section>
  )
}

interface GridProps {
  query: string
  onEdit: (item: MediaItem) => void
}

function LibraryGrid({ libraryId, query, onEdit }: GridProps & { libraryId: string }) {
  const items = useInfiniteQuery(
    LibraryService.method.listItems,
    { libraryId, query, offset: 0, limit: pageSize },
    {
      pageParamKey: 'offset',
      getNextPageParam: (last, pages) => {
        const loaded = pages.reduce((n, page) => n + page.items.length, 0)
        return loaded < last.total ? loaded : undefined
      },
    },
  )

  if (items.isPending) return <GridSkeleton />
  if (items.isError) return <Notice tone="error">{errorMessage(items.error)}</Notice>

  const all = items.data.pages.flatMap((page) => page.items)
  if (all.length === 0) {
    return <EmptyState>{query ? `Nothing matches “${query}”.` : 'This library is empty.'}</EmptyState>
  }

  return (
    <>
      <Grid items={all} onEdit={onEdit} />
      {items.hasNextPage && (
        <LoadMoreOnScroll
          loading={items.isFetchingNextPage}
          onLoad={() => !items.isFetchingNextPage && void items.fetchNextPage()}
        />
      )}
    </>
  )
}

function SavedGrid({ query, onEdit }: GridProps) {
  const rules = useQuery(VolumeService.method.listVolumeRules)
  if (rules.isPending) return <GridSkeleton />
  if (rules.isError) return <Notice tone="error">{errorMessage(rules.error)}</Notice>

  const needle = query.toLowerCase()
  const items = rules.data.items.filter((item) => item.title.toLowerCase().includes(needle))
  if (items.length === 0) {
    return (
      <EmptyState>
        {query ? `No saved volumes match “${query}”.` : 'No saved volumes yet. Pick a show or movie to give it one.'}
      </EmptyState>
    )
  }
  return <Grid items={items} onEdit={onEdit} />
}

const gridClass = 'grid grid-cols-3 gap-x-4 gap-y-6 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6'

function Grid({ items, onEdit }: { items: MediaItem[]; onEdit: (item: MediaItem) => void }) {
  return (
    <ul className={gridClass}>
      {items.map((item) => (
        <li key={item.ratingKey}>
          <button
            type="button"
            onClick={() => onEdit(item)}
            className="group block w-full rounded-lg text-left focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-accent"
          >
            <div className="relative transition-transform group-hover:-translate-y-0.5">
              <Poster title={item.title} thumbUrl={item.thumbUrl} displayWidth={160} />
              {item.volume !== undefined && (
                <span className="absolute top-2 right-2 flex items-center gap-1 rounded-full bg-accent px-2 py-0.5 text-xs font-semibold text-on-accent shadow">
                  <Volume2 aria-hidden className="size-3" />
                  <span className="sr-only">Saved volume</span>
                  {item.volume}
                </span>
              )}
            </div>
            <p className="mt-2 truncate text-sm font-medium">{item.title}</p>
            {item.year > 0 && <p className="text-xs text-muted">{item.year}</p>}
          </button>
        </li>
      ))}
    </ul>
  )
}

function LoadMoreOnScroll({ loading, onLoad }: { loading: boolean; onLoad: () => void }) {
  const ref = useRef<HTMLDivElement>(null)
  const onLoadRef = useRef(onLoad)
  useEffect(() => {
    onLoadRef.current = onLoad
  })

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const observer = new IntersectionObserver(([entry]) => entry.isIntersecting && onLoadRef.current(), {
      rootMargin: '400px',
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return (
    <div ref={ref} className="mt-8 flex justify-center">
      <Button onClick={onLoad} loading={loading}>
        Load more
      </Button>
    </div>
  )
}

function GridSkeleton() {
  return (
    <div className={gridClass} aria-busy aria-label="Loading">
      {Array.from({ length: 12 }, (_, i) => (
        <div key={i}>
          <div className="aspect-[2/3] animate-pulse rounded-lg bg-surface-2" />
          <div className="mt-2 h-3.5 w-3/4 animate-pulse rounded bg-surface-2" />
        </div>
      ))}
    </div>
  )
}

function EmptyState({ children }: { children: ReactNode }) {
  return <p className="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-muted">{children}</p>
}
