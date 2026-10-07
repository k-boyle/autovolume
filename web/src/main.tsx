import { Code, ConnectError } from '@connectrpc/connect'
import { TransportProvider } from '@connectrpc/connect-query'
import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router'
import { shouldRetry } from './api/errors'
import { createTransport, isMock } from './api/transport'
import { MockBadge } from './components/MockBadge'
import './index.css'
import { createAppRouter } from './router'

const router = createAppRouter()

function navigateForErrorCode(err: unknown) {
  const { code } = ConnectError.from(err)
  const path = router.state.location.pathname
  if (code === Code.Unauthenticated && !path.startsWith('/login')) {
    void router.navigate('/login', { replace: true })
  } else if (code === Code.FailedPrecondition && path !== '/setup') {
    void router.navigate('/setup', { replace: true })
  }
}

const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: navigateForErrorCode }),
  mutationCache: new MutationCache({ onError: navigateForErrorCode }),
  defaultOptions: {
    queries: { retry: shouldRetry, refetchOnWindowFocus: false },
  },
})

const transport = await createTransport()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <TransportProvider transport={transport}>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        {isMock && <MockBadge />}
      </QueryClientProvider>
    </TransportProvider>
  </StrictMode>,
)
