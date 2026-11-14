import type { TodayTimelineLabels } from './types'

export function formatTime(date: Date): string {
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

export function formatTimeUntil(
  minutes: number,
  labels: Pick<TodayTimelineLabels, 'minutesUntil' | 'hoursMinutesUntil'>,
): string {
  if (minutes < 60) return labels.minutesUntil(minutes)
  return labels.hoursMinutesUntil(Math.floor(minutes / 60), minutes % 60)
}
