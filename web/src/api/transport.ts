import type { Transport } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

export const isMock = import.meta.env.VITE_MOCK === 'true'

export async function createTransport(): Promise<Transport> {
  if (isMock) {
    const { createMockTransport } = await import('../mock/transport')
    return createMockTransport()
  }
  return createConnectTransport({ baseUrl: '/api' })
}
