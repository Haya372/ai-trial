import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { EventResponse } from '../../../api/generated'
import { deleteEvent } from '../../../api/generated'
import { runEventMutation } from '../runEventMutation'

export function useEventDelete(
  event: EventResponse | null,
  onDeleted: () => void,
) {
  const queryClient = useQueryClient()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!event) return
    setIsDeleting(true)
    await runEventMutation({
      queryClient,
      request: deleteEvent(event.id),
      expectedStatus: 204,
      mode: 'delete',
      successMessage: '予定を削除しました',
      onSuccess: onDeleted,
    })
    setIsDeleting(false)
  }

  return { handleDelete, isDeleting }
}
