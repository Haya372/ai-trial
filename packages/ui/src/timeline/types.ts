// Labels are injected so the host app can localize them (apps/web uses i18next).
export interface TodayTimelineLabels {
  title: string
  noEventsToday: string
  noMoreEventsToday: string
  fromPreviousDay: string
  untilNextDay: string
  minutesUntil: (minutes: number) => string
  hoursMinutesUntil: (hours: number, minutes: number) => string
  freeTime: (minutes: number) => string
}
