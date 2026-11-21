import FullCalendar from '@fullcalendar/react'
import interactionPlugin from '@fullcalendar/react/interaction'
import timeGridPlugin from '@fullcalendar/react/timegrid'
import classicThemePlugin from '@fullcalendar/react/themes/classic'
import type {
  CalendarRef,
  DateClickInfo,
  DatesSetInfo,
  EventClickInfo,
} from '@fullcalendar/react'
import { useEffect, useRef } from 'react'
import { extractDateClickResult } from './createDateClickHandler'
import { eventClassNames } from './eventClassNames'
import { eventColorProps } from './eventColorProps'
import { renderEventContent } from './renderEventContent'
import { shouldNotifyDateChange } from './shouldNotifyDateChange'
import type { CalendarEvent } from './types'

import '@fullcalendar/react/skeleton.css'
import '@fullcalendar/react/themes/classic/theme.css'
import '@fullcalendar/react/themes/classic/palette.css'

interface WeekCalendarProps {
  events: CalendarEvent[]
  currentDate: Date
  onEventClick: (event: CalendarEvent) => void
  onTimeSlotClick: (date: Date, anchor: { x: number; y: number }) => void
  onDateChange: (date: Date) => void
}

export function WeekCalendar({
  events,
  currentDate,
  onEventClick,
  onTimeSlotClick,
  onDateChange,
}: WeekCalendarProps) {
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
    extendedProps: { calendarEvent: e },
  }))

  function handleEventClick(arg: EventClickInfo) {
    const calendarEvent = arg.event.extendedProps.calendarEvent as CalendarEvent
    onEventClick(calendarEvent)
  }

  function handleDateClick(arg: DateClickInfo) {
    const { date, anchor } = extractDateClickResult(arg)
    onTimeSlotClick(date, anchor)
  }

  function handleDatesSet(info: DatesSetInfo) {
    if (!shouldNotifyDateChange(currentDateRef.current, info)) return
    onDateChange(info.view.currentStart)
  }

  return (
    <FullCalendar
      ref={calendarRef}
      plugins={[timeGridPlugin, interactionPlugin, classicThemePlugin]}
      initialView="timeGridWeek"
      initialDate={currentDate}
      events={fullCalendarEvents}
      eventContent={renderEventContent}
      eventClick={handleEventClick}
      dateClick={handleDateClick}
      datesSet={handleDatesSet}
      locale="ja"
      headerToolbar={false}
    />
  )
}
