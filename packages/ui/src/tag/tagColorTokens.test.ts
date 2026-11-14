import { describe, expect, it } from 'vitest'
import tagColorsCss from './tagColors.css?raw'
import { TAG_COLOR_KEYS, tagColorVars } from './tagColorTokens'

function relativeLuminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const channel = Number.parseInt(hex.slice(i, i + 2), 16) / 255
    return channel <= 0.04045
      ? channel / 12.92
      : ((channel + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

function contrastRatio(a: string, b: string): number {
  const [lighter, darker] = [relativeLuminance(a), relativeLuminance(b)].sort(
    (x, y) => y - x,
  )
  return (lighter + 0.05) / (darker + 0.05)
}

function tokenValue(name: string): string | undefined {
  return tagColorsCss.match(new RegExp(`--${name}:\\s*([^;]+);`))?.[1].trim()
}

describe('TAG_COLOR_KEYS', () => {
  it('SPEC-008 の表示順で12色を並べる', () => {
    expect(TAG_COLOR_KEYS).toEqual([
      'red',
      'orange',
      'yellow',
      'green',
      'teal',
      'cyan',
      'blue',
      'indigo',
      'purple',
      'pink',
      'brown',
      'gray',
    ])
  })
})

describe('tagColorVars', () => {
  it('色キーに対応する背景色と文字色の CSS 変数を返す', () => {
    expect(tagColorVars('teal')).toEqual({
      background: 'var(--tag-teal-bg)',
      foreground: 'var(--tag-teal-fg)',
    })
  })
})

describe('tagColors.css のタグ色定義', () => {
  it.each(TAG_COLOR_KEYS)('%s の背景色と文字色を HEX で定義する', (key) => {
    expect(tokenValue(`tag-${key}-bg`)).toMatch(/^#[0-9a-f]{6}$/i)
    expect(tokenValue(`tag-${key}-fg`)).toMatch(/^#[0-9a-f]{6}$/i)
  })

  it.each(TAG_COLOR_KEYS)(
    '%s の背景色と文字色のコントラスト比が WCAG AA（4.5:1）以上',
    (key) => {
      const background = tokenValue(`tag-${key}-bg`) ?? ''
      const foreground = tokenValue(`tag-${key}-fg`) ?? ''
      expect(contrastRatio(background, foreground)).toBeGreaterThanOrEqual(4.5)
    },
  )
})
