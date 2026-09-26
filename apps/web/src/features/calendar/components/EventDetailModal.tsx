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
import { useState } from 'react'
import type { EventResponse } from '../../../api/generated'
import { pad } from '../../../lib/dateFormat'
import { useEventDelete } from '../../event/hooks/useEventDelete'

interface EventDetailModalProps {
  open: boolean
  event: EventResponse | null
  onClose: () => void
  onEdit: (event: EventResponse) => void
}

function formatTime(date: Date): string {
  const h = pad(date.getHours())
  const m = pad(date.getMinutes())
  return `${h}:${m}`
}

function formatDateTime(iso: string): string {
  const date = new Date(iso)
  const y = date.getFullYear()
  const mo = date.getMonth() + 1
  const d = date.getDate()
  const time = formatTime(date)
  return `${y}年${mo}月${d}日 ${time}`
}

export default function EventDetailModal({
  open,
  event,
  onClose,
  onEdit,
}: EventDetailModalProps) {
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
            <div>
              <span className="text-muted-foreground">開始: </span>
              <span>{formatDateTime(event.startAt)}</span>
            </div>
            <div>
              <span className="text-muted-foreground">終了: </span>
              <span>{formatDateTime(event.endAt)}</span>
            </div>
            {event.description && (
              <div>
                <span className="text-muted-foreground">メモ: </span>
                <span>{event.description}</span>
              </div>
            )}
            {event.location && (
              <div>
                <span className="text-muted-foreground">場所: </span>
                <span>{event.location}</span>
              </div>
            )}
            {event.url && (
              <div>
                <span className="text-muted-foreground">URL: </span>
                <a
                  href={event.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary underline"
                >
                  {event.url}
                </a>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
              削除
            </Button>
            <Button variant="secondary" onClick={() => onEdit(event)}>
              編集
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>この予定を削除しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              この操作は取り消せません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>
              キャンセル
            </AlertDialogCancel>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? '削除中…' : '削除'}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
