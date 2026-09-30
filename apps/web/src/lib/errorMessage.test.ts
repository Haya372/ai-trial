import { describe, expect, it } from 'vitest'
import { hasErrorCode, mapErrorToMessage } from './errorMessage'

describe('hasErrorCode', () => {
  it('code プロパティを持つオブジェクトの場合 true を返す', () => {
    expect(hasErrorCode({ code: 'SOMETHING' })).toBe(true)
  })

  it('code プロパティを持たないオブジェクトの場合 false を返す', () => {
    expect(hasErrorCode({ message: 'oops' })).toBe(false)
  })

  it('null の場合 false を返す', () => {
    expect(hasErrorCode(null)).toBe(false)
  })

  it('オブジェクトでない場合 false を返す', () => {
    expect(hasErrorCode('error')).toBe(false)
  })
})

describe('mapErrorToMessage', () => {
  const t = (key: 'known' | 'fallback') =>
    key === 'known' ? 'known message' : 'fallback message'
  const codeToKey = { KNOWN_CODE: 'known' } as const

  it('codeToKey に含まれるコードの場合、対応するメッセージを返す', () => {
    expect(
      mapErrorToMessage({ code: 'KNOWN_CODE' }, t, codeToKey, 'fallback'),
    ).toBe('known message')
  })

  it('codeToKey に含まれないコードの場合、フォールバックメッセージを返す', () => {
    expect(
      mapErrorToMessage({ code: 'UNKNOWN_CODE' }, t, codeToKey, 'fallback'),
    ).toBe('fallback message')
  })

  it('code を持たないエラーの場合、フォールバックメッセージを返す', () => {
    expect(mapErrorToMessage(new Error('boom'), t, codeToKey, 'fallback')).toBe(
      'fallback message',
    )
  })
})
