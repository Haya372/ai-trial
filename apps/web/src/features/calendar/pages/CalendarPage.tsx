import { Button } from '@repo/ui'
import type { CalendarEvent } from '@repo/ui'
import { MonthCalendar, WeekCalendar } from '@repo/ui'
import { useState } from 'react'
import type { EventResponse } from '../../../api/generated'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { useCalendarStore } from '../../../store/calendarStore'
import EventFormModal from '../../event/components/EventFormModal'
import type { EventFormMode } from '../../event/types'
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

  const [selectedEvent, setSelectedEvent] = useState<EventResponse | null>(null)
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [formModalOpen, setFormModalOpen] = useState(false)
  const [formMode, setFormMode] = useState<EventFormMode>('create')
  const [initialStart, setInitialStart] = useState<Date | null>(null)

  const { startDate, endDate } = getViewDateRange(view, currentDate)
  const { data, isPending, isError } = useEventsQuery(startDate, endDate)

  const calendarEvents: CalendarEvent[] = (data?.events ?? []).map((e) => ({
    id: e.id,
    title: e.title,
    start: new Date(e.startAt),
    end: new Date(e.endAt),
  }))

  function handleEventClick(event: CalendarEvent) {
    const fullEvent = data?.events.find((e) => e.id === event.id) ?? null
    setSelectedEvent(fullEvent)
    setDetailModalOpen(true)
  }

  function handleCreateClick() {
    setSelectedEvent(null)
    setFormMode('create')
    setInitialStart(null)
    setFormModalOpen(true)
  }

  function handleDateClick(date: Date) {
    setSelectedEvent(null)
    setFormMode('create')
    setInitialStart(date)
    setFormModalOpen(true)
  }

  function handleEditClick(event: EventResponse) {
    setSelectedEvent(event)
    setFormMode('edit')
    setDetailModalOpen(false)
    setFormModalOpen(true)
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
        <div className="flex items-center gap-2">
          <Button variant="primary" onClick={handleCreateClick}>
            新規作成
          </Button>
          <CalendarViewTabs view={view} onViewChange={setView} />
        </div>
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
              onDateClick={handleDateClick}
              onDateChange={setCurrentDate}
            />
          ) : (
            <WeekCalendar
              events={calendarEvents}
              currentDate={currentDate}
              onEventClick={handleEventClick}
              onDateClick={handleDateClick}
              onDateChange={setCurrentDate}
            />
          )}
        </>
      )}

      <EventDetailModal
        open={detailModalOpen}
        event={selectedEvent}
        onClose={() => setDetailModalOpen(false)}
        onEdit={handleEditClick}
      />

      <EventFormModal
        open={formModalOpen}
        mode={formMode}
        event={selectedEvent}
        initialStart={initialStart}
        onClose={() => setFormModalOpen(false)}
      />
    </div>
  )
}
