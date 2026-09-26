import type { TFunction } from 'i18next'
import {
  ConflictErrorResponseCode,
  InternalErrorResponseCode,
  UnauthorizedErrorResponseCode,
  ValidationErrorResponseCode,
} from '../../api/generated'

function hasCode(value: unknown): value is { code: string } {
  return typeof value === 'object' && value !== null && 'code' in value
}

export function getLoginErrorMessage(
  error: unknown,
  t: TFunction<'auth'>,
): string {
  if (hasCode(error)) {
    switch (error.code) {
      case UnauthorizedErrorResponseCode.UNAUTHORIZED:
        return t('login.errors.unauthorized')
      case ValidationErrorResponseCode.VALIDATION_ERROR:
        return t('errors.validationError')
      case InternalErrorResponseCode.INTERNAL_ERROR:
        return t('errors.internalError')
    }
  }
  return t('login.errors.fallback')
}

export function getSignupErrorMessage(
  error: unknown,
  t: TFunction<'auth'>,
): string {
  if (hasCode(error)) {
    switch (error.code) {
      case ConflictErrorResponseCode.CONFLICT:
        return t('signup.errors.conflict')
      case ValidationErrorResponseCode.VALIDATION_ERROR:
        return t('errors.validationError')
      case InternalErrorResponseCode.INTERNAL_ERROR:
        return t('errors.internalError')
    }
  }
  return t('signup.errors.fallback')
}
