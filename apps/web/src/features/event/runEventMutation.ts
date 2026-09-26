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
  mode: EventFormMode | 'delete'
  successMessage: string
  t: TFunction<['event', 'common']>
  onSuccess: (data: unknown) => void
}

export async function runEventMutation({
  queryClient,
  request,
  expectedStatus,
  mode,
  successMessage,
  t,
  onSuccess,
}: RunEventMutationParams): Promise<void> {
  try {
    const res = await request
    if (res.status !== expectedStatus) {
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
