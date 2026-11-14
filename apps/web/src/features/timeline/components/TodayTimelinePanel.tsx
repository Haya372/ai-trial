import type { CalendarEvent, TodayTimelineLabels } from '@repo/ui'
import { Loading, TodayTimeline } from '@repo/ui'
import { useTranslation } from 'react-i18next'
import type { EventResponse } from '../../../api/generated'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { getDayRange } from '../dayRange'
import { useNow } from '../hooks/useNow'

interface TodayTimelinePanelProps {
  onEventClick: (event: EventResponse) => void
}

export default function TodayTimelinePanel({
  onEventClick,
}: TodayTimelinePanelProps) {
  const { t } = useTranslation('calendar')
  const now = useNow()
  // now の日付が変わるとクエリキーが変わり、新しい日の予定を取得する。
  // タグの絞り込みはパネルに適用しない（SPEC-007）
  const { startAt, endAt } = getDayRange(now)
  const { data, isPending, isError } = useEventsQuery(startAt, endAt)

  const labels: TodayTimelineLabels = {
    title: t('timeline.title'),
    noEventsToday: t('timeline.noEventsToday'),
    noMoreEventsToday: t('timeline.noMoreEventsToday'),
    fromPreviousDay: t('timeline.fromPreviousDay'),
    untilNextDay: t('timeline.untilNextDay'),
    minutesUntil: (minutes) => t('timeline.minutesUntil', { minutes }),
    hoursMinutesUntil: (hours, minutes) =>
      t('timeline.hoursMinutesUntil', { hours, minutes }),
    freeTime: (minutes) => t('timeline.freeTime', { minutes }),
  }

  if (isPending) {
    return <Loading label={t('timeline.loading')} />
  }

  if (isError) {
    return <p className="text-sm text-destructive">{t('timeline.loadError')}</p>
  }

  const events: CalendarEvent[] = data.events.map((e) => ({
    id: e.id,
    title: e.title,
    start: new Date(e.startAt),
    end: new Date(e.endAt),
    isSubscribed: e.isSubscribed,
  }))

  function handleEventClick(event: CalendarEvent) {
    const fullEvent = data?.events.find((e) => e.id === event.id)
    if (fullEvent) onEventClick(fullEvent)
  }

  return (
    <TodayTimeline
      events={events}
      now={now}
      labels={labels}
      onEventClick={handleEventClick}
    />
  )
}
