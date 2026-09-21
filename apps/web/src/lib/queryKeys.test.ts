import { describe, expect, it } from 'vitest'
import { eventsKeys } from './queryKeys'

describe('eventsKeys', () => {
  it('list は all の末尾に startDate/endDate を連結したキーを返す', () => {
    expect(eventsKeys.list('2026-09-01', '2026-09-30')).toEqual([
      'events',
      '2026-09-01',
      '2026-09-30',
    ])
  })

  it('startDate/endDate が異なれば異なるキーを返す', () => {
    const key1 = eventsKeys.list('2026-09-01', '2026-09-30')
    const key2 = eventsKeys.list('2026-10-01', '2026-10-31')
    expect(key1).not.toEqual(key2)
  })
})
