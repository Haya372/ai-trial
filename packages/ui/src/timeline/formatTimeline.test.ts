import { describe, expect, it } from 'vitest'
import { formatTime, formatTimeUntil } from './formatTimeline'

const labels = {
  minutesUntil: (minutes: number) => `あと${minutes}分`,
  hoursMinutesUntil: (hours: number, minutes: number) =>
    `あと${hours}時間${minutes}分`,
}

describe('formatTime', () => {
  it('ローカル時刻をゼロ埋めのHH:mmで返す', () => {
    expect(formatTime(new Date(2026, 10, 14, 9, 5))).toBe('09:05')
  })

  it('23時台も2桁で返す', () => {
    expect(formatTime(new Date(2026, 10, 14, 23, 59))).toBe('23:59')
  })
})

describe('formatTimeUntil', () => {
  it('60分未満は分だけの文言を返す', () => {
    expect(formatTimeUntil(59, labels)).toBe('あと59分')
  })

  it('60分ちょうどは時間と分の文言を返す', () => {
    expect(formatTimeUntil(60, labels)).toBe('あと1時間0分')
  })

  it('60分以上は時間と分に分けた文言を返す', () => {
    expect(formatTimeUntil(125, labels)).toBe('あと2時間5分')
  })
})
