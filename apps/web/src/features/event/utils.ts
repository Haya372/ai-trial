import type { EventResponse } from '../../api/generated'
import { pad } from '../../lib/dateFormat'
import type { EventFormMode, EventFormValues } from './types'

export function toDateTimeLocalValue(iso: string): string {
  const date = new Date(iso)
  const y = date.getFullYear()
  const mo = pad(date.getMonth() + 1)
  const d = pad(date.getDate())
  const h = pad(date.getHours())
  const mi = pad(date.getMinutes())
  return `${y}-${mo}-${d}T${h}:${mi}`
}

export function toIsoString(value: string): string {
  return new Date(value).toISOString()
}

function defaultFormValues(): EventFormValues {
  const now = new Date()
  const oneHourLater = new Date(now.getTime() + 60 * 60 * 1000)
  return {
    title: '',
    startAt: toDateTimeLocalValue(now.toISOString()),
    endAt: toDateTimeLocalValue(oneHourLater.toISOString()),
    description: '',
    location: '',
    url: '',
  }
}

export function toFormValues(event: EventResponse | null): EventFormValues {
  if (!event) return defaultFormValues()
  return {
    title: event.title,
    startAt: toDateTimeLocalValue(event.startAt),
    endAt: toDateTimeLocalValue(event.endAt),
    description: event.description ?? '',
    location: event.location ?? '',
    url: event.url ?? '',
  }
}

function hasCode(value: unknown): value is { code: string } {
  return typeof value === 'object' && value !== null && 'code' in value
}

export function getEventErrorMessage(
  error: unknown,
  mode: EventFormMode | 'delete',
): string {
  if (hasCode(error)) {
    switch (error.code) {
      case 'VALIDATION_ERROR':
        return '入力内容を確認してください'
      case 'NOT_FOUND':
        return '予定が見つかりませんでした（削除された可能性があります）'
      case 'UNAUTHORIZED':
        return 'ログインが必要です'
      case 'INTERNAL_ERROR':
        return 'サーバーエラーが発生しました。しばらく経ってから再試行してください'
    }
  }
  if (mode === 'create') return '予定の登録に失敗しました'
  if (mode === 'edit') return '予定の更新に失敗しました'
  return '予定の削除に失敗しました'
}
