import type { EventDisplayInfo } from '@fullcalendar/react'
import type { ReactElement } from 'react'
import { CalendarEventContent } from './CalendarEventContent'
import type { CalendarEvent } from './types'

// Shared eventContent renderer for MonthCalendar and WeekCalendar.
// Returns true (FullCalendar default rendering) for events without tags so that
// the month view's list-item dot style is preserved for untagged events.
export function renderEventContent(arg: EventDisplayInfo): ReactElement | true {
  const calendarEvent = arg.event.extendedProps.calendarEvent as
    | CalendarEvent
    | undefined
  const tagColors = calendarEvent?.tagColors
  if (!tagColors || tagColors.length === 0) {
    return true
  }
  return (
    <CalendarEventContent
      timeText={arg.timeText}
      title={arg.event.title}
      tagColors={tagColors}
      timeClass={arg.timeClass}
      titleClass={arg.titleClass}
    />
  )
}
