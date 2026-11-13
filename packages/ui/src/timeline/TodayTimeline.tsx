import { cn } from 'cn'
import type { CalendarEvent } from '../calendar/types'
import { buildTimeline, type TimelineItem } from './buildTimeline'
import { formatTime, formatTimeUntil } from './formatTimeline'
import type { TodayTimelineLabels } from './types'

interface TodayTimelineProps {
  events: CalendarEvent[]
  now: Date
  labels: TodayTimelineLabels
  onEventClick: (event: CalendarEvent) => void
}

type EventItem = Extract<TimelineItem, { type: 'event' }>

function timeRange(item: EventItem, labels: TodayTimelineLabels): string {
  const start = item.startsBeforeToday
    ? labels.fromPreviousDay
    : formatTime(item.start)
  if (item.event.start.getTime() === item.event.end.getTime()) return start
  const end = item.endsAfterToday ? labels.untilNextDay : formatTime(item.end)
  return `${start}〜${end}`
}

export function TodayTimeline({
  events,
  now,
  labels,
  onEventClick,
}: TodayTimelineProps) {
  const { items, nextEvent, hasEvents } = buildTimeline(events, now)

  return (
    <div data-slot="today-timeline" className="flex flex-col gap-3">
      <div className="rounded-md border border-border p-3">
        {nextEvent ? (
          <>
            <p className="truncate font-semibold" title={nextEvent.event.title}>
              {nextEvent.event.title}
            </p>
            <p className="text-sm text-muted-foreground">
              {formatTimeUntil(nextEvent.minutesUntil, labels)}
            </p>
          </>
        ) : (
          <p className="text-sm text-muted-foreground">
            {hasEvents ? labels.noMoreEventsToday : labels.noEventsToday}
          </p>
        )}
      </div>

      <ol className="flex flex-col gap-1">
        {items.map((item, index) => {
          if (item.type === 'now') {
            return (
              <li
                key="now"
                className="flex items-center gap-2 text-xs font-semibold text-destructive"
              >
                <span>{formatTime(item.time)}</span>
                <span className="h-px flex-1 bg-destructive" />
              </li>
            )
          }
          if (item.type === 'gap') {
            return (
              <li
                // A gap has no identity of its own; it is always followed by
                // the next event, whose id is unique.
                key={`gap-${index}`}
                className="px-3 py-1 text-xs text-muted-foreground"
              >
                {labels.freeTime(item.minutes)}
              </li>
            )
          }
          return (
            <li key={item.event.id}>
              <button
                type="button"
                onClick={() => onEventClick(item.event)}
                className={cn(
                  'flex w-full flex-col items-start rounded-md border-l-4 border-transparent px-3 py-2 text-left hover:bg-muted',
                  item.isOngoing && 'border-primary bg-primary/10',
                )}
              >
                <span className="text-xs text-muted-foreground">
                  {timeRange(item, labels)}
                </span>
                <span
                  className="w-full truncate text-sm"
                  title={item.event.title}
                >
                  {item.event.title}
                </span>
              </button>
            </li>
          )
        })}
      </ol>
    </div>
  )
}
