import type { ParseKeys, TFunction } from 'i18next'
import type { EventResponse } from '../../api/generated'
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

type ShareTFunction = TFunction<['eventshare', 'common']>

const shareCodeToKey: Record<string, ParseKeys<['eventshare', 'common']>> = {
  FORBIDDEN: 'eventshare:errors.forbidden',
  NOT_FOUND: 'eventshare:errors.eventNotFound',
  UNAUTHORIZED: 'common:errors.unauthorized',
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
  EventFormMode | 'delete',
  ParseKeys<['event', 'common']>
> = {
  create: 'errors.createFallback',
  edit: 'errors.editFallback',
  delete: 'errors.deleteFallback',
}

const codeToKey: Record<string, ParseKeys<['event', 'common']>> = {
  VALIDATION_ERROR: 'common:errors.validationError',
  NOT_FOUND: 'errors.notFound',
  UNAUTHORIZED: 'errors.unauthorized',
  INTERNAL_ERROR: 'common:errors.internalError',
}

export function getEventErrorMessage(
  error: unknown,
  mode: EventFormMode | 'delete',
  t: EventTFunction,
): string {
  return mapErrorToMessage(error, t, codeToKey, fallbackKeyByMode[mode])
}
