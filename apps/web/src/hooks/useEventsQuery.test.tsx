import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
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
  const endDate = new Date(2026, 8, 30) // 2026-09-30

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
    const endDate2 = new Date(2026, 9, 31) // 2026-10-31

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
