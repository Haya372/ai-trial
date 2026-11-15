import { describe, expect, it } from 'vitest'
import { eventColorProps } from './eventColorProps'
import type { CalendarEvent } from './types'

function baseEvent(overrides: Partial<CalendarEvent> = {}): CalendarEvent {
  return {
    id: 'event-1',
    title: 'Team meeting',
    start: new Date(2026, 8, 13, 10, 0),
    end: new Date(2026, 8, 13, 11, 0),
    ...overrides,
  }
}

describe('eventColorProps', () => {
  it('tagColorsが未指定の場合、空オブジェクトを返す', () => {
    expect(eventColorProps(baseEvent())).toEqual({})
  })

  it('tagColorsが空配列の場合、空オブジェクトを返す', () => {
    expect(eventColorProps(baseEvent({ tagColors: [] }))).toEqual({})
  })

  it('tagColorsが1個の場合、先頭タグのbg/fgのCSS変数を返す', () => {
    const result = eventColorProps(baseEvent({ tagColors: ['red'] }))
    expect(result).toEqual({
      color: 'var(--tag-red-bg)',
      contrastColor: 'var(--tag-red-fg)',
    })
  })

  it('tagColorsが複数の場合、先頭タグの色のみ使う', () => {
    const result = eventColorProps(
      baseEvent({ tagColors: ['blue', 'green', 'red'] }),
    )
    expect(result).toEqual({
      color: 'var(--tag-blue-bg)',
      contrastColor: 'var(--tag-blue-fg)',
    })
  })
})
