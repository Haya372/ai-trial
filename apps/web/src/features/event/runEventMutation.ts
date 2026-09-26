import type { QueryClient } from '@tanstack/react-query'
import { toast } from '@repo/ui'
import { eventsKeys } from '../../lib/queryKeys'
import type { EventFormMode } from './types'
import { getEventErrorMessage } from './utils'

interface RunEventMutationParams {
  queryClient: QueryClient
  request: Promise<{ status: number; data: unknown }>
  expectedStatus: number
  mode: EventFormMode | 'delete'
  successMessage: string
  onSuccess: () => void
}

export async function runEventMutation({
  queryClient,
  request,
  expectedStatus,
  mode,
  successMessage,
  onSuccess,
}: RunEventMutationParams): Promise<void> {
  try {
    const res = await request
    if (res.status !== expectedStatus) {
      toast.error(getEventErrorMessage(res.data, mode))
      return
    }
    await queryClient.invalidateQueries({ queryKey: eventsKeys.all })
    toast.success(successMessage)
    onSuccess()
  } catch (error) {
    toast.error(getEventErrorMessage(error, mode))
  }
}
