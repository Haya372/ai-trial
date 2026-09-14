import { AxiosError } from 'axios'
import { describe, expect, it } from 'vitest'
import { getLoginErrorMessage, getSignupErrorMessage } from './utils'

function makeAxiosError(code: string) {
  return new AxiosError(
    'Request failed',
    'ERR_BAD_REQUEST',
    undefined,
    undefined,
    {
      data: { code, message: 'error' },
      status: 400,
      statusText: '',
      headers: {},
      config: {} as never,
    },
  )
}

describe('getLoginErrorMessage', () => {
  it('returns unauthorized message for UNAUTHORIZED error', () => {
    expect(getLoginErrorMessage(makeAxiosError('UNAUTHORIZED'))).toBe(
      'メールアドレスまたはパスワードが正しくありません',
    )
  })

  it('returns validation message for VALIDATION_ERROR', () => {
    expect(getLoginErrorMessage(makeAxiosError('VALIDATION_ERROR'))).toBe(
      '入力内容を確認してください',
    )
  })

  it('returns server error message for INTERNAL_ERROR', () => {
    expect(getLoginErrorMessage(makeAxiosError('INTERNAL_ERROR'))).toBe(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
  })

  it('returns fallback message for unknown error', () => {
    expect(getLoginErrorMessage(new Error('unknown'))).toBe(
      'ログインに失敗しました',
    )
  })
})

describe('getSignupErrorMessage', () => {
  it('returns conflict message for CONFLICT error', () => {
    expect(getSignupErrorMessage(makeAxiosError('CONFLICT'))).toBe(
      'このメールアドレスはすでに使用されています',
    )
  })

  it('returns validation message for VALIDATION_ERROR', () => {
    expect(getSignupErrorMessage(makeAxiosError('VALIDATION_ERROR'))).toBe(
      '入力内容を確認してください',
    )
  })

  it('returns server error message for INTERNAL_ERROR', () => {
    expect(getSignupErrorMessage(makeAxiosError('INTERNAL_ERROR'))).toBe(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
  })

  it('returns fallback message for unknown error', () => {
    expect(getSignupErrorMessage(new Error('unknown'))).toBe(
      'アカウント登録に失敗しました',
    )
  })
})
