import { describe, expect, it } from 'vitest'
import { CALENDAR_VIEW } from './constants'
import { navigateDate } from './navigateDate'

describe('navigateDate', () => {
  describe('月ビュー', () => {
    it('通常の日付では単純に月を+1する', () => {
      const result = navigateDate(CALENDAR_VIEW.MONTH, new Date(2026, 8, 15), 1)
      expect(result).toEqual(new Date(2026, 9, 15))
    })

    it('通常の日付では単純に月を-1する', () => {
      const result = navigateDate(
        CALENDAR_VIEW.MONTH,
        new Date(2026, 8, 15),
        -1,
      )
      expect(result).toEqual(new Date(2026, 7, 15))
    })

    it('1月31日から次へ進むと、2月は28日までしかないため2月28日にクランプされる', () => {
      const result = navigateDate(CALENDAR_VIEW.MONTH, new Date(2026, 0, 31), 1)
      expect(result).toEqual(new Date(2026, 1, 28))
    })

    it('うるう年の1月31日から次へ進むと2月29日にクランプされる', () => {
      const result = navigateDate(CALENDAR_VIEW.MONTH, new Date(2028, 0, 31), 1)
      expect(result).toEqual(new Date(2028, 1, 29))
    })

    it('3月31日から次へ進むと、4月は30日までしかないため4月30日にクランプされる', () => {
      const result = navigateDate(CALENDAR_VIEW.MONTH, new Date(2026, 2, 31), 1)
      expect(result).toEqual(new Date(2026, 3, 30))
    })

    it('3月31日から前へ戻ると、2月は28日までしかないため2月28日にクランプされる', () => {
      const result = navigateDate(
        CALENDAR_VIEW.MONTH,
        new Date(2026, 2, 31),
        -1,
      )
      expect(result).toEqual(new Date(2026, 1, 28))
    })
  })

  describe('週ビュー', () => {
    it('次へ進むと7日進む', () => {
      const result = navigateDate(CALENDAR_VIEW.WEEK, new Date(2026, 8, 13), 1)
      expect(result).toEqual(new Date(2026, 8, 20))
    })

    it('前へ戻ると7日戻る', () => {
      const result = navigateDate(CALENDAR_VIEW.WEEK, new Date(2026, 8, 13), -1)
      expect(result).toEqual(new Date(2026, 8, 6))
    })
  })
})
