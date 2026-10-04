import i18next from 'i18next'
import { describe, expect, it } from 'vitest'
import {
  toShareDetailError,
  getSubscribeErrorMessage,
  getUnsubscribeErrorMessage,
  formatShareDateTime,
} from './utils'

const t = i18next.getFixedT('ja', ['share', 'common'])

describe('toShareDetailError', () => {
  it('404 は notFound を返す', () => {
    expect(toShareDetailError(404, {})).toEqual({ kind: 'notFound' })
  })

  it('410 は expired を返す', () => {
    expect(toShareDetailError(410, {})).toEqual({ kind: 'expired' })
  })

  it('500 は server を返す', () => {
    expect(toShareDetailError(500, {})).toEqual({ kind: 'server' })
  })

  it('503 など 5xx は server を返す', () => {
    expect(toShareDetailError(503, {})).toEqual({ kind: 'server' })
  })

  it('その他のステータスは unknown を返す', () => {
    expect(toShareDetailError(400, {})).toEqual({ kind: 'unknown' })
  })
})

describe('getSubscribeErrorMessage', () => {
  it('FORBIDDEN コードは自分の予定は追加できません を返す', () => {
    expect(getSubscribeErrorMessage({ code: 'FORBIDDEN' }, t)).toBe(
      '自分の予定は追加できません',
    )
  })

  it('NOT_FOUND コードは共有リンクが見つかりません を返す', () => {
    expect(getSubscribeErrorMessage({ code: 'NOT_FOUND' }, t)).toBe(
      '共有リンクが見つかりません',
    )
  })

  it('GONE コードは期限切れメッセージを返す', () => {
    expect(getSubscribeErrorMessage({ code: 'GONE' }, t)).toBe(
      'この共有リンクは期限切れです',
    )
  })

  it('UNAUTHORIZED コードはログインが必要です を返す', () => {
    expect(getSubscribeErrorMessage({ code: 'UNAUTHORIZED' }, t)).toBe(
      'ログインが必要です',
    )
  })

  it('未知のコードはフォールバックメッセージを返す', () => {
    expect(getSubscribeErrorMessage({ code: 'UNKNOWN' }, t)).toBe(
      'カレンダーへの追加に失敗しました',
    )
  })

  it('code なしはフォールバックメッセージを返す', () => {
    expect(getSubscribeErrorMessage(new Error('network'), t)).toBe(
      'カレンダーへの追加に失敗しました',
    )
  })
})

describe('getUnsubscribeErrorMessage', () => {
  it('FORBIDDEN コードは他のユーザーのサブスクリプションです を返す', () => {
    expect(getUnsubscribeErrorMessage({ code: 'FORBIDDEN' }, t)).toBe(
      '他のユーザーのサブスクリプションです',
    )
  })

  it('NOT_FOUND コードは既に削除されています を返す', () => {
    expect(getUnsubscribeErrorMessage({ code: 'NOT_FOUND' }, t)).toBe(
      '既に削除されています',
    )
  })

  it('未知のコードはフォールバックメッセージを返す', () => {
    expect(getUnsubscribeErrorMessage(new Error('network'), t)).toBe(
      'カレンダーからの削除に失敗しました',
    )
  })
})

describe('formatShareDateTime', () => {
  it('ISO 文字列をローカルの年月日時分表示にフォーマットする', () => {
    // 2026-10-08T10:00:00 ローカル時刻
    const date = new Date(2026, 9, 8, 10, 0, 0)
    const result = formatShareDateTime(date.toISOString())
    expect(result).toContain('2026')
    expect(result).toContain('10')
    expect(result).toContain('8')
    expect(result).toContain('10:00')
  })
})
