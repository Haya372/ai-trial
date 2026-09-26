import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { CalendarState } from '../../../store/calendarStore'

// このファイルはCalendarPage.test.tsxと異なりQuickRegistrationPanelを
// モックしない。パネル自身の再マウント（内部状態のリセット）まで含めて
// 検証したいため。

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, createEvent: vi.fn() }
})

vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    MonthCalendar: vi.fn(() => <div data-testid="month-calendar" />),
    WeekCalendar: vi.fn(() => <div data-testid="week-calendar" />),
    toast: Object.assign(vi.fn(), { error: vi.fn(), success: vi.fn() }),
  }
})

vi.mock('../../../hooks/useEventsQuery', () => ({
  useEventsQuery: vi.fn(),
}))

vi.mock('../../../store/calendarStore', () => ({
  useCalendarStore: vi.fn((selector: (state: CalendarState) => unknown) => {
    const state: CalendarState = {
      view: 'week',
      currentDate: new Date(2026, 8, 13),
      setView: vi.fn(),
      setCurrentDate: vi.fn(),
    }
    return selector ? selector(state) : state
  }),
}))

import CalendarPage from './CalendarPage'

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

function getLatestCallProps<P>(mockFn: (props: P) => unknown): P {
  const call = vi.mocked(mockFn).mock.calls.at(-1)
  if (!call) throw new Error('mock function was not called')
  return call[0]
}

describe('CalendarPage クイック登録パネルの再オープン', () => {
  it('登録完了後に同じ時間帯セルを再度クリックすると入力フォームに戻る', async () => {
    const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
    vi.mocked(useEventsQuery).mockReturnValue({
      data: { events: [] },
      isPending: false,
      isError: false,
      isSuccess: true,
      error: null,
    } as never)

    const { createEvent } = await import('../../../api/generated')
    vi.mocked(createEvent).mockResolvedValue({
      data: {
        id: 'new-event',
        title: '仮予定',
        description: null,
        location: null,
        url: null,
        startAt: new Date(2026, 8, 15, 10, 0).toISOString(),
        endAt: new Date(2026, 8, 15, 11, 0).toISOString(),
      },
      status: 201,
      headers: new Headers(),
    } as never)

    const { WeekCalendar } = await import('@repo/ui')
    render(<CalendarPage />, { wrapper: createWrapper() })

    const props = getLatestCallProps(WeekCalendar)
    const slot = new Date(2026, 8, 15, 10, 0)

    act(() => {
      props.onTimeSlotClick(slot, { x: 10, y: 20 })
    })

    fireEvent.change(screen.getByLabelText('タイトル'), {
      target: { value: '仮予定' },
    })
    fireEvent.click(screen.getByRole('button', { name: '保存' }))

    await screen.findByRole('button', { name: '詳細を編集' })

    act(() => {
      props.onTimeSlotClick(slot, { x: 10, y: 20 })
    })

    expect(screen.getByLabelText('タイトル')).toHaveValue('')
    expect(
      screen.queryByRole('button', { name: '詳細を編集' }),
    ).not.toBeInTheDocument()
  })
})
