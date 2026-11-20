import type { CalendarEvent } from '@repo/ui'
import { Button, Loading, MonthCalendar, WeekCalendar } from '@repo/ui'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { useCalendarStore } from '../../../stores/calendarStore'
import LogoutButton from '../../auth/components/LogoutButton'
import EventFormModal from '../../event/components/EventFormModal'
import QuickRegistrationPanel from '../../event/components/QuickRegistrationPanel'
import ShareLinkDialog from '../../event/components/ShareLinkDialog'
import type { EventFormMode } from '../../event/types'
import TagFilterButton from '../../tag/components/TagFilterButton'
import TagManagementButton from '../../tag/components/TagManagementButton'
import { useTagsQuery } from '../../tag/hooks/useTagsQuery'
import TodayTimelinePanel from '../../timeline/components/TodayTimelinePanel'
import CalendarNavigation from '../components/CalendarNavigation'
import CalendarViewTabs from '../components/CalendarViewTabs'
import EventDetailModal from '../components/EventDetailModal'
import { CALENDAR_VIEW } from '../constants'
import { navigateDate } from '../navigateDate'
import { getViewDateRange } from '../viewDateRange'

export default function CalendarPage() {
  const { t } = useTranslation('calendar')
  const view = useCalendarStore((s) => s.view)
  const currentDate = useCalendarStore((s) => s.currentDate)
  const setView = useCalendarStore((s) => s.setView)
  const setCurrentDate = useCalendarStore((s) => s.setCurrentDate)

  const [selectedEvent, setSelectedEvent] = useState<EventResponse | null>(null)
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [formModalOpen, setFormModalOpen] = useState(false)
  const [shareModalOpen, setShareModalOpen] = useState(false)
  const [formMode, setFormMode] = useState<EventFormMode>('create')
  const [initialStart, setInitialStart] = useState<Date | null>(null)
  // Kept as page-local state so it survives month/week switches and prev/next
  // date navigation but resets whenever this page unmounts (reload, logout, or
  // leaving the route), as SPEC-008 requires.
  const [filterTagId, setFilterTagId] = useState<string | null>(null)
  // A tag deleted in the management modal or in another tab drops out of the
  // tag list, and filtering by it would keep the calendar empty. It is derived
  // rather than reset, so a stale list that predates a just-created tag does
  // not lose the user's choice.
  const { data: tags } = useTagsQuery(filterTagId !== null)
  const activeTagId =
    tags && !tags.some((tag) => tag.id === filterTagId) ? null : filterTagId
  const [quickPanel, setQuickPanel] = useState<{
    key: number
    start: Date
    anchor: { x: number; y: number }
  } | null>(null)
  // 同じ時間帯セルを連続でクリックしても必ずパネルを再マウントし、
  // 前回の登録完了状態が残らないようにするためのクリック連番
  const quickPanelClickSeqRef = useRef(0)

  const { startAt, endAt } = getViewDateRange(view, currentDate)
  const { data, isPending, isError, isPlaceholderData } = useEventsQuery(
    startAt,
    endAt,
    activeTagId,
  )

  const calendarEvents: CalendarEvent[] = (data?.events ?? []).map((e) => ({
    id: e.id,
    title: e.title,
    start: new Date(e.startAt),
    end: new Date(e.endAt),
    isSubscribed: e.isSubscribed,
    tagColors: e.tags.map((tag) => tag.color),
  }))

  function openEventDetail(event: EventResponse | null) {
    setSelectedEvent(event)
    setDetailModalOpen(true)
  }

  function handleEventClick(event: CalendarEvent) {
    // While a filter switch is loading, the events on screen belong to the
    // previous filter. Mouse and keyboard both go through this handler.
    if (isPlaceholderData) return
    openEventDetail(data?.events.find((e) => e.id === event.id) ?? null)
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

  function handleShareClick(event: EventResponse) {
    setSelectedEvent(event)
    setDetailModalOpen(false)
    setShareModalOpen(true)
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
    <div className="flex h-svh flex-col gap-4 p-4">
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
          <TagFilterButton
            selectedTagId={activeTagId}
            onChange={setFilterTagId}
          />
          <TagManagementButton />
          <CalendarViewTabs view={view} onViewChange={setView} />
          <LogoutButton />
        </div>
      </div>

      <div className="flex min-h-0 flex-1 gap-4">
        <div className="flex min-w-0 flex-1 flex-col">
          {isPending && (
            <div className="p-16">
              <Loading label={t('page.loading')} />
            </div>
          )}

          {isError && (
            <div className="flex items-center justify-center p-16 text-destructive">
              {t('page.loadError')}
            </div>
          )}

          {/* While switching the tag filter the previous filter's events stay
              visible, so they are dimmed to show they are not the result yet.
              Input is not disabled on the calendar itself because the week
              view scrolls inside it; handleEventClick ignores clicks instead. */}
          {!isPending && !isError && (
            <div
              aria-busy={isPlaceholderData}
              className={`min-h-0 flex-1 overflow-y-auto transition-opacity ${isPlaceholderData ? 'opacity-50' : ''}`}
            >
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
            </div>
          )}
        </div>

        <aside className="w-72 shrink-0 overflow-y-auto">
          <TodayTimelinePanel onEventClick={openEventDetail} />
        </aside>
      </div>

      <EventDetailModal
        open={detailModalOpen}
        event={selectedEvent}
        onClose={() => setDetailModalOpen(false)}
        onEdit={handleEditClick}
        onShare={handleShareClick}
      />

      <EventFormModal
        open={formModalOpen}
        mode={formMode}
        event={selectedEvent}
        initialStart={initialStart}
        onClose={() => setFormModalOpen(false)}
      />

      <ShareLinkDialog
        open={shareModalOpen}
        event={selectedEvent}
        onClose={() => setShareModalOpen(false)}
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
