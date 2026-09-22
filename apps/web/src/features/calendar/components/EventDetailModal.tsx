import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@repo/ui'
import type { EventResponse } from '../../../api/generated'
import { pad } from '../../../lib/dateFormat'

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
  if (!event) return null

  return (
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
        </div>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onEdit(event)}>
            編集
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
