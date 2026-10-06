import type { ParseKeys, TFunction } from 'i18next'
import type { CreateEventRequest, EventResponse } from '../../api/generated'
import { pad } from '../../lib/dateFormat'
import { mapErrorToMessage } from '../../lib/errorMessage'
import type { EventFormMode, EventFormValues } from './types'

export function defaultExpiresAtValue(event: EventResponse): string {
  // SPEC-004 では「予定終了日時」をデフォルトとする。
  // endAt が過去の場合、クライアント側 zod の「未来チェック」に引っかかるため
  // ユーザーが明示的に変更して送信することで吸収する設計（design doc の判断を踏襲）。
  return toDateTimeLocalValue(event.endAt)
}

export function toFullShareUrl(path: string): string {
  // バックエンドは "/share/{token}" のパスを返す。コピー用途では完全 URL が必要なため
  // SSR 想定外の本アプリでは window.location.origin を直参照する（KISS）。
  return `${window.location.origin}${path}`
}

type ShareTFunction = TFunction<['eventshare', 'common', 'event']>

const shareCodeToKey: Record<
  string,
  ParseKeys<['eventshare', 'common', 'event']>
> = {
  FORBIDDEN: 'eventshare:errors.forbidden',
  // 予定が見つからない・未ログインの文言は event 名前空間の既存キーを再利用する
  // （eventshare 用に同じ文言を複製しない）
  NOT_FOUND: 'event:errors.notFound',
  UNAUTHORIZED: 'event:errors.unauthorized',
  INTERNAL_ERROR: 'common:errors.internalError',
}

export function getShareErrorMessage(
  error: unknown,
  t: ShareTFunction,
): string {
  return mapErrorToMessage(
    error,
    t,
    shareCodeToKey,
    'eventshare:errors.createFallback',
  )
}

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

export function toCreateEventRequest(params: {
  title: string
  startAt: string
  endAt: string
  description?: string | null
  location?: string | null
  url?: string | null
}): CreateEventRequest {
  return {
    title: params.title,
    startAt: params.startAt,
    endAt: params.endAt,
    description: params.description || null,
    location: params.location || null,
    url: params.url || null,
  }
}

export const DEFAULT_EVENT_DURATION_MS = 60 * 60 * 1000

function defaultFormValues(initialStart?: Date | null): EventFormValues {
  const start = initialStart ?? new Date()
  const end = new Date(start.getTime() + DEFAULT_EVENT_DURATION_MS)
  return {
    title: '',
    startAt: toDateTimeLocalValue(start.toISOString()),
    endAt: toDateTimeLocalValue(end.toISOString()),
    description: '',
    location: '',
    url: '',
  }
}

export function toFormValues(
  event: EventResponse | null,
  initialStart?: Date | null,
): EventFormValues {
  if (!event) return defaultFormValues(initialStart)
  return {
    title: event.title,
    startAt: toDateTimeLocalValue(event.startAt),
    endAt: toDateTimeLocalValue(event.endAt),
    description: event.description ?? '',
    location: event.location ?? '',
    url: event.url ?? '',
  }
}

type EventTFunction = TFunction<['event', 'common']>

const fallbackKeyByMode: Record<
  EventFormMode | 'delete' | 'unsubscribe',
  ParseKeys<['event', 'common']>
> = {
  create: 'errors.createFallback',
  edit: 'errors.editFallback',
  delete: 'errors.deleteFallback',
  unsubscribe: 'errors.unsubscribeFallback',
}

const codeToKey: Record<string, ParseKeys<['event', 'common']>> = {
  VALIDATION_ERROR: 'common:errors.validationError',
  NOT_FOUND: 'errors.notFound',
  UNAUTHORIZED: 'errors.unauthorized',
  INTERNAL_ERROR: 'common:errors.internalError',
}

// unsubscribe only removes the caller's EventSubscription, not the event
// itself, so a 404 here means "already removed from your calendar", not
// "event not found" — distinct enough from codeToKey's NOT_FOUND to warrant
// its own map rather than a mode-keyed exception inside one shared map.
const unsubscribeCodeToKey: Record<string, ParseKeys<['event', 'common']>> = {
  ...codeToKey,
  NOT_FOUND: 'errors.subscriptionNotFound',
}

export function getEventErrorMessage(
  error: unknown,
  mode: EventFormMode | 'delete' | 'unsubscribe',
  t: EventTFunction,
): string {
  const map = mode === 'unsubscribe' ? unsubscribeCodeToKey : codeToKey
  return mapErrorToMessage(error, t, map, fallbackKeyByMode[mode])
}
