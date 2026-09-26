import { Button } from '@repo/ui'
import type { TFunction } from 'i18next'
import { useTranslation } from 'react-i18next'
import { CALENDAR_VIEW, type CalendarView } from '../constants'

interface CalendarNavigationProps {
  view: CalendarView
  currentDate: Date
  onPrev: () => void
  onNext: () => void
  onToday: () => void
}

function formatPeriodLabel(
  t: TFunction<'calendar'>,
  view: CalendarView,
  date: Date,
): string {
  const year = date.getFullYear()
  const month = date.getMonth() + 1

  if (view === CALENDAR_VIEW.MONTH) {
    return t('navigation.monthLabel', { year, month })
  }

  // 週ビュー: 週の開始日(日曜)と終了日(土曜)を表示
  const dayOfWeek = date.getDay()
  const weekStart = new Date(date)
  weekStart.setDate(date.getDate() - dayOfWeek)
  const weekEnd = new Date(weekStart)
  weekEnd.setDate(weekStart.getDate() + 6)

  const startYear = weekStart.getFullYear()
  const startMonth = weekStart.getMonth() + 1
  const startDay = weekStart.getDate()
  const endMonth = weekEnd.getMonth() + 1
  const endDay = weekEnd.getDate()

  if (startMonth === endMonth) {
    return t('navigation.weekLabelSameMonth', {
      year: startYear,
      month: startMonth,
      startDay,
      endDay,
    })
  }
  return t('navigation.weekLabelDiffMonth', {
    year: startYear,
    month: startMonth,
    startDay,
    endMonth,
    endDay,
  })
}

export default function CalendarNavigation({
  view,
  currentDate,
  onPrev,
  onNext,
  onToday,
}: CalendarNavigationProps) {
  const { t } = useTranslation('calendar')
  const label = formatPeriodLabel(t, view, currentDate)

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={onPrev}
        aria-label={t('navigation.prev')}
      >
        {t('navigation.prev')}
      </Button>
      <Button
        variant="outline"
        size="sm"
        onClick={onNext}
        aria-label={t('navigation.next')}
      >
        {t('navigation.next')}
      </Button>
      <Button
        variant="outline"
        size="sm"
        onClick={onToday}
        aria-label={t('navigation.today')}
      >
        {t('navigation.today')}
      </Button>
      <span className="text-lg font-semibold">{label}</span>
    </div>
  )
}
