import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { Spinner } from './Spinner'

describe('Spinner', () => {
  it('label を渡すと role="status" とアクセシブルネームを持つ', () => {
    const html = renderToStaticMarkup(<Spinner label="読み込み中" />)

    expect(html).toContain('role="status"')
    expect(html).toContain('aria-label="読み込み中"')
    expect(html).not.toContain('aria-hidden')
  })

  it('label を省略すると装飾扱いになり支援技術から隠れる', () => {
    const html = renderToStaticMarkup(<Spinner />)

    expect(html).toContain('aria-hidden="true"')
    expect(html).not.toContain('role="status"')
  })

  it.each([
    ['sm', 'size-4'],
    ['md', 'size-5'],
    ['lg', 'size-8'],
  ] as const)('size=%s のとき %s クラスを持つ', (size, className) => {
    const html = renderToStaticMarkup(<Spinner size={size} />)

    expect(html).toContain(className)
  })

  it('size を省略すると md になる', () => {
    const html = renderToStaticMarkup(<Spinner />)

    expect(html).toContain('size-5')
  })
})
