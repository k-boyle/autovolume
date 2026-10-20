import { createConnectQueryKey } from '@connectrpc/connect-query'
import type { DescMethodUnary } from '@bufbuild/protobuf'
import type { QueryClient } from '@tanstack/react-query'
import { LibraryService } from '../gen/autovolume/v1/library_pb'
import { VolumeService } from '../gen/autovolume/v1/volume_pb'

export function allQueriesKey(method: DescMethodUnary) {
  return createConnectQueryKey({ schema: method, cardinality: undefined })
}

export function invalidateVolumeRuleQueries(queryClient: QueryClient) {
  void queryClient.invalidateQueries({ queryKey: allQueriesKey(LibraryService.method.listItems) })
  void queryClient.resetQueries({ queryKey: allQueriesKey(LibraryService.method.listSeasons) })
  void queryClient.invalidateQueries({ queryKey: allQueriesKey(VolumeService.method.listVolumeRules) })
}
