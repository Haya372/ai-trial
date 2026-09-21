import type { CalendarEvent } from '@repo/ui'
import { MonthCalendar, WeekCalendar } from '@repo/ui'
import { useState } from 'react'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { useCalendarStore } from '../../../store/calendarStore'
import CalendarNavigation from '../components/CalendarNavigation'
import CalendarViewTabs from '../components/CalendarViewTabs'
import EventDetailModal from '../components/EventDetailModal'
import { CALENDAR_VIEW, type CalendarView } from '../constants'
import { navigateDate } from '../navigateDate'

function getViewDateRange(
  view: CalendarView,
  date: Date,
): { startDate: Date; endDate: Date } {
  if (view === CALENDAR_VIEW.MONTH) {
    const startDate = new Date(date.getFullYear(), date.getMonth(), 1)
    const endDate = new Date(date.getFullYear(), date.getMonth() + 1, 0)
    return { startDate, endDate }
  }
  // 週ビュー: 日曜始まり
  const dayOfWeek = date.getDay()
  const startDate = new Date(date)
  startDate.setDate(date.getDate() - dayOfWeek)
  const endDate = new Date(startDate)
  endDate.setDate(startDate.getDate() + 6)
  return { startDate, endDate }
}

export default function CalendarPage() {
  const view = useCalendarStore((s) => s.view)
  const currentDate = useCalendarStore((s) => s.currentDate)
  const setView = useCalendarStore((s) => s.setView)
  const setCurrentDate = useCalendarStore((s) => s.setCurrentDate)

  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null)
  const [modalOpen, setModalOpen] = useState(false)

  const { startDate, endDate } = getViewDateRange(view, currentDate)
  const { data, isPending, isError } = useEventsQuery(startDate, endDate)

  const calendarEvents: CalendarEvent[] = (data?.events ?? []).map((e) => ({
    id: e.id,
    title: e.title,
    start: new Date(e.startAt),
    end: new Date(e.endAt),
  }))

  function handleEventClick(event: CalendarEvent) {
    setSelectedEvent(event)
    setModalOpen(true)
  }

  function handlePrev() {
    setCurrentDate(navigateDate(view, currentDate, -1))
  }

  function handleNext() {
    setCurrentDate(navigateDate(view, currentDate, 1))
  }

  function handleToday() {
    setCurrentDate(new Date())
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <CalendarNavigation
          view={view}
          currentDate={currentDate}
          onPrev={handlePrev}
          onNext={handleNext}
          onToday={handleToday}
        />
        <CalendarViewTabs view={view} onViewChange={setView} />
      </div>

      {isPending && (
        <div className="flex items-center justify-center p-16 text-muted-foreground">
          読み込み中...
        </div>
      )}

      {isError && (
        <div className="flex items-center justify-center p-16 text-destructive">
          予定を読み込めませんでした
        </div>
      )}

      {!isPending && !isError && (
        <>
          {view === CALENDAR_VIEW.MONTH ? (
            <MonthCalendar
              events={calendarEvents}
              currentDate={currentDate}
              onEventClick={handleEventClick}
              onDateChange={setCurrentDate}
            />
          ) : (
            <WeekCalendar
              events={calendarEvents}
              currentDate={currentDate}
              onEventClick={handleEventClick}
              onDateChange={setCurrentDate}
            />
          )}
        </>
      )}

      <EventDetailModal
        open={modalOpen}
        event={selectedEvent}
        onClose={() => setModalOpen(false)}
      />
    </div>
  )
}
