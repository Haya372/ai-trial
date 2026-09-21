import FullCalendar from '@fullcalendar/react'
import dayGridPlugin from '@fullcalendar/daygrid'
import type { DatesSetArg, EventClickArg } from '@fullcalendar/core'
import { useEffect, useRef } from 'react'
import type { CalendarEvent } from './types'

interface MonthCalendarProps {
  events: CalendarEvent[]
  currentDate: Date
  onEventClick: (event: CalendarEvent) => void
  onDateChange: (date: Date) => void
}

export function MonthCalendar({
  events,
  currentDate,
  onEventClick,
  onDateChange,
}: MonthCalendarProps) {
  const calendarRef = useRef<FullCalendar>(null)
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

  function handleEventClick(arg: EventClickArg) {
    const calendarEvent = arg.event.extendedProps.calendarEvent as CalendarEvent
    onEventClick(calendarEvent)
  }

  function handleDatesSet(info: DatesSetArg) {
    if (isProgrammaticNavRef.current) {
      isProgrammaticNavRef.current = false
      return
    }
    onDateChange(info.view.currentStart)
  }

  return (
    <FullCalendar
      ref={calendarRef}
      plugins={[dayGridPlugin]}
      initialView="dayGridMonth"
      initialDate={currentDate}
      events={fullCalendarEvents}
      eventClick={handleEventClick}
      datesSet={handleDatesSet}
      locale="ja"
      headerToolbar={false}
    />
  )
}
