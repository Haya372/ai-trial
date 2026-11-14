import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, getEvents: vi.fn() }
})

// 分の境界・タブ再表示による更新は useNow のテストで検証済みのため、
// ここでは now を直接差し替えてパネルの振る舞いだけを検証する
vi.mock('../hooks/useNow', () => ({
  useNow: vi.fn(),
}))

import { getEvents } from '../../../api/generated'
import { useNow } from '../hooks/useNow'
import TodayTimelinePanel from './TodayTimelinePanel'

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

function event(overrides: Partial<EventResponse>): EventResponse {
  return {
    id: 'event-1',
    title: '定例MTG',
    description: null,
    startAt: new Date(2026, 10, 14, 10, 30).toISOString(),
    endAt: new Date(2026, 10, 14, 11, 30).toISOString(),
    location: null,
    url: null,
    isSubscribed: false,
    subscriptionId: null,
    tags: [],
    ...overrides,
  }
}

function resolveEvents(events: EventResponse[]) {
  vi.mocked(getEvents).mockResolvedValue({
    data: { events },
    status: 200,
    headers: new Headers(),
  } as never)
}

describe('TodayTimelinePanel', () => {
  beforeEach(() => {
    vi.mocked(getEvents).mockReset()
    vi.mocked(useNow).mockReturnValue(new Date(2026, 10, 14, 10, 5))
  })

  it('今日のローカル0時〜翌日0時をUTCで指定し、タグを指定せずに予定を取得する', async () => {
    resolveEvents([])

    render(<TodayTimelinePanel onEventClick={vi.fn()} />, {
      wrapper: createWrapper(),
    })

    await screen.findByText('今日の予定はありません')
    expect(getEvents).toHaveBeenCalledWith({
      startAt: new Date(2026, 10, 14).toISOString(),
      endAt: new Date(2026, 10, 15).toISOString(),
    })
  })

  it('取得した予定と次の予定までの時間を表示する', async () => {
    resolveEvents([event({})])

    render(<TodayTimelinePanel onEventClick={vi.fn()} />, {
      wrapper: createWrapper(),
    })

    expect(await screen.findByText('あと25分')).toBeInTheDocument()
    expect(screen.getAllByText('定例MTG').length).toBeGreaterThan(0)
  })

  it('取得中はパネル内にローディングを表示する', () => {
    vi.mocked(getEvents).mockReturnValue(new Promise(() => {}))

    render(<TodayTimelinePanel onEventClick={vi.fn()} />, {
      wrapper: createWrapper(),
    })

    expect(screen.getByText('読み込み中...')).toBeInTheDocument()
  })

  it('取得に失敗したらパネル内にエラーメッセージを表示する', async () => {
    vi.mocked(getEvents).mockResolvedValue({
      data: { code: 'INTERNAL_ERROR', message: 'error' },
      status: 500,
      headers: new Headers(),
    } as never)

    render(<TodayTimelinePanel onEventClick={vi.fn()} />, {
      wrapper: createWrapper(),
    })

    expect(
      await screen.findByText('今日の予定を読み込めませんでした'),
    ).toBeInTheDocument()
  })

  it('日付が変わったら新しい日の範囲で予定を取得し直す', async () => {
    resolveEvents([])
    const { rerender } = render(<TodayTimelinePanel onEventClick={vi.fn()} />, {
      wrapper: createWrapper(),
    })
    await screen.findByText('今日の予定はありません')

    vi.mocked(useNow).mockReturnValue(new Date(2026, 10, 15, 0, 0))
    rerender(<TodayTimelinePanel onEventClick={vi.fn()} />)

    await vi.waitFor(() =>
      expect(getEvents).toHaveBeenLastCalledWith({
        startAt: new Date(2026, 10, 15).toISOString(),
        endAt: new Date(2026, 10, 16).toISOString(),
      }),
    )
  })

  it('予定をクリックすると取得済みの EventResponse を通知する', async () => {
    const target = event({ id: 'event-2', title: '1on1', location: '会議室A' })
    resolveEvents([event({}), target])
    const onEventClick = vi.fn()

    render(<TodayTimelinePanel onEventClick={onEventClick} />, {
      wrapper: createWrapper(),
    })

    fireEvent.click(await screen.findByRole('button', { name: /1on1/ }))
    expect(onEventClick).toHaveBeenCalledWith(target)
  })
})
