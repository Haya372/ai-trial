import { AxiosError } from 'axios'
import {
  ConflictErrorResponseCode,
  InternalErrorResponseCode,
  UnauthorizedErrorResponseCode,
  ValidationErrorResponseCode,
} from '../../api/generated'

type ErrorWithCode = { code: string }

export function getLoginErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ErrorWithCode | undefined
    switch (data?.code) {
      case UnauthorizedErrorResponseCode.UNAUTHORIZED:
        return 'メールアドレスまたはパスワードが正しくありません'
      case ValidationErrorResponseCode.VALIDATION_ERROR:
        return '入力内容を確認してください'
      case InternalErrorResponseCode.INTERNAL_ERROR:
        return 'サーバーエラーが発生しました。しばらく経ってから再試行してください'
    }
  }
  return 'ログインに失敗しました'
}

export function getSignupErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ErrorWithCode | undefined
    switch (data?.code) {
      case ConflictErrorResponseCode.CONFLICT:
        return 'このメールアドレスはすでに使用されています'
      case ValidationErrorResponseCode.VALIDATION_ERROR:
        return '入力内容を確認してください'
      case InternalErrorResponseCode.INTERNAL_ERROR:
        return 'サーバーエラーが発生しました。しばらく経ってから再試行してください'
    }
  }
  return 'アカウント登録に失敗しました'
}
