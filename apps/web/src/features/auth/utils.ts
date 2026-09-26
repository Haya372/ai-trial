import type { ParseKeys, TFunction } from 'i18next'
import {
  ConflictErrorResponseCode,
  InternalErrorResponseCode,
  UnauthorizedErrorResponseCode,
  ValidationErrorResponseCode,
} from '../../api/generated'

function hasCode(value: unknown): value is { code: string } {
  return typeof value === 'object' && value !== null && 'code' in value
}

function mapErrorToMessage<K extends ParseKeys<'auth'>>(
  error: unknown,
  t: TFunction<'auth'>,
  codeToKey: Record<string, K>,
  fallbackKey: K,
): string {
  if (hasCode(error) && error.code in codeToKey) {
    return t(codeToKey[error.code])
  }
  return t(fallbackKey)
}

export function getLoginErrorMessage(
  error: unknown,
  t: TFunction<'auth'>,
): string {
  return mapErrorToMessage(
    error,
    t,
    {
      [UnauthorizedErrorResponseCode.UNAUTHORIZED]: 'login.errors.unauthorized',
      [ValidationErrorResponseCode.VALIDATION_ERROR]: 'errors.validationError',
      [InternalErrorResponseCode.INTERNAL_ERROR]: 'errors.internalError',
    },
    'login.errors.fallback',
  )
}

export function getSignupErrorMessage(
  error: unknown,
  t: TFunction<'auth'>,
): string {
  return mapErrorToMessage(
    error,
    t,
    {
      [ConflictErrorResponseCode.CONFLICT]: 'signup.errors.conflict',
      [ValidationErrorResponseCode.VALIDATION_ERROR]: 'errors.validationError',
      [InternalErrorResponseCode.INTERNAL_ERROR]: 'errors.internalError',
    },
    'signup.errors.fallback',
  )
}
