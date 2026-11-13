import type { CalendarEvent } from '../calendar/types'

const MINUTE_MS = 60_000

export type TimelineItem =
  | {
      type: 'event'
      event: CalendarEvent
      // Clipped to today's range.
      start: Date
      end: Date
      startsBeforeToday: boolean
      endsAfterToday: boolean
      isOngoing: boolean
    }
  | { type: 'gap'; minutes: number }
  | { type: 'now'; time: Date }

export interface Timeline {
  items: TimelineItem[]
  nextEvent: { event: CalendarEvent; minutesUntil: number } | null
  hasEvents: boolean
}

function startOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

function overlapsDay(e: CalendarEvent, dayStart: Date, dayEnd: Date): boolean {
  if (e.start >= dayEnd) return false
  if (e.start.getTime() === e.end.getTime()) return e.start >= dayStart
  return e.end > dayStart
}

export function buildTimeline(events: CalendarEvent[], now: Date): Timeline {
  // SPEC-007 compares against the current time truncated to the minute, so an
  // event starting this minute is treated as ongoing rather than upcoming.
  const nowMinute = new Date(Math.floor(now.getTime() / MINUTE_MS) * MINUTE_MS)
  const dayStart = startOfDay(now)
  // setDate instead of adding 24h so DST transitions still land on midnight.
  const dayEnd = new Date(dayStart)
  dayEnd.setDate(dayEnd.getDate() + 1)

  const todays = events
    .filter((e) => overlapsDay(e, dayStart, dayEnd))
    .sort(
      (a, b) =>
        a.start.getTime() - b.start.getTime() ||
        a.end.getTime() - b.end.getTime(),
    )

  const items: TimelineItem[] = []
  let maxEnd: Date | null = null
  let nowPlaced = false

  for (const e of todays) {
    const start = e.start < dayStart ? dayStart : e.start
    const end = e.end > dayEnd ? dayEnd : e.end
    const gapMinutes =
      maxEnd === null
        ? 0
        : Math.floor((start.getTime() - maxEnd.getTime()) / MINUTE_MS)

    if (!nowPlaced && e.start > nowMinute) {
      // When a preceding event is still running, the indicator belongs before
      // the gap; otherwise the current time falls inside (or after) the gap.
      const isPrecedingOngoing = maxEnd !== null && nowMinute < maxEnd
      if (gapMinutes > 0 && !isPrecedingOngoing) {
        items.push({ type: 'gap', minutes: gapMinutes })
        items.push({ type: 'now', time: nowMinute })
      } else {
        items.push({ type: 'now', time: nowMinute })
        if (gapMinutes > 0) items.push({ type: 'gap', minutes: gapMinutes })
      }
      nowPlaced = true
    } else if (gapMinutes > 0) {
      items.push({ type: 'gap', minutes: gapMinutes })
    }

    items.push({
      type: 'event',
      event: e,
      start,
      end,
      startsBeforeToday: e.start < dayStart,
      endsAfterToday: e.end > dayEnd,
      isOngoing: e.start <= nowMinute && nowMinute < e.end,
    })
    if (maxEnd === null || end > maxEnd) maxEnd = end
  }

  if (!nowPlaced) items.push({ type: 'now', time: nowMinute })

  const next = todays.find((e) => e.start > nowMinute)
  return {
    items,
    nextEvent: next
      ? {
          event: next,
          minutesUntil: Math.ceil(
            (next.start.getTime() - nowMinute.getTime()) / MINUTE_MS,
          ),
        }
      : null,
    hasEvents: todays.length > 0,
  }
}
