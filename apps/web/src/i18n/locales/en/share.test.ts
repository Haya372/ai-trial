import { describe, expect, it } from 'vitest'
import enShare from './share.json'

// 日本語文字（ひらがな・カタカナ・漢字）を含む正規表現
const JAPANESE_PATTERN = /[぀-ヿ一-鿿]/

function collectValues(
  obj: unknown,
  path = '',
): { path: string; value: string }[] {
  if (typeof obj === 'string') return [{ path, value: obj }]
  if (typeof obj === 'object' && obj !== null) {
    return Object.entries(obj as Record<string, unknown>).flatMap(([k, v]) =>
      collectValues(v, path ? `${path}.${k}` : k),
    )
  }
  return []
}

describe('en/share.json', () => {
  it('すべての値に日本語文字が含まれていないこと', () => {
    const values = collectValues(enShare)
    const japanese = values.filter(({ value }) => JAPANESE_PATTERN.test(value))
    expect(japanese).toHaveLength(0)
  })
})
