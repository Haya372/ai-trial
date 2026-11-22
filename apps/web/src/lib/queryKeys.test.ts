import { describe, expect, it } from 'vitest'
import { eventsKeys, sharesKeys, tagsKeys } from './queryKeys'

describe('eventsKeys', () => {
  it('list は all の末尾に "list" と startAt/endAt と tagId（なしは null）を連結したキーを返す', () => {
    expect(
      eventsKeys.list('2026-08-31T15:00:00.000Z', '2026-09-30T15:00:00.000Z'),
    ).toEqual([
      'events',
      'list',
      '2026-08-31T15:00:00.000Z',
      '2026-09-30T15:00:00.000Z',
      null,
    ])
  })

  it('tagId が異なれば異なるキーを返す', () => {
    const range = [
      '2026-08-31T15:00:00.000Z',
      '2026-09-30T15:00:00.000Z',
    ] as const
    expect(eventsKeys.list(...range, 'tag-1')).toEqual([
      'events',
      'list',
      ...range,
      'tag-1',
    ])
    expect(eventsKeys.list(...range, 'tag-1')).not.toEqual(
      eventsKeys.list(...range),
    )
  })

  it('tagId に null を渡すと省略したときと同じキーを返す', () => {
    const range = [
      '2026-08-31T15:00:00.000Z',
      '2026-09-30T15:00:00.000Z',
    ] as const
    expect(eventsKeys.list(...range, null)).toEqual(eventsKeys.list(...range))
  })

  it('startAt/endAt が異なれば異なるキーを返す', () => {
    const key1 = eventsKeys.list(
      '2026-08-31T15:00:00.000Z',
      '2026-09-30T15:00:00.000Z',
    )
    const key2 = eventsKeys.list(
      '2026-09-30T15:00:00.000Z',
      '2026-10-31T15:00:00.000Z',
    )
    expect(key1).not.toEqual(key2)
  })
})

describe('eventsKeys.parseList', () => {
  it('list で作ったキーから startAt/endAt/tagId を取り出す', () => {
    const key = eventsKeys.list(
      '2026-08-31T15:00:00.000Z',
      '2026-09-30T15:00:00.000Z',
      'tag-1',
    )
    expect(eventsKeys.parseList(key)).toEqual({
      startAt: '2026-08-31T15:00:00.000Z',
      endAt: '2026-09-30T15:00:00.000Z',
      tagId: 'tag-1',
    })
  })

  it('list 以外のキーでは undefined を返す', () => {
    expect(eventsKeys.parseList(eventsKeys.all)).toBeUndefined()
    expect(eventsKeys.parseList(['tags'])).toBeUndefined()
  })

  it('先頭が "events" でも "list" の区別用の要素がないキーでは undefined を返す', () => {
    expect(
      eventsKeys.parseList(['events', 'detail', 'event-1', 'extra']),
    ).toBeUndefined()
    expect(
      eventsKeys.parseList([
        'events',
        'detail',
        '2026-08-31T15:00:00.000Z',
        '2026-09-30T15:00:00.000Z',
        'tag-1',
      ]),
    ).toBeUndefined()
  })
})

describe('sharesKeys', () => {
  it('all は ["shares"] を返す', () => {
    expect(sharesKeys.all).toEqual(['shares'])
  })

  it('detail は all の末尾に token を連結したキーを返す', () => {
    expect(sharesKeys.detail('abc123')).toEqual(['shares', 'abc123'])
  })

  it('token が異なれば異なるキーを返す', () => {
    const key1 = sharesKeys.detail('token1')
    const key2 = sharesKeys.detail('token2')
    expect(key1).not.toEqual(key2)
  })
})

describe('tagsKeys', () => {
  it('all は ["tags"] を返す', () => {
    expect(tagsKeys.all).toEqual(['tags'])
  })
})
