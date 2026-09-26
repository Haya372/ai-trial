import type { DateClickInfo } from '@fullcalendar/react'

// MonthCalendarのdateClickハンドラー。WeekCalendarはクイック登録パネルの
// 表示位置算出のためjsEventも扱う必要があり、シグネチャが異なるため個別実装している。
export function createDateClickHandler(onDateClick: (date: Date) => void) {
  return (arg: DateClickInfo) => onDateClick(arg.date)
}
