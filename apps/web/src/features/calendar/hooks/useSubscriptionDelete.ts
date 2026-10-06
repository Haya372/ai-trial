import { useTranslation } from 'react-i18next'
import { deleteSubscription } from '../../../api/generated'
import { useDeleteMutation } from '../../event/hooks/useDeleteMutation'

export function useSubscriptionDelete(
  subscriptionId: string | null,
  onDeleted: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
  return useDeleteMutation(
    subscriptionId,
    deleteSubscription,
    'unsubscribe',
    t('toast.removeFromCalendarSuccess'),
    onDeleted,
  )
}
