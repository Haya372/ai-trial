import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { Loading } from './Loading'

describe('Loading', () => {
  it('role="status" のライブリージョンとしてラベルを表示する', () => {
    const html = renderToStaticMarkup(<Loading label="読み込み中..." />)

    expect(html).toMatch(/^<div[^>]*role="status"/)
    expect(html).toContain('aria-live="polite"')
    expect(html).toContain('読み込み中...')
  })

  it('スピナーは装飾扱いにしてラベルの重複読み上げを防ぐ', () => {
    const html = renderToStaticMarkup(<Loading label="読み込み中..." />)

    expect(html).toContain('data-slot="spinner"')
    expect(html).toContain('aria-hidden="true"')
  })

  it('label を省略するとスピナーのみ表示する', () => {
    const html = renderToStaticMarkup(<Loading />)

    expect(html).toContain('role="status"')
    expect(html).toContain('data-slot="spinner"')
    expect(html).not.toContain('data-slot="loading-label"')
  })
})
