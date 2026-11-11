import { CALENDAR_VIEW, type CalendarView } from './constants'

// 表示期間を、ローカル0時を区切りとする [startAt, endAt) で返す。
// 1日の区切りはユーザーのタイムゾーンに依存するため、クライアントで計算する（ADR-036）
export function getViewDateRange(
  view: CalendarView,
  date: Date,
): { startAt: Date; endAt: Date } {
  if (view === CALENDAR_VIEW.MONTH) {
    return {
      startAt: new Date(date.getFullYear(), date.getMonth(), 1),
      endAt: new Date(date.getFullYear(), date.getMonth() + 1, 1),
    }
  }
  // 週ビュー: 日曜始まり
  const startAt = new Date(
    date.getFullYear(),
    date.getMonth(),
    date.getDate() - date.getDay(),
  )
  const endAt = new Date(
    startAt.getFullYear(),
    startAt.getMonth(),
    startAt.getDate() + 7,
  )
  return { startAt, endAt }
}
