import type { DatesSetInfo } from '@fullcalendar/react'
import { describe, expect, it } from 'vitest'
import { shouldNotifyDateChange } from './shouldNotifyDateChange'

function datesSetInfo(currentStart: Date, currentEnd: Date): DatesSetInfo {
  return { view: { currentStart, currentEnd } } as unknown as DatesSetInfo
}

// 2030年6月の月ビュー（currentEndは排他的な翌月1日）
const juneView = datesSetInfo(new Date(2030, 5, 1), new Date(2030, 6, 1))

describe('shouldNotifyDateChange', () => {
  it('表示期間の途中の日付を表示中なら通知しない（初回マウント時に今日が月初へ書き換わらない）', () => {
    expect(shouldNotifyDateChange(new Date(2030, 5, 12), juneView)).toBe(false)
  })

  it('表示期間の先頭の日付を表示中なら通知しない', () => {
    expect(shouldNotifyDateChange(new Date(2030, 5, 1), juneView)).toBe(false)
  })

  it('表示期間より前の日付を表示中なら通知する', () => {
    expect(shouldNotifyDateChange(new Date(2030, 4, 31), juneView)).toBe(true)
  })

  it('表示期間の終端（排他的）の日付を表示中なら通知する', () => {
    expect(shouldNotifyDateChange(new Date(2030, 6, 1), juneView)).toBe(true)
  })
})
