import type { TFunction } from 'i18next'
import {
  ConflictErrorResponseCode,
  InternalErrorResponseCode,
  UnauthorizedErrorResponseCode,
  ValidationErrorResponseCode,
} from '../../api/generated'
import { mapErrorToMessage } from '../../lib/errorMessage'

type AuthTFunction = TFunction<['auth', 'common']>

export function getLoginErrorMessage(error: unknown, t: AuthTFunction): string {
  return mapErrorToMessage(
    error,
    t,
    {
      [UnauthorizedErrorResponseCode.UNAUTHORIZED]: 'login.errors.unauthorized',
      [ValidationErrorResponseCode.VALIDATION_ERROR]:
        'common:errors.validationError',
      [InternalErrorResponseCode.INTERNAL_ERROR]: 'common:errors.internalError',
    },
    'login.errors.fallback',
  )
}

export function getSignupErrorMessage(
  error: unknown,
  t: AuthTFunction,
): string {
  return mapErrorToMessage(
    error,
    t,
    {
      [ConflictErrorResponseCode.CONFLICT]: 'signup.errors.conflict',
      [ValidationErrorResponseCode.VALIDATION_ERROR]:
        'common:errors.validationError',
      [InternalErrorResponseCode.INTERNAL_ERROR]: 'common:errors.internalError',
    },
    'signup.errors.fallback',
  )
}
