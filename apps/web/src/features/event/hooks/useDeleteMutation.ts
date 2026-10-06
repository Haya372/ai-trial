import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { runEventMutation } from '../runEventMutation'

// Shared by useEventDelete and useSubscriptionDelete: both are a single DELETE
// call keyed by an id, differing only in which id/endpoint/error-message mode
// they use. Kept here so a future fix to the delete flow (loading state,
// error handling, query invalidation) can't be applied to one and missed in
// the other.
export function useDeleteMutation(
  id: string | null,
  deleteFn: (id: string) => Promise<{ status: number; data: unknown }>,
  mode: 'delete' | 'unsubscribe',
  successMessage: string,
  onDeleted: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
  const queryClient = useQueryClient()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!id) return
    setIsDeleting(true)
    await runEventMutation({
      queryClient,
      request: deleteFn(id),
      expectedStatus: 204,
      mode,
      successMessage,
      t,
      onSuccess: onDeleted,
    })
    setIsDeleting(false)
  }

  return { handleDelete, isDeleting }
}
