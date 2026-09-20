import {
  ConflictErrorResponseCode,
  InternalErrorResponseCode,
  UnauthorizedErrorResponseCode,
  ValidationErrorResponseCode,
} from '../../api/generated'

function hasCode(value: unknown): value is { code: string } {
  return typeof value === 'object' && value !== null && 'code' in value
}

export function getLoginErrorMessage(error: unknown): string {
  if (hasCode(error)) {
    switch (error.code) {
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
  if (hasCode(error)) {
    switch (error.code) {
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
