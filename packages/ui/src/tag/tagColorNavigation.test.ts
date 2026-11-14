import { describe, expect, it } from 'vitest'
import { tagColorForKey } from './tagColorNavigation'

describe('tagColorForKey', () => {
  it.each(['ArrowRight', 'ArrowDown'])('%s で次の色を返す', (key) => {
    expect(tagColorForKey('red', key)).toBe('orange')
  })

  it.each(['ArrowLeft', 'ArrowUp'])('%s で前の色を返す', (key) => {
    expect(tagColorForKey('orange', key)).toBe('red')
  })

  it('最後の色の次は最初の色に戻る', () => {
    expect(tagColorForKey('gray', 'ArrowRight')).toBe('red')
  })

  it('最初の色の前は最後の色に戻る', () => {
    expect(tagColorForKey('red', 'ArrowLeft')).toBe('gray')
  })

  it('Home で最初の色、End で最後の色を返す', () => {
    expect(tagColorForKey('blue', 'Home')).toBe('red')
    expect(tagColorForKey('blue', 'End')).toBe('gray')
  })

  it('移動に使わないキーでは undefined を返す', () => {
    expect(tagColorForKey('blue', 'Enter')).toBeUndefined()
  })
})
