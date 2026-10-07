import type { SetupStatus } from '../gen/autovolume/v1/setup_pb'

export const setupSteps = [
  { id: 'server', label: 'Plex server' },
  { id: 'player', label: 'Player' },
  { id: 'tv', label: 'TV connection' },
  { id: 'volume', label: 'Default volume' },
] as const

export type SetupStepId = (typeof setupSteps)[number]['id']

export function hasRequiredSetup(status: SetupStatus | undefined): boolean {
  return !!status?.server && !!status.player && status.tvConfigured
}

export function firstIncompleteStep(
  status: SetupStatus | undefined,
  defaultVolume: number | undefined,
): SetupStepId | null {
  if (!status?.server) return 'server'
  if (!status.player) return 'player'
  if (!status.tvConfigured) return 'tv'
  if (defaultVolume === undefined) return 'volume'
  return null
}
