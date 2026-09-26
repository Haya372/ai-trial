import { useQueryClient } from '@tanstack/react-query'
import { toast } from '@repo/ui'
import { useState } from 'react'
import type { EventResponse } from '../../../api/generated'
import { deleteEvent } from '../../../api/generated'
import { eventsKeys } from '../../../lib/queryKeys'
import { getEventErrorMessage } from '../utils'

export function useEventDelete(
  event: EventResponse | null,
  onDeleted: () => void,
) {
  const queryClient = useQueryClient()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!event) return
    setIsDeleting(true)
    try {
      const res = await deleteEvent(event.id)
      if (res.status !== 204) {
        toast.error(getEventErrorMessage(res.data, 'delete'))
        return
      }
      await queryClient.invalidateQueries({ queryKey: eventsKeys.all })
      toast.success('予定を削除しました')
      onDeleted()
    } catch (error) {
      toast.error(getEventErrorMessage(error, 'delete'))
    } finally {
      setIsDeleting(false)
    }
  }

  return { handleDelete, isDeleting }
}
