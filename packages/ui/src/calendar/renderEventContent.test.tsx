import type { EventDisplayInfo } from '@fullcalendar/react'
import { isValidElement } from 'react'
import { describe, expect, it } from 'vitest'
import { CalendarEventContent } from './CalendarEventContent'
import { renderEventContent } from './renderEventContent'
import type { CalendarEvent } from './types'

function displayInfo(
  calendarEvent: CalendarEvent | undefined,
): EventDisplayInfo {
  return {
    event: {
      title: calendarEvent?.title ?? 'External event',
      extendedProps: calendarEvent ? { calendarEvent } : {},
    },
    timeText: '10:00',
    timeClass: 'fc-time',
    titleClass: 'fc-title',
  } as unknown as EventDisplayInfo
}

function baseEvent(overrides: Partial<CalendarEvent> = {}): CalendarEvent {
  return {
    id: 'event-1',
    title: 'Team meeting',
    start: new Date(2026, 8, 13, 10, 0),
    end: new Date(2026, 8, 13, 11, 0),
    ...overrides,
  }
}

describe('renderEventContent', () => {
  it('tagColorsが未指定の場合、デフォルト描画に委ねるためtrueを返す', () => {
    expect(renderEventContent(displayInfo(baseEvent()))).toBe(true)
  })

  it('tagColorsが空配列の場合、trueを返す', () => {
    expect(renderEventContent(displayInfo(baseEvent({ tagColors: [] })))).toBe(
      true,
    )
  })

  it('extendedPropsにcalendarEventがない場合、trueを返す', () => {
    expect(renderEventContent(displayInfo(undefined))).toBe(true)
  })

  it('タグ付きの場合、CalendarEventContentに表示情報とtagColorsを渡す', () => {
    const result = renderEventContent(
      displayInfo(baseEvent({ tagColors: ['red', 'blue'] })),
    )
    expect(isValidElement(result)).toBe(true)
    if (!isValidElement(result)) return
    expect(result.type).toBe(CalendarEventContent)
    expect(result.props).toEqual({
      timeText: '10:00',
      title: 'Team meeting',
      tagColors: ['red', 'blue'],
      timeClass: 'fc-time',
      titleClass: 'fc-title',
    })
  })
})
