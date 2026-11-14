import { describe, expect, it } from 'vitest'
import type { CalendarEvent } from '../calendar/types'
import { buildTimeline, type TimelineItem } from './buildTimeline'

const at = (hours: number, minutes = 0, seconds = 0) =>
  new Date(2026, 10, 14, hours, minutes, seconds)

function event(
  id: string,
  start: Date,
  end: Date,
  overrides: Partial<CalendarEvent> = {},
): CalendarEvent {
  return { id, title: `event-${id}`, start, end, ...overrides }
}

// Compact representation of items to keep assertions on ordering readable.
function shape(items: TimelineItem[]): string[] {
  return items.map((item) => {
    if (item.type === 'event') return item.event.id
    if (item.type === 'gap') return `gap:${item.minutes}`
    return 'now'
  })
}

describe('buildTimeline', () => {
  describe('並び順', () => {
    it('開始日時の昇順、同じ場合は終了日時の昇順で並べる', () => {
      const timeline = buildTimeline(
        [
          event('c', at(13), at(14)),
          event('b', at(10), at(12)),
          event('a', at(10), at(11)),
        ],
        at(8),
      )
      expect(
        timeline.items.flatMap((i) => (i.type === 'event' ? [i.event.id] : [])),
      ).toEqual(['a', 'b', 'c'])
    })
  })

  describe('今日の範囲', () => {
    it('今日と重ならない予定を除外する', () => {
      const timeline = buildTimeline(
        [
          event(
            'yesterday',
            new Date(2026, 10, 13, 10),
            new Date(2026, 10, 13, 11),
          ),
          event('endsAtMidnight', new Date(2026, 10, 13, 23), at(0)),
          event(
            'tomorrow',
            new Date(2026, 10, 15, 0),
            new Date(2026, 10, 15, 1),
          ),
          event('today', at(10), at(11)),
        ],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['now', 'today'])
    })

    it('前日から翌日まで続く予定を今日の範囲に切り詰め、フラグを立てる', () => {
      const timeline = buildTimeline(
        [event('long', new Date(2026, 10, 13, 22), new Date(2026, 10, 15, 2))],
        at(8),
      )
      const item = timeline.items.find((i) => i.type === 'event')
      expect(item).toMatchObject({
        start: at(0),
        end: new Date(2026, 10, 15, 0),
        startsBeforeToday: true,
        endsAfterToday: true,
      })
    })

    it('今日の範囲内の予定はフラグを立てない', () => {
      const timeline = buildTimeline([event('a', at(10), at(11))], at(8))
      expect(timeline.items.find((i) => i.type === 'event')).toMatchObject({
        startsBeforeToday: false,
        endsAfterToday: false,
      })
    })

    it('0時ちょうどに始まる開始＝終了の予定を含める', () => {
      const timeline = buildTimeline([event('point', at(0), at(0))], at(8))
      expect(shape(timeline.items)).toEqual(['point', 'now'])
    })
  })

  describe('予定の有無', () => {
    it('予定が0件ならインジケーターだけを返し、hasEventsはfalse', () => {
      const timeline = buildTimeline([], at(9, 30, 45))
      expect(timeline.hasEvents).toBe(false)
      expect(timeline.nextEvent).toBeNull()
      expect(timeline.items).toEqual([{ type: 'now', time: at(9, 30) }])
    })

    it('予定があればhasEventsはtrue', () => {
      expect(buildTimeline([event('a', at(10), at(11))], at(8)).hasEvents).toBe(
        true,
      )
    })
  })

  describe('次の予定', () => {
    it('開始日時が現在時刻より後の最初の予定と、あと何分かを返す', () => {
      const next = event('next', at(10, 30), at(11))
      const timeline = buildTimeline(
        [event('past', at(9), at(9, 30)), next, event('later', at(12), at(13))],
        at(10, 5),
      )
      expect(timeline.nextEvent).toEqual({ event: next, minutesUntil: 25 })
    })

    it('現在時刻を分単位に切り捨ててから、開始までの分数を切り上げる', () => {
      const next = event('next', at(10, 0, 30), at(11))
      const timeline = buildTimeline([next], at(9, 59, 50))
      // nowMinute = 9:59 → 10:00:30 まで 1.5 分 → 切り上げて 2 分
      expect(timeline.nextEvent?.minutesUntil).toBe(2)
    })

    it('開始日時がちょうど現在時刻（分単位）の予定は次の予定ではなく開催中とする', () => {
      const timeline = buildTimeline(
        [event('starting', at(10), at(11))],
        at(10, 0, 40),
      )
      expect(timeline.nextEvent).toBeNull()
      expect(timeline.items.find((i) => i.type === 'event')).toMatchObject({
        isOngoing: true,
      })
    })

    it('残りの予定がなければnullを返す', () => {
      const timeline = buildTimeline([event('a', at(9), at(10))], at(11))
      expect(timeline.nextEvent).toBeNull()
    })
  })

  describe('開催中', () => {
    it('開始日時 ≦ 現在時刻 < 終了日時の予定を開催中とする', () => {
      const timeline = buildTimeline(
        [
          event('ended', at(8), at(10)),
          event('ongoing', at(9), at(11)),
          event('future', at(12), at(13)),
        ],
        at(10),
      )
      const ongoing = timeline.items.flatMap((i) =>
        i.type === 'event' && i.isOngoing ? [i.event.id] : [],
      )
      expect(ongoing).toEqual(['ongoing'])
    })

    it('開始＝終了の予定は開催中にしない', () => {
      const timeline = buildTimeline([event('point', at(10), at(10))], at(10))
      expect(timeline.items.find((i) => i.type === 'event')).toMatchObject({
        isOngoing: false,
      })
    })

    it('前日から続く予定も開催中と判定する', () => {
      const timeline = buildTimeline(
        [event('overnight', new Date(2026, 10, 13, 23), at(1))],
        at(0, 30),
      )
      expect(timeline.items.find((i) => i.type === 'event')).toMatchObject({
        isOngoing: true,
      })
    })
  })

  describe('空き時間と現在時刻インジケーター', () => {
    it('連続する予定の間に空き時間を挿入し、最初の前と最後の後には挿入しない', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(10)), event('b', at(10, 30), at(11))],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['now', 'a', 'gap:30', 'b'])
    })

    it('重なっている予定の間には空き時間を挿入しない', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(11)), event('b', at(10), at(12))],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['now', 'a', 'b'])
    })

    it('隣接する予定（終了＝次の開始）の間には空き時間を挿入しない', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(10)), event('b', at(10), at(11))],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['now', 'a', 'b'])
    })

    it('先行する予定までの終了日時の最大値から空き時間を計算する', () => {
      const timeline = buildTimeline(
        [
          event('long', at(9), at(12)),
          event('short', at(9, 30), at(10)),
          event('after', at(12, 15), at(13)),
        ],
        at(8),
      )
      expect(shape(timeline.items)).toEqual([
        'now',
        'long',
        'short',
        'gap:15',
        'after',
      ])
    })

    it('開始＝終了の予定は開始日時＝終了日時として空き時間を計算する', () => {
      const timeline = buildTimeline(
        [event('point', at(9), at(9)), event('b', at(9, 45), at(10))],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['now', 'point', 'gap:45', 'b'])
    })

    it('前日から続く予定は今日の範囲で切り詰めた日時で空き時間を計算する', () => {
      const timeline = buildTimeline(
        [
          event('overnight', new Date(2026, 10, 13, 23), at(1)),
          event('b', at(2), at(3)),
        ],
        at(8),
      )
      expect(shape(timeline.items)).toEqual(['overnight', 'gap:60', 'b', 'now'])
    })

    it('インジケーターは開始済みの予定の後、未開始の予定の前に置く', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(9, 30)), event('b', at(9, 30), at(11))],
        at(9, 45),
      )
      expect(shape(timeline.items)).toEqual(['a', 'b', 'now'])
    })

    it('現在時刻が空き時間中ならインジケーターを空き時間の後に置く', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(10)), event('b', at(11), at(12))],
        at(10, 20),
      )
      expect(shape(timeline.items)).toEqual(['a', 'gap:60', 'now', 'b'])
    })

    it('先行する予定が開催中ならインジケーターを空き時間の前に置く', () => {
      const timeline = buildTimeline(
        [event('a', at(9), at(10)), event('b', at(11), at(12))],
        at(9, 20),
      )
      expect(shape(timeline.items)).toEqual(['a', 'now', 'gap:60', 'b'])
    })

    it('インジケーターの時刻は分単位に切り捨てる', () => {
      const timeline = buildTimeline([], at(9, 20, 59))
      expect(timeline.items).toEqual([{ type: 'now', time: at(9, 20) }])
    })
  })
})
