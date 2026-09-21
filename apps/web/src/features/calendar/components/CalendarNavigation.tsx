import { Button } from '@repo/ui'
import { CALENDAR_VIEW, type CalendarView } from '../constants'

interface CalendarNavigationProps {
  view: CalendarView
  currentDate: Date
  onPrev: () => void
  onNext: () => void
  onToday: () => void
}

function formatPeriodLabel(view: CalendarView, date: Date): string {
  const year = date.getFullYear()
  const month = date.getMonth() + 1

  if (view === CALENDAR_VIEW.MONTH) {
    return `${year}年${month}月`
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
    return `${startYear}年${startMonth}月${startDay}日〜${endDay}日`
  }
  return `${startYear}年${startMonth}月${startDay}日〜${endMonth}月${endDay}日`
}

export default function CalendarNavigation({
  view,
  currentDate,
  onPrev,
  onNext,
  onToday,
}: CalendarNavigationProps) {
  const label = formatPeriodLabel(view, currentDate)

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={onPrev} aria-label="前へ">
        前へ
      </Button>
      <Button variant="outline" size="sm" onClick={onNext} aria-label="次へ">
        次へ
      </Button>
      <Button variant="outline" size="sm" onClick={onToday} aria-label="今日">
        今日
      </Button>
      <span className="text-lg font-semibold">{label}</span>
    </div>
  )
}
