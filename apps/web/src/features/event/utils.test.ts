import { describe, expect, it } from 'vitest'
import type { EventResponse } from '../../api/generated'
import {
  getEventErrorMessage,
  toDateTimeLocalValue,
  toFormValues,
  toIsoString,
} from './utils'

describe('toDateTimeLocalValue', () => {
  it('ISO文字列をdatetime-local形式（YYYY-MM-DDTHH:mm）に変換する', () => {
    const date = new Date(2026, 8, 22, 9, 5) // 2026-09-22 09:05 ローカル時刻
    expect(toDateTimeLocalValue(date.toISOString())).toBe('2026-09-22T09:05')
  })
})

describe('toIsoString', () => {
  it('datetime-local形式の文字列をISO文字列に変換する', () => {
    const result = toIsoString('2026-09-22T09:05')
    expect(new Date(result).getFullYear()).toBe(2026)
    expect(new Date(result).getMonth()).toBe(8)
    expect(new Date(result).getDate()).toBe(22)
    expect(new Date(result).getHours()).toBe(9)
    expect(new Date(result).getMinutes()).toBe(5)
  })
})

describe('toFormValues', () => {
  it('eventがnullの場合、タイトル等は空文字のデフォルト値になる', () => {
    const values = toFormValues(null)
    expect(values.title).toBe('')
    expect(values.description).toBe('')
    expect(values.location).toBe('')
    expect(values.url).toBe('')
  })

  it('eventが渡された場合、既存の値をフォーム値にマッピングする', () => {
    const event: EventResponse = {
      id: 'event-1',
      title: 'デザインレビュー',
      description: 'メモ',
      startAt: '2026-09-22T01:00:00.000Z',
      endAt: '2026-09-22T02:00:00.000Z',
      location: '会議室A',
      url: 'https://example.com',
    }
    const values = toFormValues(event)
    expect(values.title).toBe('デザインレビュー')
    expect(values.description).toBe('メモ')
    expect(values.location).toBe('会議室A')
    expect(values.url).toBe('https://example.com')
    expect(values.startAt).toBe(toDateTimeLocalValue(event.startAt))
    expect(values.endAt).toBe(toDateTimeLocalValue(event.endAt))
  })

  it('eventがnullでinitialStartが渡された場合、開始日時にinitialStartを設定する', () => {
    const initialStart = new Date(2026, 8, 22, 14, 30)
    const values = toFormValues(null, initialStart)
    expect(values.startAt).toBe(
      toDateTimeLocalValue(initialStart.toISOString()),
    )
  })

  it('eventがnullでinitialStartが渡された場合、終了日時は開始日時の1時間後にする', () => {
    const initialStart = new Date(2026, 8, 22, 14, 30)
    const values = toFormValues(null, initialStart)
    const expectedEnd = new Date(initialStart.getTime() + 60 * 60 * 1000)
    expect(values.endAt).toBe(toDateTimeLocalValue(expectedEnd.toISOString()))
  })

  it('eventが渡された場合、initialStartが渡されても既存の開始日時を使う', () => {
    const initialStart = new Date(2026, 8, 22, 14, 30)
    const event: EventResponse = {
      id: 'event-1',
      title: 'デザインレビュー',
      description: null,
      startAt: '2026-09-22T01:00:00.000Z',
      endAt: '2026-09-22T02:00:00.000Z',
      location: null,
      url: null,
    }
    const values = toFormValues(event, initialStart)
    expect(values.startAt).toBe(toDateTimeLocalValue(event.startAt))
  })

  it('eventのdescription/location/urlがnullの場合は空文字にする', () => {
    const event: EventResponse = {
      id: 'event-1',
      title: 'デザインレビュー',
      description: null,
      startAt: '2026-09-22T01:00:00.000Z',
      endAt: '2026-09-22T02:00:00.000Z',
      location: null,
      url: null,
    }
    const values = toFormValues(event)
    expect(values.description).toBe('')
    expect(values.location).toBe('')
    expect(values.url).toBe('')
  })
})

describe('getEventErrorMessage', () => {
  it('VALIDATION_ERRORの場合は入力確認メッセージを返す', () => {
    expect(getEventErrorMessage({ code: 'VALIDATION_ERROR' }, 'create')).toBe(
      '入力内容を確認してください',
    )
  })

  it('NOT_FOUNDの場合は予定が見つからない旨のメッセージを返す', () => {
    expect(getEventErrorMessage({ code: 'NOT_FOUND' }, 'edit')).toBe(
      '予定が見つかりませんでした（削除された可能性があります）',
    )
  })

  it('UNAUTHORIZEDの場合はログインが必要な旨のメッセージを返す', () => {
    expect(getEventErrorMessage({ code: 'UNAUTHORIZED' }, 'create')).toBe(
      'ログインが必要です',
    )
  })

  it('INTERNAL_ERRORの場合はサーバーエラーメッセージを返す', () => {
    expect(getEventErrorMessage({ code: 'INTERNAL_ERROR' }, 'create')).toBe(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
  })

  it('未知のエラーの場合、createモードでは登録失敗メッセージを返す', () => {
    expect(getEventErrorMessage(new Error('unknown'), 'create')).toBe(
      '予定の登録に失敗しました',
    )
  })

  it('未知のエラーの場合、editモードでは更新失敗メッセージを返す', () => {
    expect(getEventErrorMessage(new Error('unknown'), 'edit')).toBe(
      '予定の更新に失敗しました',
    )
  })

  it('未知のエラーの場合、deleteモードでは削除失敗メッセージを返す', () => {
    expect(getEventErrorMessage(new Error('unknown'), 'delete')).toBe(
      '予定の削除に失敗しました',
    )
  })

  it('deleteモードでもNOT_FOUND等のコード別メッセージを返す', () => {
    expect(getEventErrorMessage({ code: 'NOT_FOUND' }, 'delete')).toBe(
      '予定が見つかりませんでした（削除された可能性があります）',
    )
  })
})
