import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { Button } from './Button'

function openingTag(html: string) {
  return html.match(/^<button[^>]*>/)?.[0] ?? ''
}

describe('Button', () => {
  describe('loading', () => {
    it('loading のとき無効化され aria-busy="true" が付与される', () => {
      const html = renderToStaticMarkup(<Button loading>保存</Button>)
      const tag = openingTag(html)

      expect(tag).toContain('disabled=""')
      expect(tag).toContain('aria-busy="true"')
    })

    it('loading のときスピナーを表示する', () => {
      const html = renderToStaticMarkup(<Button loading>保存</Button>)

      expect(html).toContain('data-slot="spinner"')
    })

    it('loading のときもラベルを残してボタン幅とアクセシブルネームを保つ', () => {
      const html = renderToStaticMarkup(<Button loading>保存</Button>)

      expect(html).toContain('保存')
      expect(html).toMatch(
        /<span[^>]*class="[^"]*opacity-0[^"]*"[^>]*>保存<\/span>/,
      )
    })

    it('loading でないときはスピナーも aria-busy も付かない', () => {
      const html = renderToStaticMarkup(<Button>保存</Button>)
      const tag = openingTag(html)

      expect(html).not.toContain('data-slot="spinner"')
      expect(tag).not.toContain('aria-busy')
      expect(tag).not.toContain('disabled=""')
    })

    it('loading でなくても disabled を指定すれば無効化される', () => {
      const html = renderToStaticMarkup(<Button disabled>保存</Button>)

      expect(openingTag(html)).toContain('disabled=""')
    })
  })
})
