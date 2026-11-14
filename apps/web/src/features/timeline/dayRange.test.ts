import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'
import { getDayRange } from './dayRange'

describe('getDayRange', () => {
  it('指定日時のローカル0時〜翌日ローカル0時を返す', () => {
    const { startAt, endAt } = getDayRange(new Date(2026, 10, 14, 15, 30))
    expect(startAt).toEqual(new Date(2026, 10, 14))
    expect(endAt).toEqual(new Date(2026, 10, 15))
  })

  it('月末は翌月1日を endAt にする', () => {
    const { endAt } = getDayRange(new Date(2026, 10, 30, 23, 59))
    expect(endAt).toEqual(new Date(2026, 11, 1))
  })

  it('年末は翌年1月1日を endAt にする', () => {
    const { endAt } = getDayRange(new Date(2026, 11, 31, 12, 0))
    expect(endAt).toEqual(new Date(2027, 0, 1))
  })

  it('0時ちょうどはその日の範囲を返す', () => {
    const { startAt, endAt } = getDayRange(new Date(2026, 10, 14))
    expect(startAt).toEqual(new Date(2026, 10, 14))
    expect(endAt).toEqual(new Date(2026, 10, 15))
  })
})

describe('getDayRange（夏時間のあるタイムゾーン）', () => {
  beforeAll(() => {
    vi.stubEnv('TZ', 'America/New_York')
  })
  afterAll(() => {
    vi.unstubAllEnvs()
  })

  it('夏時間の開始日（23時間の日）でも endAt を翌日0時にする', () => {
    const { startAt, endAt } = getDayRange(new Date(2026, 2, 8, 12, 0))
    expect(endAt).toEqual(new Date(2026, 2, 9))
    expect(endAt.getTime() - startAt.getTime()).toBe(23 * 60 * 60 * 1000)
  })
})
