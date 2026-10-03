import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@repo/ui'
import type { TFunction } from 'i18next'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { pad } from '../../../lib/dateFormat'
import { useEventDelete } from '../../event/hooks/useEventDelete'
import LabeledField from './LabeledField'

interface EventDetailModalProps {
  open: boolean
  event: EventResponse | null
  onClose: () => void
  onEdit: (event: EventResponse) => void
  onShare: (event: EventResponse) => void
}

function formatTime(date: Date): string {
  const h = pad(date.getHours())
  const m = pad(date.getMinutes())
  return `${h}:${m}`
}

function formatDateTime(t: TFunction<'calendar'>, iso: string): string {
  const date = new Date(iso)
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const time = formatTime(date)
  return t('eventDetail.dateTimeFormat', { year, month, day, time })
}

export default function EventDetailModal({
  open,
  event,
  onClose,
  onEdit,
  onShare,
}: EventDetailModalProps) {
  const { t } = useTranslation('calendar')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const { handleDelete, isDeleting } = useEventDelete(event, () => {
    setConfirmOpen(false)
    onClose()
  })

  if (!event) return null

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={(isOpen) => {
          if (!isOpen) onClose()
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{event.title}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-2 text-sm">
            <LabeledField label={t('eventDetail.startLabel')}>
              {formatDateTime(t, event.startAt)}
            </LabeledField>
            <LabeledField label={t('eventDetail.endLabel')}>
              {formatDateTime(t, event.endAt)}
            </LabeledField>
            {event.description && (
              <LabeledField label={t('eventDetail.noteLabel')}>
                {event.description}
              </LabeledField>
            )}
            {event.location && (
              <LabeledField label={t('eventDetail.locationLabel')}>
                {event.location}
              </LabeledField>
            )}
            {event.url && (
              <LabeledField label={t('eventDetail.urlLabel')}>
                <a
                  href={event.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary underline"
                >
                  {event.url}
                </a>
              </LabeledField>
            )}
          </div>
          <DialogFooter>
            <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
              {t('eventDetail.delete')}
            </Button>
            <Button variant="secondary" onClick={() => onEdit(event)}>
              {t('eventDetail.edit')}
            </Button>
            <Button variant="secondary" onClick={() => onShare(event)}>
              {t('eventDetail.share')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t('eventDetail.confirmDeleteTitle')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t('eventDetail.confirmDeleteDescription')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>
              {t('eventDetail.cancel')}
            </AlertDialogCancel>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? t('eventDetail.deleting') : t('eventDetail.delete')}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
