import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'
import { CALENDAR_VIEW } from './constants'
import { getViewDateRange } from './viewDateRange'

describe('getViewDateRange', () => {
  describe('月ビュー', () => {
    it('当月1日のローカル0時から翌月1日のローカル0時までを返す', () => {
      const result = getViewDateRange(
        CALENDAR_VIEW.MONTH,
        new Date(2026, 10, 15, 12, 34),
      )
      expect(result).toEqual({
        startAt: new Date(2026, 10, 1),
        endAt: new Date(2026, 11, 1),
      })
    })

    it('12月は翌年1月1日を終了にする', () => {
      const result = getViewDateRange(
        CALENDAR_VIEW.MONTH,
        new Date(2026, 11, 31),
      )
      expect(result).toEqual({
        startAt: new Date(2026, 11, 1),
        endAt: new Date(2027, 0, 1),
      })
    })
  })

  describe('週ビュー', () => {
    it('日曜のローカル0時から翌週日曜のローカル0時までを返し、時刻は切り捨てる', () => {
      // 2026-11-11 は水曜
      const result = getViewDateRange(
        CALENDAR_VIEW.WEEK,
        new Date(2026, 10, 11, 15, 30),
      )
      expect(result).toEqual({
        startAt: new Date(2026, 10, 8),
        endAt: new Date(2026, 10, 15),
      })
    })

    it('日曜を渡すとその日から始まる', () => {
      const result = getViewDateRange(
        CALENDAR_VIEW.WEEK,
        new Date(2026, 10, 8, 9),
      )
      expect(result).toEqual({
        startAt: new Date(2026, 10, 8),
        endAt: new Date(2026, 10, 15),
      })
    })

    it('月をまたぐ週を返す', () => {
      // 2026-12-31 は木曜
      const result = getViewDateRange(
        CALENDAR_VIEW.WEEK,
        new Date(2026, 11, 31),
      )
      expect(result).toEqual({
        startAt: new Date(2026, 11, 27),
        endAt: new Date(2027, 0, 3),
      })
    })
  })

  describe('UTC以外のタイムゾーン（JST）', () => {
    beforeAll(() => {
      vi.stubEnv('TZ', 'Asia/Tokyo')
    })
    afterAll(() => {
      vi.unstubAllEnvs()
    })

    it('11月の月ビューの範囲をUTCにするとJSTの時差の分だけ前にずれる', () => {
      const { startAt, endAt } = getViewDateRange(
        CALENDAR_VIEW.MONTH,
        new Date(2026, 10, 1, 5),
      )
      expect(startAt.toISOString()).toBe('2026-10-31T15:00:00.000Z')
      expect(endAt.toISOString()).toBe('2026-11-30T15:00:00.000Z')
    })
  })
})
