import type { MessageInitShape } from '@bufbuild/protobuf'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { Code, ConnectError, createRouterTransport, type Transport } from '@connectrpc/connect'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import { LibraryService } from '../gen/autovolume/v1/library_pb'
import type { MediaItemSchema } from '../gen/autovolume/v1/media_pb'
import {
  NowPlayingService,
  PlaybackState,
  type WatchNowPlayingResponseSchema,
} from '../gen/autovolume/v1/now_playing_pb'
import { SetupService, type SetupStatusSchema } from '../gen/autovolume/v1/setup_pb'
import { TvPower, TvService, type TvStateSchema } from '../gen/autovolume/v1/tv_pb'
import { VolumeService } from '../gen/autovolume/v1/volume_pb'
import { catalog, findByTitle, initialRules, libraries, players, servers, type CatalogEntry } from './data'
import { mockStorageKey } from './storageKey'

const pskTheTvRejects = 'wrong'
const unreachableTvAddress = 'offline'

interface State {
  loggedIn: boolean
  pendingLoginId?: string
  loginPolls: number
  serverId?: string
  playerId?: string
  tvAddress?: string
  tvPsk?: string
  tvVolume: number
  defaultVolume?: number
  rules: Record<string, number>
}

const state: State = load()

function load(): State {
  const initial: State = { loggedIn: false, loginPolls: 0, tvVolume: 20, rules: initialRules }
  const saved = localStorage.getItem(mockStorageKey)
  return saved ? { ...initial, ...(JSON.parse(saved) as Partial<State>) } : initial
}

function commit() {
  localStorage.setItem(mockStorageKey, JSON.stringify(state))
  for (const listener of [...listeners]) listener()
}

const listeners = new Set<() => void>()

function waitForChange(signal: AbortSignal, timeoutMs: number): Promise<void> {
  return new Promise((resolve) => {
    const done = () => {
      clearTimeout(timer)
      listeners.delete(done)
      signal.removeEventListener('abort', done)
      resolve()
    }
    const timer = setTimeout(done, timeoutMs)
    listeners.add(done)
    signal.addEventListener('abort', done)
  })
}

const delay = (ms = 150 + Math.random() * 250) => new Promise((r) => setTimeout(r, ms))

const user = { username: 'demo', avatarUrl: '' }

function requireSession() {
  if (!state.loggedIn) throw new ConnectError('Not signed in.', Code.Unauthenticated)
}

function requireServer() {
  requireSession()
  if (!state.serverId) throw new ConnectError('No Plex server selected.', Code.FailedPrecondition)
}

function requirePlayer() {
  requireServer()
  if (!state.playerId) throw new ConnectError('No player selected.', Code.FailedPrecondition)
}

function checkVolume(volume: number) {
  if (!Number.isInteger(volume) || volume < 0 || volume > 100) {
    throw new ConnectError('Volume must be between 0 and 100.', Code.InvalidArgument)
  }
}

function toItem(entry: CatalogEntry): MessageInitShape<typeof MediaItemSchema> {
  return {
    ratingKey: entry.ratingKey,
    type: entry.type,
    title: entry.title,
    year: entry.year,
    thumbUrl: '',
    volume: state.rules[entry.ratingKey],
  }
}

function findItem(ratingKey: string): CatalogEntry {
  const entry = catalog.find((e) => e.ratingKey === ratingKey)
  if (!entry) throw new ConnectError(`No movie or show with ratingKey ${ratingKey}.`, Code.NotFound)
  return entry
}

function setupStatus(): MessageInitShape<typeof SetupStatusSchema> {
  const player = players.find((p) => p.id === state.playerId)
  return {
    server: servers.find((s) => s.id === state.serverId),
    player: player && toPlayer(player),
    plexConnected: !!state.serverId,
    tvConfigured: !!state.tvAddress,
  }
}

function toPlayer(p: (typeof players)[number]) {
  return {
    id: p.id,
    name: p.name,
    product: p.product,
    lastSeen:
      p.lastSeenMinutesAgo === undefined
        ? undefined
        : timestampFromDate(new Date(Date.now() - p.lastSeenMinutesAgo * 60_000)),
  }
}

function tvState(): MessageInitShape<typeof TvStateSchema> {
  if (!state.tvAddress) return { power: TvPower.UNSPECIFIED }
  return { power: TvPower.ACTIVE, volume: state.tvVolume, muted: false }
}

const severance = findByTitle('Severance')
const scenes = [
  { entry: severance, state: PlaybackState.PLAYING, seconds: 45 },
  { entry: severance, state: PlaybackState.PAUSED, seconds: 15 },
  { entry: findByTitle('Dune: Part Two'), state: PlaybackState.PLAYING, seconds: 45 },
  { entry: undefined, state: PlaybackState.UNSPECIFIED, seconds: 20 },
]
const cycleSeconds = scenes.reduce((n, s) => n + s.seconds, 0)

function currentScene() {
  let t = (Date.now() / 1000) % cycleSeconds
  for (const scene of scenes) {
    if (t < scene.seconds) return scene
    t -= scene.seconds
  }
  return scenes[0]
}

let lastSceneKey = ''
function applyVolumeOnPlaybackChange() {
  const scene = currentScene()
  const key = `${scene.entry?.ratingKey}:${scene.state}`
  if (key === lastSceneKey) return
  lastSceneKey = key
  if (!state.tvAddress) return
  const target =
    scene.state === PlaybackState.PLAYING && scene.entry
      ? (state.rules[scene.entry.ratingKey] ?? state.defaultVolume)
      : state.defaultVolume
  if (target !== undefined && target !== state.tvVolume) {
    state.tvVolume = target
    commit()
  }
}

function snapshot(): MessageInitShape<typeof WatchNowPlayingResponseSchema> {
  const scene = currentScene()
  return {
    playback: scene.entry && {
      state: scene.state,
      item: toItem(scene.entry),
      episode:
        scene.entry === severance ? { title: 'Good News About Hell', seasonNumber: 1, episodeNumber: 1 } : undefined,
    },
    tv: tvState(),
    defaultVolume: state.defaultVolume,
  }
}

export function createMockTransport(): Transport {
  return createRouterTransport(({ service }) => {
    service(AuthService, {
      async startLogin() {
        await delay()
        state.pendingLoginId = crypto.randomUUID()
        state.loginPolls = 0
        commit()
        return { authUrl: 'about:blank', loginId: state.pendingLoginId }
      },
      async completeLogin({ loginId }) {
        await delay()
        if (!loginId || loginId !== state.pendingLoginId) throw new ConnectError('No sign-in in progress.', Code.NotFound)
        state.loginPolls++
        if (state.loginPolls < 2) {
          commit()
          return { complete: false }
        }
        state.pendingLoginId = undefined
        state.loggedIn = true
        commit()
        return { complete: true, user }
      },
      async getSession() {
        await delay()
        requireSession()
        return { user }
      },
      async logout() {
        await delay()
        state.loggedIn = false
        commit()
        return {}
      },
    })

    service(SetupService, {
      async listServers() {
        await delay()
        requireSession()
        return { servers }
      },
      async selectServer({ serverId }) {
        await delay(600)
        requireSession()
        const server = servers.find((s) => s.id === serverId)
        if (!server) throw new ConnectError('Unknown server.', Code.NotFound)
        if (!server.owned) {
          throw new ConnectError('Only the server owner gets playback events, so a shared server can’t be used.', Code.PermissionDenied)
        }
        if (state.serverId !== serverId) state.playerId = undefined
        state.serverId = serverId
        commit()
        return { status: setupStatus() }
      },
      async listPlayers() {
        await delay()
        requireServer()
        return { players: players.map(toPlayer) }
      },
      async selectPlayer({ playerId }) {
        await delay()
        requireServer()
        if (!players.some((p) => p.id === playerId)) throw new ConnectError('Unknown player.', Code.NotFound)
        state.playerId = playerId
        commit()
        return { status: setupStatus() }
      },
      async getSetupStatus() {
        await delay()
        requireSession()
        return { status: setupStatus() }
      },
    })

    service(TvService, {
      async getTvConfig() {
        await delay()
        requireSession()
        return { config: { address: state.tvAddress ?? '', pskSet: !!state.tvPsk } }
      },
      async updateTvConfig({ address, psk }) {
        await delay()
        requireSession()
        if (!address) throw new ConnectError('Enter the TV’s IP address.', Code.InvalidArgument)
        if (!psk && !state.tvPsk) throw new ConnectError('Enter the pre-shared key.', Code.InvalidArgument)
        state.tvAddress = address
        if (psk) state.tvPsk = psk
        commit()
        return { config: { address, pskSet: true } }
      },
      async testTvConnection({ address, psk }) {
        await delay(800)
        requireSession()
        if (!address) throw new ConnectError('Enter the TV’s IP address.', Code.InvalidArgument)
        const key = psk ?? state.tvPsk
        if (!key) throw new ConnectError('Enter the pre-shared key.', Code.InvalidArgument)
        if (address === unreachableTvAddress) throw new ConnectError(`No response from ${address}.`, Code.Unavailable)
        if (key === pskTheTvRejects) throw new ConnectError('The TV rejected the pre-shared key.', Code.PermissionDenied)
        return { model: 'KD-55X90J', state: { power: TvPower.ACTIVE, volume: state.tvVolume, muted: false } }
      },
      async setTvVolume({ volume }) {
        await delay()
        requireSession()
        checkVolume(volume)
        if (!state.tvAddress) throw new ConnectError('No TV configured.', Code.FailedPrecondition)
        state.tvVolume = volume
        commit()
        return { state: tvState() }
      },
    })

    service(LibraryService, {
      async listLibraries() {
        await delay()
        requireServer()
        return { libraries }
      },
      async listItems({ libraryId, query, offset, limit }) {
        await delay()
        requireServer()
        const needle = query.toLowerCase()
        const matches = catalog
          .filter((e) => e.libraryId === libraryId && e.title.toLowerCase().includes(needle))
          .sort((a, b) => a.title.localeCompare(b.title))
        const size = Math.min(limit || 50, 200)
        return { items: matches.slice(offset, offset + size).map(toItem), total: matches.length }
      },
    })

    service(VolumeService, {
      async setVolumeRule({ ratingKey, volume }) {
        await delay()
        requireServer()
        checkVolume(volume)
        const entry = findItem(ratingKey)
        state.rules = { ...state.rules, [ratingKey]: volume }
        commit()
        return { item: toItem(entry) }
      },
      async clearVolumeRule({ ratingKey }) {
        await delay()
        requireServer()
        const entry = findItem(ratingKey)
        const { [ratingKey]: _removed, ...rest } = state.rules
        state.rules = rest
        commit()
        return { item: toItem(entry) }
      },
      async listVolumeRules() {
        await delay()
        requireServer()
        const items = catalog
          .filter((e) => state.rules[e.ratingKey] !== undefined)
          .sort((a, b) => a.title.localeCompare(b.title))
          .map(toItem)
        return { items }
      },
      async getDefaultVolume() {
        await delay()
        requireSession()
        return { volume: state.defaultVolume }
      },
      async setDefaultVolume({ volume }) {
        await delay()
        requireSession()
        checkVolume(volume)
        state.defaultVolume = volume
        commit()
        return { volume }
      },
    })

    service(NowPlayingService, {
      async *watchNowPlaying(_req, ctx) {
        requirePlayer()
        let last = ''
        while (!ctx.signal.aborted) {
          applyVolumeOnPlaybackChange()
          const next = snapshot()
          const json = JSON.stringify(next)
          if (json !== last) {
            last = json
            yield next
          }
          await waitForChange(ctx.signal, 1000)
        }
      },
    })
  })
}
