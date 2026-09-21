import { CALENDAR_VIEW, type CalendarView } from './constants'

export function navigateDate(
  view: CalendarView,
  date: Date,
  direction: 1 | -1,
): Date {
  if (view === CALENDAR_VIEW.MONTH) {
    return addMonthsClamped(date, direction)
  }
  const next = new Date(date)
  next.setDate(date.getDate() + direction * 7)
  return next
}

// Date.setMonth は日をクランプしないため、月末付近の日付から日数の少ない月へ
// 移動すると翌月にオーバーフローする（例: 1/31 + 1ヶ月 が 3/3 になる）。
// 移動先の月の最終日を上限にクランプすることでこれを防ぐ。
function addMonthsClamped(date: Date, months: number): Date {
  const day = date.getDate()
  const firstOfTargetMonth = new Date(
    date.getFullYear(),
    date.getMonth() + months,
    1,
  )
  const lastDayOfTargetMonth = new Date(
    firstOfTargetMonth.getFullYear(),
    firstOfTargetMonth.getMonth() + 1,
    0,
  ).getDate()
  firstOfTargetMonth.setDate(Math.min(day, lastDayOfTargetMonth))
  return firstOfTargetMonth
}
