import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
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

// EventDetailModal/EventFormModal はそれぞれ個別にテスト済みのためスタブ化し、
// CalendarPage の状態管理（どのイベントをどのモードで開くか）だけを検証する
vi.mock('../components/EventDetailModal', () => ({
  default: vi.fn(({ open, event, onEdit }) =>
    open && event ? (
      <div data-testid="event-detail-modal">
        <button onClick={() => onEdit(event)}>編集-{event.id}</button>
      </div>
    ) : null,
  ),
}))

vi.mock('../../event/components/EventFormModal', () => ({
  default: vi.fn(({ open, mode, event }) =>
    open ? (
      <div data-testid="event-form-modal">
        mode:{mode} event:{event?.id ?? 'none'}
      </div>
    ) : null,
  ),
}))

import CalendarPage from './CalendarPage'

function createWrapper() {
  const queryClient = new QueryClient()
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

function getLatestCallProps<P>(mockFn: (props: P) => unknown): P {
  const call = vi.mocked(mockFn).mock.calls.at(-1)
  if (!call) throw new Error('mock function was not called')
  return call[0]
}

const fullEvent: EventResponse = {
  id: 'event-1',
  title: 'デザインレビュー',
  description: 'メモ',
  startAt: new Date(2026, 8, 13, 10, 0).toISOString(),
  endAt: new Date(2026, 8, 13, 11, 0).toISOString(),
  location: '会議室A',
  url: null,
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

  describe('予定の新規作成', () => {
    it('「新規作成」ボタンをクリックするとcreateモードのフォームモーダルを開く', async () => {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })
      fireEvent.click(screen.getByRole('button', { name: '新規作成' }))

      const modal = screen.getByTestId('event-form-modal')
      expect(modal).toHaveTextContent('mode:create')
      expect(modal).toHaveTextContent('event:none')
    })
  })

  describe('予定の詳細表示・編集', () => {
    it('カレンダー上のイベントクリックで対応するEventResponseを詳細モーダルに渡す', async () => {
      const { useCalendarStore } = await import('../../../store/calendarStore')
      vi.mocked(useCalendarStore).mockImplementation(
        (selector: (state: CalendarState) => unknown) => {
          const state: CalendarState = {
            view: 'month',
            currentDate: new Date(2026, 8, 13),
            setView: vi.fn(),
            setCurrentDate: vi.fn(),
          }
          return selector ? selector(state) : state
        },
      )
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [fullEvent] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)
      const { MonthCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(MonthCalendar)
      act(() => {
        props.onEventClick({
          id: fullEvent.id,
          title: fullEvent.title,
          start: new Date(fullEvent.startAt),
          end: new Date(fullEvent.endAt),
        })
      })

      expect(screen.getByTestId('event-detail-modal')).toBeInTheDocument()
    })

    it('詳細モーダルの編集操作でフォームモーダルをeditモード・該当イベントで開く', async () => {
      const { useCalendarStore } = await import('../../../store/calendarStore')
      vi.mocked(useCalendarStore).mockImplementation(
        (selector: (state: CalendarState) => unknown) => {
          const state: CalendarState = {
            view: 'month',
            currentDate: new Date(2026, 8, 13),
            setView: vi.fn(),
            setCurrentDate: vi.fn(),
          }
          return selector ? selector(state) : state
        },
      )
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [fullEvent] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)
      const { MonthCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(MonthCalendar)
      act(() => {
        props.onEventClick({
          id: fullEvent.id,
          title: fullEvent.title,
          start: new Date(fullEvent.startAt),
          end: new Date(fullEvent.endAt),
        })
      })
      fireEvent.click(
        screen.getByRole('button', { name: `編集-${fullEvent.id}` }),
      )

      const modal = screen.getByTestId('event-form-modal')
      expect(modal).toHaveTextContent('mode:edit')
      expect(modal).toHaveTextContent(`event:${fullEvent.id}`)
    })
  })
})
