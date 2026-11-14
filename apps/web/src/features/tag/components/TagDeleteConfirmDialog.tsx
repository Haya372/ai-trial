import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
} from '@repo/ui'
import { useTranslation } from 'react-i18next'
import type { TagResponse } from '../../../api/generated'

interface TagDeleteConfirmDialogProps {
  tag: TagResponse | null
  isDeleting: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function TagDeleteConfirmDialog({
  tag,
  isDeleting,
  onConfirm,
  onCancel,
}: TagDeleteConfirmDialogProps) {
  const { t } = useTranslation('tag')
  return (
    <AlertDialog
      open={tag !== null}
      onOpenChange={(open) => {
        // Escape still requests a close while deleting; ignore it so the dialog
        // stays tied to the request until it settles.
        if (!open && !isDeleting) onCancel()
      }}
    >
      <AlertDialogContent>
        {tag && (
          <>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {t('deleteConfirm.title', { name: tag.name })}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {t('deleteConfirm.description', { count: tag.eventCount })}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={isDeleting}>
                {t('deleteConfirm.cancel')}
              </AlertDialogCancel>
              <Button
                variant="destructive"
                onClick={onConfirm}
                disabled={isDeleting}
              >
                {isDeleting
                  ? t('deleteConfirm.deleting')
                  : t('deleteConfirm.confirm')}
              </Button>
            </AlertDialogFooter>
          </>
        )}
      </AlertDialogContent>
    </AlertDialog>
  )
}
