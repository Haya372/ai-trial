import { describe, expect, it } from 'vitest'
import { eventClassNames } from './eventClassNames'
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

describe('eventClassNames', () => {
  it('isSubscribedがtrueの場合、破線ボーダーとopacityのクラスを返す', () => {
    const classNames = eventClassNames(baseEvent({ isSubscribed: true }))
    expect(classNames).toEqual(['opacity-70', 'border', 'border-dashed'])
  })

  it('isSubscribedがfalseの場合、クラスを返さない', () => {
    expect(eventClassNames(baseEvent({ isSubscribed: false }))).toEqual([])
  })

  it('isSubscribedが未指定の場合、クラスを返さない', () => {
    expect(eventClassNames(baseEvent())).toEqual([])
  })
})
