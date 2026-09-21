import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { CalendarState } from '../../../store/calendarStore'

// MonthCalendar/WeekCalendar はFullCalendarラッパーなのでモック
vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    MonthCalendar: vi.fn(() => <div data-testid="month-calendar" />),
    WeekCalendar: vi.fn(() => <div data-testid="week-calendar" />),
  }
})

// useEventsQuery をモック
vi.mock('../../../hooks/useEventsQuery', () => ({
  useEventsQuery: vi.fn(),
}))

// calendarStore をモック（初期値: month ビュー）
vi.mock('../../../store/calendarStore', () => ({
  useCalendarStore: vi.fn((selector: (state: CalendarState) => unknown) => {
    const state: CalendarState = {
      view: 'month',
      currentDate: new Date(2026, 8, 13),
      setView: vi.fn(),
      setCurrentDate: vi.fn(),
    }
    return selector ? selector(state) : state
  }),
}))

import CalendarPage from './CalendarPage'

function createWrapper() {
  const queryClient = new QueryClient()
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('CalendarPage', () => {
  describe('ローディング状態', () => {
    it('isPending のとき「読み込み中...」を表示する', async () => {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: undefined,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })
      expect(screen.getByText(/読み込み中/)).toBeInTheDocument()
    })
  })

  describe('エラー状態', () => {
    it('isError のとき「予定を読み込めませんでした」を表示する', async () => {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: undefined,
        isPending: false,
        isError: true,
        isSuccess: false,
        error: new Error('fetch error'),
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })
      expect(screen.getByText(/予定を読み込めませんでした/)).toBeInTheDocument()
    })
  })

  describe('月ビュー表示', () => {
    it('view が "month" のとき MonthCalendar を表示する', async () => {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })
      expect(screen.getByTestId('month-calendar')).toBeInTheDocument()
      expect(screen.queryByTestId('week-calendar')).not.toBeInTheDocument()
    })
  })

  describe('週ビュー表示', () => {
    it('view が "week" のとき WeekCalendar を表示する', async () => {
      const { useCalendarStore } = await import('../../../store/calendarStore')
      vi.mocked(useCalendarStore).mockImplementation(
        (selector: (state: CalendarState) => unknown) => {
          const state: CalendarState = {
            view: 'week',
            currentDate: new Date(2026, 8, 13),
            setView: vi.fn(),
            setCurrentDate: vi.fn(),
          }
          return selector ? selector(state) : state
        },
      )

      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })
      expect(screen.getByTestId('week-calendar')).toBeInTheDocument()
      expect(screen.queryByTestId('month-calendar')).not.toBeInTheDocument()
    })
  })
})
