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
  // gotoDate自身が発火させるdatesSetをonDateChangeとして親に伝播させないためのフラグ
  const isProgrammaticNavRef = useRef(false)

  useEffect(() => {
    const api = calendarRef.current?.getApi()
    if (!api) return
    isProgrammaticNavRef.current = true
    api.gotoDate(currentDate)
  }, [currentDate])

  const fullCalendarEvents = events.map((e) => ({
    id: e.id,
    title: e.title,
    start: e.start,
    end: e.end,
    extendedProps: { calendarEvent: e },
  }))

  function handleEventClick(arg: EventClickInfo) {
    const calendarEvent = arg.event.extendedProps.calendarEvent as CalendarEvent
    onEventClick(calendarEvent)
  }

  const handleDateClick = createDateClickHandler(onDateClick)

  function handleDatesSet(info: DatesSetInfo) {
    if (isProgrammaticNavRef.current) {
      isProgrammaticNavRef.current = false
      return
    }
    onDateChange(info.view.currentStart)
  }

  return (
    <FullCalendar
      ref={calendarRef}
      plugins={[dayGridPlugin, interactionPlugin, classicThemePlugin]}
      initialView="dayGridMonth"
      initialDate={currentDate}
      events={fullCalendarEvents}
      eventClick={handleEventClick}
      dateClick={handleDateClick}
      datesSet={handleDatesSet}
      locale="ja"
      headerToolbar={false}
    />
  )
}
