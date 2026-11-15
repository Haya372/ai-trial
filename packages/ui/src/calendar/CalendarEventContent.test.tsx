import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { CalendarEventContent } from './CalendarEventContent'

describe('CalendarEventContent', () => {
  it('タイトルを表示する', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent timeText="" title="チームミーティング" />,
    )
    expect(html).toContain('チームミーティング')
  })

  it('時刻テキストを表示する', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent timeText="10:00" title="ミーティング" />,
    )
    expect(html).toContain('10:00')
  })

  it('tagColorsが未指定の場合、ドットコンテナを表示しない', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent timeText="10:00" title="ミーティング" />,
    )
    expect(html).not.toContain('calendar-event-tag-dots')
  })

  it('tagColorsが1個の場合、ドットコンテナを表示しない（先頭タグは背景色）', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent
        timeText="10:00"
        title="ミーティング"
        tagColors={['red']}
      />,
    )
    expect(html).not.toContain('calendar-event-tag-dots')
  })

  it('tagColorsが3個の場合、2番目以降の2個のドットを付与順に描画する', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent
        timeText="10:00"
        title="ミーティング"
        tagColors={['red', 'blue', 'green']}
      />,
    )
    // 2番目: blue, 3番目: green の順でドットが並ぶ
    const blueIdx = html.indexOf('var(--tag-blue-bg)')
    const greenIdx = html.indexOf('var(--tag-green-bg)')
    expect(blueIdx).toBeGreaterThan(-1)
    expect(greenIdx).toBeGreaterThan(-1)
    expect(blueIdx).toBeLessThan(greenIdx)
  })

  it('ドット列コンテナがaria-hidden="true"である', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent
        timeText="10:00"
        title="ミーティング"
        tagColors={['red', 'blue']}
      />,
    )
    expect(html).toContain('data-slot="calendar-event-tag-dots"')
    expect(html).toContain('aria-hidden="true"')
  })

  it('ドット列コンテナに省略用クラス（1行分の高さ・flex-wrap・overflow-hidden）が付いている', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent
        timeText="10:00"
        title="ミーティング"
        tagColors={['red', 'blue']}
      />,
    )
    // jsdomなしのためclass属性で検証
    // 高さを1行分に固定しないと、折り返したドットが隠れず省略にならない
    expect(html).toMatch(
      /data-slot="calendar-event-tag-dots"[^>]*class="[^"]*\bh-2\b/,
    )
    expect(html).toContain('overflow-hidden')
    expect(html).toContain('flex-wrap')
  })

  it('timeClass・titleClassを時刻・タイトルの要素に適用する', () => {
    const html = renderToStaticMarkup(
      <CalendarEventContent
        timeText="10:00"
        title="ミーティング"
        timeClass="fc-time"
        titleClass="fc-title"
      />,
    )
    expect(html).toContain('<span class="fc-time">10:00</span>')
    expect(html).toContain('<span class="fc-title">ミーティング</span>')
  })
})
