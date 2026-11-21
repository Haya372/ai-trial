import type { DatesSetInfo } from '@fullcalendar/react'

// MonthCalendar/WeekCalendar共通のdatesSet判定。FullCalendarは初回マウント時や
// gotoDate呼び出し時にもdatesSetを発火するが、そのとき表示期間はcurrentDateを含む。
// currentDateが表示期間から外れたときだけ親に通知し、表示中の日付が
// 表示期間の先頭（月初など）に書き換わるのを防ぐ。
export function shouldNotifyDateChange(
  currentDate: Date,
  info: DatesSetInfo,
): boolean {
  const { currentStart, currentEnd } = info.view
  return currentDate < currentStart || currentDate >= currentEnd
}
