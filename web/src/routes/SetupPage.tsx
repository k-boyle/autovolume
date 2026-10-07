import { useMutation, useQuery } from '@connectrpc/connect-query'
import { useQueryClient } from '@tanstack/react-query'
import { Check } from 'lucide-react'
import { Link, useNavigate, useSearchParams } from 'react-router'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { DefaultVolumeForm } from '../components/DefaultVolumeForm'
import { Logo } from '../components/Logo'
import { PlayerStep } from '../components/setup/PlayerStep'
import { ServerStep } from '../components/setup/ServerStep'
import { StepHeading } from '../components/setup/StepHeading'
import { FullPageSpinner } from '../components/Spinner'
import { TvConnectionForm } from '../components/TvConnectionForm'
import { AuthService } from '../gen/autovolume/v1/auth_pb'
import { SetupService } from '../gen/autovolume/v1/setup_pb'
import { TvPower } from '../gen/autovolume/v1/tv_pb'
import { VolumeService } from '../gen/autovolume/v1/volume_pb'
import { useNowPlayingStream } from '../hooks/useNowPlayingStream'
import { firstIncompleteStep, hasRequiredSetup, setupSteps, type SetupStepId } from '../lib/setup'

export function SetupPage() {
  const [params, setParams] = useSearchParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const status = useQuery(SetupService.method.getSetupStatus)
  const defaults = useQuery(VolumeService.method.getDefaultVolume)
  const logout = useMutation(AuthService.method.logout, {
    onSuccess: () => {
      queryClient.clear()
      void navigate('/login', { replace: true })
    },
  })

  if (status.isPending || defaults.isPending) return <FullPageSpinner />

  const setup = status.data?.status
  const firstIncomplete = firstIncompleteStep(setup, defaults.data?.volume)
  const reachable = (id: SetupStepId) =>
    firstIncomplete === null || stepIndex(id) <= stepIndex(firstIncomplete)
  const requested = params.get('step') as SetupStepId | null
  const current: SetupStepId =
    requested && setupSteps.some((s) => s.id === requested) && reachable(requested)
      ? requested
      : (firstIncomplete ?? 'server')
  const returnTo = params.get('return') ?? '/'

  async function stepDone() {
    const [s, d] = await Promise.all([status.refetch(), defaults.refetch()])
    const next = firstIncompleteStep(s.data?.status, d.data?.volume)
    if (next && next !== current) setParams({ step: next, return: returnTo }, { replace: true })
    else void navigate(returnTo, { replace: true })
  }

  return (
    <div className="min-h-dvh">
      <header className="mx-auto flex h-16 max-w-2xl items-center justify-between px-4">
        <Logo />
        <div className="flex items-center gap-1">
          {hasRequiredSetup(setup) && (
            <Link to={returnTo} className="rounded-lg px-3 py-2 text-sm text-muted hover:bg-surface-2 hover:text-fg">
              Back to app
            </Link>
          )}
          <Button variant="ghost" size="sm" onClick={() => logout.mutate({})} loading={logout.isPending}>
            Sign out
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-2xl px-4 pt-4 pb-16">
        <h1 className="text-2xl font-semibold tracking-tight">Set up AutoVolume</h1>
        <ol className="mt-6 flex flex-wrap gap-2">
          {setupSteps.map((step, i) => {
            const done = firstIncomplete === null || i < stepIndex(firstIncomplete)
            const active = step.id === current
            return (
              <li key={step.id}>
                <button
                  type="button"
                  disabled={!reachable(step.id)}
                  onClick={() => setParams({ step: step.id, return: returnTo })}
                  aria-current={active ? 'step' : undefined}
                  className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                    active ? 'border-accent bg-surface' : 'border-border hover:border-muted'
                  }`}
                >
                  <span
                    className={`grid size-5 shrink-0 place-items-center rounded-full text-xs font-semibold ${
                      done ? 'bg-accent text-on-accent' : 'bg-surface-2 text-muted'
                    }`}
                  >
                    {done ? <Check aria-hidden className="size-3" /> : i + 1}
                  </span>
                  <span className={active ? 'font-medium' : 'text-muted'}>{step.label}</span>
                </button>
              </li>
            )
          })}
        </ol>

        <Card className="mt-6">
          {current === 'server' && <ServerStep selectedId={setup?.server?.id} onDone={stepDone} />}
          {current === 'player' && <PlayerStep selectedId={setup?.player?.id} onDone={stepDone} />}
          {current === 'tv' && (
            <>
              <StepHeading
                title="Connect to your Bravia"
                description="AutoVolume uses the TV's IP control API to read and set its volume."
              />
              <TvConnectionForm onSaved={stepDone} />
            </>
          )}
          {current === 'volume' && <DefaultVolumeStep onDone={stepDone} />}
        </Card>
      </main>
    </div>
  )
}

function DefaultVolumeStep({ onDone }: { onDone: () => void }) {
  const { snapshot } = useNowPlayingStream()
  const tv = snapshot?.tv
  const tvVolume = tv?.power === TvPower.ACTIVE ? tv.volume : undefined
  return (
    <>
      <StepHeading
        title="Pick a default volume"
        description="The TV goes back to this when playback pauses or stops. Shows and movies without their own volume play at it too."
      />
      <DefaultVolumeForm tvVolume={tvVolume} onSaved={onDone} saveLabel="Save and finish" />
      <button type="button" onClick={onDone} className="mt-4 text-sm text-muted underline-offset-4 hover:text-fg hover:underline">
        Skip for now
      </button>
    </>
  )
}

function stepIndex(id: SetupStepId): number {
  return setupSteps.findIndex((s) => s.id === id)
}
