import type { DateClickInfo } from '@fullcalendar/react'

export interface DateClickResult {
  date: Date
  anchor: { x: number; y: number }
}

// MonthCalendar/WeekCalendar共通のdateClick情報抽出処理。FullCalendarの
// DateClickInfoからdate/座標を取り出す箇所を1箇所に集約し、
// 両者の実装が分岐した際に片方の修正漏れが起きるのを防ぐ。
export function extractDateClickResult(arg: DateClickInfo): DateClickResult {
  return {
    date: arg.date,
    anchor: { x: arg.jsEvent.clientX, y: arg.jsEvent.clientY },
  }
}

export function createDateClickHandler(onDateClick: (date: Date) => void) {
  return (arg: DateClickInfo) => onDateClick(extractDateClickResult(arg).date)
}
