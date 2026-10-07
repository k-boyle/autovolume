import { Code, ConnectError } from '@connectrpc/connect'

export function errorCode(err: unknown): Code | undefined {
  return err == null ? undefined : ConnectError.from(err).code
}

export function errorMessage(err: unknown): string {
  return ConnectError.from(err).rawMessage || 'Something went wrong.'
}

const nonRetryableCodes = new Set([
  Code.Unauthenticated,
  Code.FailedPrecondition,
  Code.PermissionDenied,
  Code.NotFound,
  Code.InvalidArgument,
])

export function shouldRetry(failureCount: number, err: unknown): boolean {
  return failureCount < 2 && !nonRetryableCodes.has(ConnectError.from(err).code)
}
