import type { DateClickInfo } from '@fullcalendar/react'

// MonthCalendar/WeekCalendar共通のdateClickハンドラー。両者の実装が
// 分岐した際に片方の修正漏れが起きるのを防ぐため共有関数として抽出している。
export function createDateClickHandler(onDateClick: (date: Date) => void) {
  return (arg: DateClickInfo) => onDateClick(arg.date)
}
