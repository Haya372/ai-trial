import { Button } from '@repo/ui'
import type { CalendarEvent } from '@repo/ui'
import { MonthCalendar, WeekCalendar } from '@repo/ui'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { useCalendarStore } from '../../../store/calendarStore'
import EventFormModal from '../../event/components/EventFormModal'
import QuickRegistrationPanel from '../../event/components/QuickRegistrationPanel'
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
  const { t } = useTranslation('calendar')
  const view = useCalendarStore((s) => s.view)
  const currentDate = useCalendarStore((s) => s.currentDate)
  const setView = useCalendarStore((s) => s.setView)
  const setCurrentDate = useCalendarStore((s) => s.setCurrentDate)

  const [selectedEvent, setSelectedEvent] = useState<EventResponse | null>(null)
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [formModalOpen, setFormModalOpen] = useState(false)
  const [formMode, setFormMode] = useState<EventFormMode>('create')
  const [initialStart, setInitialStart] = useState<Date | null>(null)
  const [quickPanel, setQuickPanel] = useState<{
    key: number
    start: Date
    anchor: { x: number; y: number }
  } | null>(null)
  // 同じ時間帯セルを連続でクリックしても必ずパネルを再マウントし、
  // 前回の登録完了状態が残らないようにするためのクリック連番
  const quickPanelClickSeqRef = useRef(0)

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

  function openCreateForm(start: Date | null) {
    setSelectedEvent(null)
    setFormMode('create')
    setInitialStart(start)
    setFormModalOpen(true)
  }

  function handleCreateClick() {
    openCreateForm(null)
  }

  function handleDateClick(date: Date) {
    openCreateForm(date)
  }

  function handleTimeSlotClick(date: Date, anchor: { x: number; y: number }) {
    quickPanelClickSeqRef.current += 1
    setQuickPanel({ key: quickPanelClickSeqRef.current, start: date, anchor })
  }

  function handleEditClick(event: EventResponse) {
    setSelectedEvent(event)
    setFormMode('edit')
    setDetailModalOpen(false)
    setFormModalOpen(true)
  }

  function handleQuickPanelEditDetail(event: EventResponse) {
    setQuickPanel(null)
    handleEditClick(event)
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
            {t('page.createButton')}
          </Button>
          <CalendarViewTabs view={view} onViewChange={setView} />
        </div>
      </div>

      {isPending && (
        <div className="flex items-center justify-center p-16 text-muted-foreground">
          {t('page.loading')}
        </div>
      )}

      {isError && (
        <div className="flex items-center justify-center p-16 text-destructive">
          {t('page.loadError')}
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
              onTimeSlotClick={handleTimeSlotClick}
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

      {quickPanel && (
        <QuickRegistrationPanel
          key={quickPanel.key}
          open={true}
          anchor={quickPanel.anchor}
          start={quickPanel.start}
          onClose={() => setQuickPanel(null)}
          onEditDetail={handleQuickPanelEditDetail}
        />
      )}
    </div>
  )
}
