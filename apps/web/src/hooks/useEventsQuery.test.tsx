import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'
import { useEventsQuery } from './useEventsQuery'

vi.mock('../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/generated')>()
  return { ...actual, getEvents: vi.fn() }
})

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useEventsQuery', () => {
  const startDate = new Date(2026, 8, 1) // 2026-09-01
  const endDate = new Date(2026, 9, 1) // 2026-10-01（この日時を含まない）

  describe('UTC以外のタイムゾーン（JST）', () => {
    beforeAll(() => {
      vi.stubEnv('TZ', 'Asia/Tokyo')
    })
    afterAll(() => {
      vi.unstubAllEnvs()
    })

    it('範囲をUTCのRFC3339（末尾Z）に変換して startAt / endAt で送る', async () => {
      const { getEvents } = await import('../api/generated')
      vi.mocked(getEvents).mockResolvedValueOnce({
        data: { events: [] },
        status: 200,
        headers: new Headers(),
      } as never)

      const { result } = renderHook(
        () => useEventsQuery(new Date(2026, 10, 1), new Date(2026, 11, 1)),
        { wrapper: createWrapper() },
      )

      await waitFor(() => expect(result.current.isSuccess).toBe(true))
      expect(getEvents).toHaveBeenCalledWith({
        startAt: '2026-10-31T15:00:00.000Z',
        endAt: '2026-11-30T15:00:00.000Z',
      })
    })
  })

  it('成功時に events データを返す', async () => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockResolvedValueOnce({
      data: {
        events: [
          {
            id: 'event-1',
            title: 'テストイベント',
            startAt: '2026-09-13T10:00:00+09:00',
            endAt: '2026-09-13T11:00:00+09:00',
          },
        ],
      },
      status: 200,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useEventsQuery(startDate, endDate), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.events).toHaveLength(1)
    expect(result.current.data?.events[0].title).toBe('テストイベント')
  })

  it('tagId を渡すとリクエストに tagId を付ける', async () => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockResolvedValueOnce({
      data: { events: [] },
      status: 200,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(
      () => useEventsQuery(startDate, endDate, 'tag-1'),
      { wrapper: createWrapper() },
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(getEvents).toHaveBeenLastCalledWith({
      startAt: startDate.toISOString(),
      endAt: endDate.toISOString(),
      tagId: 'tag-1',
    })
  })

  it.each([
    ['省略', undefined],
    ['null', null],
  ])('tagId が%sのときはリクエストに tagId を含めない', async (_, tagId) => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockResolvedValueOnce({
      data: { events: [] },
      status: 200,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(
      () => useEventsQuery(startDate, endDate, tagId),
      { wrapper: createWrapper() },
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(vi.mocked(getEvents).mock.lastCall?.[0]).not.toHaveProperty('tagId')
  })

  describe('取得中に表示するデータ', () => {
    const previous = {
      events: [
        {
          id: 'event-1',
          title: '絞り込み前の予定',
          startAt: '2026-09-13T01:00:00.000Z',
          endAt: '2026-09-13T02:00:00.000Z',
        },
      ],
    }

    async function renderAndLoad() {
      const { getEvents } = await import('../api/generated')
      vi.mocked(getEvents).mockReset()
      vi.mocked(getEvents).mockResolvedValueOnce({
        data: previous,
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getEvents).mockImplementation(() => new Promise(() => {}))

      const hook = renderHook(
        ({
          start,
          end,
          tagId,
        }: {
          start: Date
          end: Date
          tagId: string | null
        }) => useEventsQuery(start, end, tagId),
        {
          wrapper: createWrapper(),
          initialProps: { start: startDate, end: endDate, tagId: null },
        },
      )
      await waitFor(() => expect(hook.result.current.isSuccess).toBe(true))
      return hook
    }

    it('同じ期間で tagId を切り替えたときは、取得が終わるまで前のデータを表示する', async () => {
      const { result, rerender } = await renderAndLoad()

      rerender({ start: startDate, end: endDate, tagId: 'tag-1' })

      expect(result.current.isPending).toBe(false)
      expect(result.current.isPlaceholderData).toBe(true)
      expect(result.current.data).toEqual(previous)
    })

    it('期間が変わったときは前の期間のデータを表示しない', async () => {
      const { result, rerender } = await renderAndLoad()

      rerender({
        start: new Date(2026, 9, 1),
        end: new Date(2026, 10, 1),
        tagId: null,
      })

      expect(result.current.isPending).toBe(true)
      expect(result.current.data).toBeUndefined()
    })
  })

  it('ネットワークエラー時に isError が true になる', async () => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useEventsQuery(startDate, endDate), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
  })

  it('未認証(401)レスポンス時に isError が true になる', async () => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockResolvedValueOnce({
      data: { code: 'UNAUTHORIZED', message: 'Unauthorized' },
      status: 401,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useEventsQuery(startDate, endDate), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
  })

  it('ローディング中は isPending が true になる', async () => {
    const { getEvents } = await import('../api/generated')
    // 解決しないPromiseで確認
    vi.mocked(getEvents).mockImplementationOnce(() => new Promise(() => {}))

    const { result } = renderHook(() => useEventsQuery(startDate, endDate), {
      wrapper: createWrapper(),
    })

    expect(result.current.isPending).toBe(true)
  })

  it('queryKey に日付範囲が含まれる（日付が違えば別キャッシュになる）', async () => {
    const { getEvents } = await import('../api/generated')
    vi.mocked(getEvents).mockResolvedValue({
      data: { events: [] },
      status: 200,
      headers: new Headers(),
    } as never)

    const startDate2 = new Date(2026, 9, 1) // 2026-10-01
    const endDate2 = new Date(2026, 10, 1) // 2026-11-01

    let callCount = 0
    vi.mocked(getEvents).mockImplementation(async () => {
      callCount++
      return {
        data: { events: [] },
        status: 200,
        headers: new Headers(),
      } as never
    })

    const wrapper = createWrapper()
    const { result: result1 } = renderHook(
      () => useEventsQuery(startDate, endDate),
      { wrapper },
    )
    const { result: result2 } = renderHook(
      () => useEventsQuery(startDate2, endDate2),
      { wrapper },
    )

    await waitFor(() => {
      expect(result1.current.isSuccess).toBe(true)
      expect(result2.current.isSuccess).toBe(true)
    })
    // 2つの異なる日付範囲で2回 API が呼ばれる
    expect(callCount).toBe(2)
  })
})
