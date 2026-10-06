import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { deleteEvent } from '../../../api/generated'
import { useDeleteMutation } from './useDeleteMutation'

export function useEventDelete(
  event: EventResponse | null,
  onDeleted: () => void,
) {
  const { t } = useTranslation(['event', 'common'])
  return useDeleteMutation(
    event?.id ?? null,
    deleteEvent,
    'delete',
    t('toast.deleteSuccess'),
    onDeleted,
  )
}
