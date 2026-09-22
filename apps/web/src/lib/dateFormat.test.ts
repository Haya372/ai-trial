import { describe, expect, it } from 'vitest'
import { pad } from './dateFormat'

describe('pad', () => {
  it('1桁の数値は先頭に0を付けて2桁にする', () => {
    expect(pad(5)).toBe('05')
  })

  it('2桁の数値はそのまま返す', () => {
    expect(pad(12)).toBe('12')
  })

  it('0は"00"になる', () => {
    expect(pad(0)).toBe('00')
  })
})
