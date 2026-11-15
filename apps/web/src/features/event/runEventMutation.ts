import type { QueryClient } from '@tanstack/react-query'
import { toast } from '@repo/ui'
import type { TFunction } from 'i18next'
import { eventsKeys } from '../../lib/queryKeys'
import type { EventFormMode } from './types'
import { getEventErrorMessage } from './utils'

interface RunEventMutationParams {
  queryClient: QueryClient
  request: Promise<{ status: number; data: unknown }>
  expectedStatus: number
  mode: EventFormMode | 'delete' | 'unsubscribe'
  successMessage: string
  t: TFunction<['event', 'common']>
  onSuccess: (data: unknown) => void
  // Returns true when the caller has shown the error itself (e.g. as a field error).
  handleError?: (data: unknown) => boolean
}

export async function runEventMutation({
  queryClient,
  request,
  expectedStatus,
  mode,
  successMessage,
  t,
  onSuccess,
  handleError,
}: RunEventMutationParams): Promise<void> {
  try {
    const res = await request
    if (res.status !== expectedStatus) {
      if (handleError?.(res.data)) return
      toast.error(getEventErrorMessage(res.data, mode, t))
      return
    }
    await queryClient.invalidateQueries({ queryKey: eventsKeys.all })
    toast.success(successMessage)
    onSuccess(res.data)
  } catch (error) {
    toast.error(getEventErrorMessage(error, mode, t))
  }
}
