import type { ParseKeys } from 'i18next'
import { mapErrorToMessage } from '../../lib/errorMessage'
import { pad } from '../../lib/dateFormat'

export type ShareDetailError =
  | { kind: 'notFound' }
  | { kind: 'expired' }
  | { kind: 'server' }
  | { kind: 'unknown' }

export function toShareDetailError(
  status: number,
  _data: unknown,
): ShareDetailError {
  if (status === 404) return { kind: 'notFound' }
  if (status === 410) return { kind: 'expired' }
  if (status >= 500) return { kind: 'server' }
  return { kind: 'unknown' }
}

// TFunction のブランド型を避けるため、mapErrorToMessage が実際に必要とする callable 型を使う
type ShareTFunction = (key: ParseKeys<['share', 'common']>) => string

const subscribeCodeToKey: Record<string, ParseKeys<['share', 'common']>> = {
  FORBIDDEN: 'share:errors.subscribeOwnEvent',
  NOT_FOUND: 'share:errors.notFound',
  GONE: 'share:errors.expired',
  UNAUTHORIZED: 'share:errors.unauthorized',
  INTERNAL_ERROR: 'common:errors.internalError',
}

const unsubscribeCodeToKey: Record<string, ParseKeys<['share', 'common']>> = {
  FORBIDDEN: 'share:errors.unsubscribeNotOwner',
  NOT_FOUND: 'share:errors.subscriptionGone',
  UNAUTHORIZED: 'share:errors.unauthorized',
  INTERNAL_ERROR: 'common:errors.internalError',
}

export function getSubscribeErrorMessage(
  error: unknown,
  t: ShareTFunction,
): string {
  return mapErrorToMessage(
    error,
    t,
    subscribeCodeToKey,
    'share:errors.subscribeFallback',
  )
}

export function getUnsubscribeErrorMessage(
  error: unknown,
  t: ShareTFunction,
): string {
  return mapErrorToMessage(
    error,
    t,
    unsubscribeCodeToKey,
    'share:errors.unsubscribeFallback',
  )
}

export function formatShareDateTime(isoString: string): string {
  const d = new Date(isoString)
  const year = d.getFullYear()
  const month = d.getMonth() + 1
  const day = d.getDate()
  const time = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  return `${year}年${pad(month)}月${pad(day)}日 ${time}`
}
