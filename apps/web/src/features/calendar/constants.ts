export const CALENDAR_VIEW = {
  MONTH: 'month',
  WEEK: 'week',
} as const

export type CalendarView = (typeof CALENDAR_VIEW)[keyof typeof CALENDAR_VIEW]
