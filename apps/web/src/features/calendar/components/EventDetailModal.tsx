import type { CalendarEvent } from '@repo/ui'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@repo/ui'

interface EventDetailModalProps {
  open: boolean
  event: CalendarEvent | null
  onClose: () => void
}

function formatTime(date: Date): string {
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  return `${h}:${m}`
}

function formatDateTime(date: Date): string {
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
            <span>{formatDateTime(event.start)}</span>
          </div>
          <div>
            <span className="text-muted-foreground">終了: </span>
            <span>{formatDateTime(event.end)}</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
