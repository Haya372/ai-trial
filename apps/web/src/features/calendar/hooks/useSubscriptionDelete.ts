import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { deleteSubscription } from '../../../api/generated'
import { runEventMutation } from '../../event/runEventMutation'

export function useSubscriptionDelete(
  subscriptionId: string | null,
  onDeleted: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
  const queryClient = useQueryClient()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async () => {
    if (!subscriptionId) return
    setIsDeleting(true)
    await runEventMutation({
      queryClient,
      request: deleteSubscription(subscriptionId),
      expectedStatus: 204,
      mode: 'unsubscribe',
      successMessage: t('toast.removeFromCalendarSuccess'),
      t,
      onSuccess: onDeleted,
    })
    setIsDeleting(false)
  }

  return { handleDelete, isDeleting }
}
