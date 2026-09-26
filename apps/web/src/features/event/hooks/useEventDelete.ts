import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { deleteEvent } from '../../../api/generated'
import { runEventMutation } from '../runEventMutation'

export function useEventDelete(
  event: EventResponse | null,
  onDeleted: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
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
      successMessage: t('toast.deleteSuccess'),
      t,
      onSuccess: onDeleted,
    })
    setIsDeleting(false)
  }

  return { handleDelete, isDeleting }
}
