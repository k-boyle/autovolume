import { mockStorageKey } from '../mock/storageKey'

export function MockBadge() {
  function reset() {
    localStorage.removeItem(mockStorageKey)
    window.location.assign('/')
  }

  return (
    <div className="fixed right-4 bottom-4 z-50 flex items-center gap-2 rounded-full border border-border bg-surface py-1 pr-1 pl-3 text-xs text-muted shadow-lg">
      Mock backend
      <button type="button" onClick={reset} className="rounded-full bg-surface-2 px-2.5 py-1 font-medium text-fg hover:bg-border">
        Reset
      </button>
    </div>
  )
}
