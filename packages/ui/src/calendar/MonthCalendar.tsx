import FullCalendar from '@fullcalendar/react'
import dayGridPlugin from '@fullcalendar/react/daygrid'
import interactionPlugin from '@fullcalendar/react/interaction'
import classicThemePlugin from '@fullcalendar/react/themes/classic'
import type {
  CalendarRef,
  DatesSetInfo,
  EventClickInfo,
} from '@fullcalendar/react'
import { useEffect, useRef } from 'react'
import { createDateClickHandler } from './createDateClickHandler'
import { eventClassNames } from './eventClassNames'
import { eventColorProps } from './eventColorProps'
import { renderEventContent } from './renderEventContent'
import { shouldNotifyDateChange } from './shouldNotifyDateChange'
import type { CalendarEvent } from './types'

import '@fullcalendar/react/skeleton.css'
import '@fullcalendar/react/themes/classic/theme.css'
import '@fullcalendar/react/themes/classic/palette.css'

interface MonthCalendarProps {
  events: CalendarEvent[]
  currentDate: Date
  onEventClick: (event: CalendarEvent) => void
  onDateClick: (date: Date) => void
  onDateChange: (date: Date) => void
}

export function MonthCalendar({
  events,
  currentDate,
  onEventClick,
  onDateClick,
  onDateChange,
}: MonthCalendarProps) {
  const calendarRef = useRef<CalendarRef>(null)
  // datesSetハンドラから最新のcurrentDateを参照するためのref
  const currentDateRef = useRef(currentDate)

  useEffect(() => {
    // gotoDateはその場でdatesSetを発火させるため、判定に使うrefを先に更新しておく
    currentDateRef.current = currentDate
    const api = calendarRef.current?.getApi()
    if (!api) return
    api.gotoDate(currentDate)
  }, [currentDate])

  const fullCalendarEvents = events.map((e) => ({
    id: e.id,
    title: e.title,
    start: e.start,
    end: e.end,
    classNames: eventClassNames(e),
    ...eventColorProps(e),
    // dayGrid shows timed events as "list-item" (dot + text) by default.
    // Switch to "block" for tagged events so the tag background color is visible.
    ...(e.tagColors?.length ? { display: 'block' } : {}),
    extendedProps: { calendarEvent: e },
  }))

  function handleEventClick(arg: EventClickInfo) {
    const calendarEvent = arg.event.extendedProps.calendarEvent as CalendarEvent
    onEventClick(calendarEvent)
  }

  const handleDateClick = createDateClickHandler(onDateClick)

  function handleDatesSet(info: DatesSetInfo) {
    if (!shouldNotifyDateChange(currentDateRef.current, info)) return
    onDateChange(info.view.currentStart)
  }

  return (
    <FullCalendar
      ref={calendarRef}
      plugins={[dayGridPlugin, interactionPlugin, classicThemePlugin]}
      initialView="dayGridMonth"
      initialDate={currentDate}
      events={fullCalendarEvents}
      eventContent={renderEventContent}
      eventClick={handleEventClick}
      dateClick={handleDateClick}
      datesSet={handleDatesSet}
      locale="ja"
      headerToolbar={false}
      height="100%"
      expandRows={true}
    />
  )
}
