import i18next from 'i18next'
import { describe, expect, it } from 'vitest'
import { getLoginErrorMessage, getSignupErrorMessage } from './utils'

const t = i18next.getFixedT('ja', 'auth')

describe('getLoginErrorMessage', () => {
  it('returns unauthorized message for UNAUTHORIZED error', () => {
    expect(getLoginErrorMessage({ code: 'UNAUTHORIZED' }, t)).toBe(
      'メールアドレスまたはパスワードが正しくありません',
    )
  })

  it('returns validation message for VALIDATION_ERROR', () => {
    expect(getLoginErrorMessage({ code: 'VALIDATION_ERROR' }, t)).toBe(
      '入力内容を確認してください',
    )
  })

  it('returns server error message for INTERNAL_ERROR', () => {
    expect(getLoginErrorMessage({ code: 'INTERNAL_ERROR' }, t)).toBe(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
  })

  it('returns fallback message for unknown error', () => {
    expect(getLoginErrorMessage(new Error('unknown'), t)).toBe(
      'ログインに失敗しました',
    )
  })
})

describe('getSignupErrorMessage', () => {
  it('returns conflict message for CONFLICT error', () => {
    expect(getSignupErrorMessage({ code: 'CONFLICT' }, t)).toBe(
      'このメールアドレスはすでに使用されています',
    )
  })

  it('returns validation message for VALIDATION_ERROR', () => {
    expect(getSignupErrorMessage({ code: 'VALIDATION_ERROR' }, t)).toBe(
      '入力内容を確認してください',
    )
  })

  it('returns server error message for INTERNAL_ERROR', () => {
    expect(getSignupErrorMessage({ code: 'INTERNAL_ERROR' }, t)).toBe(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
  })

  it('returns fallback message for unknown error', () => {
    expect(getSignupErrorMessage(new Error('unknown'), t)).toBe(
      'アカウント登録に失敗しました',
    )
  })
})
