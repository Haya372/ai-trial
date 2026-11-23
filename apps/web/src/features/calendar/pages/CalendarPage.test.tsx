import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
import type { CalendarState } from '../../../stores/calendarStore'

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
vi.mock('../../../stores/calendarStore', () => ({
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
  default: vi.fn(({ open, event, onEdit, onShare }) =>
    open && event ? (
      <div data-testid="event-detail-modal">
        <button onClick={() => onEdit(event)}>編集-{event.id}</button>
        <button onClick={() => onShare(event)}>共有-{event.id}</button>
      </div>
    ) : null,
  ),
}))

vi.mock('../../event/components/ShareLinkDialog', () => ({
  default: vi.fn(({ open, event }) =>
    open && event ? (
      <div data-testid="share-link-dialog">event:{event.id}</div>
    ) : null,
  ),
}))

vi.mock('../../event/components/EventFormModal', () => ({
  default: vi.fn(({ open, mode, event, initialStart }) =>
    open ? (
      <div data-testid="event-form-modal">
        mode:{mode} event:{event?.id ?? 'none'} initialStart:
        {initialStart ? (initialStart as Date).toISOString() : 'none'}
      </div>
    ) : null,
  ),
}))

vi.mock('../../event/components/QuickRegistrationPanel', () => ({
  default: vi.fn(({ open, anchor, start, onClose, onEditDetail }) =>
    open ? (
      <div data-testid="quick-registration-panel">
        start:{(start as Date).toISOString()} anchor:{anchor.x},{anchor.y}
        <button onClick={onClose}>閉じる</button>
        <button
          onClick={() =>
            onEditDetail({
              id: 'quick-event',
              title: '仮予定',
              description: null,
              location: null,
              url: null,
              startAt: (start as Date).toISOString(),
              endAt: (start as Date).toISOString(),
            })
          }
        >
          詳細を編集
        </button>
      </div>
    ) : null,
  ),
}))

// TodayTimelinePanel は個別にテスト済みのためスタブ化し、
// CalendarPage への配置と詳細モーダルへの接続だけを検証する
vi.mock('../../timeline/components/TodayTimelinePanel', () => ({
  default: vi.fn(({ onEventClick }) => (
    <div data-testid="today-timeline-panel">
      <button
        onClick={() =>
          onEventClick({
            id: 'timeline-event',
            title: '今日の予定',
            description: null,
            location: null,
            url: null,
            startAt: new Date(2026, 10, 14, 10, 0).toISOString(),
            endAt: new Date(2026, 10, 14, 11, 0).toISOString(),
            isSubscribed: false,
            subscriptionId: null,
            tags: [],
            recurringSeries: null,
          })
        }
      >
        タイムラインの予定
      </button>
    </div>
  )),
}))

// タグ一覧は絞り込み中のタグがまだ存在するかの判定にだけ使う（初期値: 未取得）
vi.mock('../../tag/hooks/useTagsQuery', () => ({
  useTagsQuery: vi.fn(() => ({ data: undefined })),
}))

// TagManagementButton は個別にテスト済みのためスタブ化し、ヘッダーへの配置だけを検証する
vi.mock('../../tag/components/TagManagementButton', () => ({
  default: vi.fn(() => <div data-testid="tag-management-button" />),
}))

// TagFilterButton は個別にテスト済みのためスタブ化し、
// 絞り込み状態の保持と useEventsQuery への受け渡しだけを検証する
vi.mock('../../tag/components/TagFilterButton', () => ({
  default: vi.fn(({ selectedTagId, onChange }) => (
    <div data-testid="tag-filter-button">
      selected:{selectedTagId ?? 'none'}
      <button onClick={() => onChange('tag-1')}>タグ1で絞り込む</button>
      <button onClick={() => onChange(null)}>絞り込みを解除する</button>
    </div>
  )),
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
  isSubscribed: false,
  subscriptionId: null,
  tags: [],
  recurringSeries: null,
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
      expect(screen.getByRole('status')).toHaveTextContent('読み込み中...')
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
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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

  describe('購読予定のカレンダー表示', () => {
    it('MonthCalendarに渡すイベントにisSubscribedを反映する', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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

      const subscribedEvent: EventResponse = {
        ...fullEvent,
        id: 'event-2',
        isSubscribed: true,
        subscriptionId: 'subscription-1',
      }
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [fullEvent, subscribedEvent] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      const { MonthCalendar } = await import('@repo/ui')
      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(MonthCalendar)
      const events = props.events as Array<{
        id: string
        isSubscribed?: boolean
      }>
      expect(events.find((e) => e.id === 'event-1')?.isSubscribed).toBe(false)
      expect(events.find((e) => e.id === 'event-2')?.isSubscribed).toBe(true)
    })
  })

  describe('タグ色のカレンダー表示', () => {
    it('MonthCalendarに渡すイベントのtagColorsがtagsのcolorを付与順に持つ', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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

      const eventWithTags: EventResponse = {
        ...fullEvent,
        id: 'event-with-tags',
        tags: [
          { id: 'tag-1', name: 'レビュー', color: 'red' as const },
          { id: 'tag-2', name: '設計', color: 'blue' as const },
        ],
      }
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [eventWithTags] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      const { MonthCalendar } = await import('@repo/ui')
      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(MonthCalendar)
      const events = props.events as Array<{ id: string; tagColors?: string[] }>
      expect(events.find((e) => e.id === 'event-with-tags')?.tagColors).toEqual(
        ['red', 'blue'],
      )
    })

    it('タグなし予定のtagColorsは空配列', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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
      const events = props.events as Array<{ id: string; tagColors?: string[] }>
      expect(events.find((e) => e.id === 'event-1')?.tagColors).toEqual([])
    })
  })

  describe('タグ管理', () => {
    it('ヘッダーにタグ管理ボタンを表示する', async () => {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)

      render(<CalendarPage />, { wrapper: createWrapper() })

      expect(screen.getByTestId('tag-management-button')).toBeInTheDocument()
    })
  })

  describe('タグによる絞り込み', () => {
    async function mockEmptyEvents() {
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockClear()
      vi.mocked(useEventsQuery).mockReturnValue({
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)
      return useEventsQuery
    }

    it('初期状態では絞り込まずに予定を取得する', async () => {
      const useEventsQuery = await mockEmptyEvents()

      render(<CalendarPage />, { wrapper: createWrapper() })

      expect(screen.getByTestId('tag-filter-button')).toHaveTextContent(
        'selected:none',
      )
      expect(vi.mocked(useEventsQuery).mock.calls.at(-1)?.[2]).toBeNull()
    })

    it('タグを選ぶとそのタグIDで予定を取得し、選択状態を渡す', async () => {
      const useEventsQuery = await mockEmptyEvents()

      render(<CalendarPage />, { wrapper: createWrapper() })
      fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))

      expect(vi.mocked(useEventsQuery).mock.calls.at(-1)?.[2]).toBe('tag-1')
      expect(screen.getByTestId('tag-filter-button')).toHaveTextContent(
        'selected:tag-1',
      )
    })

    describe('絞り込み中のタグがタグ一覧にない（削除された）とき', () => {
      const tag = (id: string) => ({
        id,
        name: id,
        color: 'blue',
        eventCount: 0,
      })

      async function mockTags(tags: ReturnType<typeof tag>[] | undefined) {
        const { useTagsQuery } = await import('../../tag/hooks/useTagsQuery')
        vi.mocked(useTagsQuery).mockReturnValue({ data: tags } as never)
        return useTagsQuery
      }

      afterEach(() => mockTags(undefined))

      it('絞り込み中はタグ一覧を購読する', async () => {
        await mockEmptyEvents()
        const useTagsQuery = await mockTags(undefined)

        render(<CalendarPage />, { wrapper: createWrapper() })
        expect(vi.mocked(useTagsQuery).mock.calls.at(-1)?.[0]).toBe(false)
        fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))

        expect(vi.mocked(useTagsQuery).mock.calls.at(-1)?.[0]).toBe(true)
      })

      it('タグ一覧に含まれなければ、最初の描画から絞り込まずに予定を取得する', async () => {
        const useEventsQuery = await mockEmptyEvents()
        await mockTags([tag('tag-2')])

        render(<CalendarPage />, { wrapper: createWrapper() })
        fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))

        expect(
          vi
            .mocked(useEventsQuery)
            .mock.calls.some(([, , id]) => id === 'tag-1'),
        ).toBe(false)
        expect(screen.getByTestId('tag-filter-button')).toHaveTextContent(
          'selected:none',
        )
      })

      it('タグ一覧が未取得のうちは選択中のタグで絞り込む', async () => {
        const useEventsQuery = await mockEmptyEvents()
        await mockTags(undefined)

        render(<CalendarPage />, { wrapper: createWrapper() })
        fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))

        expect(vi.mocked(useEventsQuery).mock.calls.at(-1)?.[2]).toBe('tag-1')
      })

      it('作成前の古いタグ一覧で外れても、選択を保持し、タグを含む一覧に更新されると絞り込みに戻る', async () => {
        const useEventsQuery = await mockEmptyEvents()
        await mockTags([tag('tag-2')])

        const { rerender } = render(<CalendarPage />, {
          wrapper: createWrapper(),
        })
        fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))
        await mockTags([tag('tag-1'), tag('tag-2')])
        rerender(<CalendarPage />)

        expect(vi.mocked(useEventsQuery).mock.calls.at(-1)?.[2]).toBe('tag-1')
        expect(screen.getByTestId('tag-filter-button')).toHaveTextContent(
          'selected:tag-1',
        )
      })
    })

    it.each([
      [true, 'true'],
      [false, 'false'],
    ])(
      '前のデータを表示中（isPlaceholderData=%s）はカレンダーの aria-busy を %s にし、取得中は予定を操作できなくする',
      async (isPlaceholderData, busy) => {
        const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
        vi.mocked(useEventsQuery).mockReturnValue({
          data: { events: [fullEvent] },
          isPending: false,
          isError: false,
          isSuccess: true,
          isPlaceholderData,
          error: null,
        } as never)
        const { MonthCalendar } = await import('@repo/ui')

        const { container } = render(<CalendarPage />, {
          wrapper: createWrapper(),
        })

        expect(
          screen.getByTestId('month-calendar').parentElement,
        ).toHaveAttribute('aria-busy', busy)
        // The week view scrolls inside FullCalendar, so disabling input on the
        // calendar (inert / pointer-events) would also block scrolling.
        expect(container.querySelector('[inert]')).toBeNull()
        expect(container.querySelector('.pointer-events-none')).toBeNull()

        act(() => {
          getLatestCallProps(MonthCalendar).onEventClick({
            id: fullEvent.id,
            title: fullEvent.title,
            start: new Date(fullEvent.startAt),
            end: new Date(fullEvent.endAt),
          })
        })
        expect(screen.queryByTestId('event-detail-modal') !== null).toBe(
          !isPlaceholderData,
        )
      },
    )

    it('絞り込みを解除すると tagId なしで予定を取得する', async () => {
      const useEventsQuery = await mockEmptyEvents()

      render(<CalendarPage />, { wrapper: createWrapper() })
      fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))
      fireEvent.click(
        screen.getByRole('button', { name: '絞り込みを解除する' }),
      )

      expect(vi.mocked(useEventsQuery).mock.calls.at(-1)?.[2]).toBeNull()
      expect(screen.getByTestId('tag-filter-button')).toHaveTextContent(
        'selected:none',
      )
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

  describe('カレンダーセルのクリックによる新規作成', () => {
    it('MonthCalendarのonDateClickでクリックした日時をinitialStartに渡してcreateモードで開く', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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
        data: { events: [] },
        isPending: false,
        isError: false,
        isSuccess: true,
        error: null,
      } as never)
      const { MonthCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(MonthCalendar)
      const clickedDate = new Date(2026, 8, 15)
      act(() => {
        props.onDateClick(clickedDate)
      })

      const modal = screen.getByTestId('event-form-modal')
      expect(modal).toHaveTextContent('mode:create')
      expect(modal).toHaveTextContent('event:none')
      expect(modal).toHaveTextContent(
        `initialStart:${clickedDate.toISOString()}`,
      )
    })
  })

  describe('週ビューの時間軸クリックによるクイック登録', () => {
    function setupWeekView() {
      return async () => {
        const { useCalendarStore } = await import(
          '../../../stores/calendarStore'
        )
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
      }
    }

    it('WeekCalendarのonTimeSlotClickでクリック日時・座標を渡してクイック登録パネルを開く', async () => {
      await setupWeekView()()
      const { WeekCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(WeekCalendar)
      const clickedDate = new Date(2026, 8, 15, 10, 0)
      act(() => {
        props.onTimeSlotClick(clickedDate, { x: 100, y: 200 })
      })

      const panel = screen.getByTestId('quick-registration-panel')
      expect(panel).toHaveTextContent(`start:${clickedDate.toISOString()}`)
      expect(panel).toHaveTextContent('anchor:100,200')
    })

    it('クイック登録パネルのonCloseでパネルを閉じる', async () => {
      await setupWeekView()()
      const { WeekCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(WeekCalendar)
      act(() => {
        props.onTimeSlotClick(new Date(2026, 8, 15, 10, 0), { x: 0, y: 0 })
      })
      fireEvent.click(screen.getByRole('button', { name: '閉じる' }))

      expect(
        screen.queryByTestId('quick-registration-panel'),
      ).not.toBeInTheDocument()
    })

    it('クイック登録パネルのonEditDetailで登録した予定をeditモードのフォームモーダルに渡す', async () => {
      await setupWeekView()()
      const { WeekCalendar } = await import('@repo/ui')

      render(<CalendarPage />, { wrapper: createWrapper() })

      const props = getLatestCallProps(WeekCalendar)
      act(() => {
        props.onTimeSlotClick(new Date(2026, 8, 15, 10, 0), { x: 0, y: 0 })
      })
      fireEvent.click(screen.getByRole('button', { name: '詳細を編集' }))

      const modal = screen.getByTestId('event-form-modal')
      expect(modal).toHaveTextContent('mode:edit')
      expect(modal).toHaveTextContent('event:quick-event')
      expect(
        screen.queryByTestId('quick-registration-panel'),
      ).not.toBeInTheDocument()
    })
  })

  describe('予定の詳細表示・編集', () => {
    it('カレンダー上のイベントクリックで対応するEventResponseを詳細モーダルに渡す', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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

  describe('共有リンク生成', () => {
    it('詳細モーダルの共有操作で ShareLinkDialog を該当イベントで開く', async () => {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
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
        screen.getByRole('button', { name: `共有-${fullEvent.id}` }),
      )

      const dialog = screen.getByTestId('share-link-dialog')
      expect(dialog).toHaveTextContent(`event:${fullEvent.id}`)
    })
  })

  describe('今日のタイムラインパネル', () => {
    async function setup(
      view: CalendarState['view'],
      query: { isPending: boolean; isError: boolean },
    ) {
      const { useCalendarStore } = await import('../../../stores/calendarStore')
      vi.mocked(useCalendarStore).mockImplementation(
        (selector: (state: CalendarState) => unknown) => {
          const state: CalendarState = {
            view,
            currentDate: new Date(2026, 8, 13),
            setView: vi.fn(),
            setCurrentDate: vi.fn(),
          }
          return selector ? selector(state) : state
        },
      )
      const { useEventsQuery } = await import('../../../hooks/useEventsQuery')
      vi.mocked(useEventsQuery).mockReturnValue({
        data: query.isPending || query.isError ? undefined : { events: [] },
        isPending: query.isPending,
        isError: query.isError,
        isSuccess: !query.isPending && !query.isError,
        error: null,
      } as never)
      render(<CalendarPage />, { wrapper: createWrapper() })
    }

    it.each(['month', 'week'] as const)(
      '%s ビューでパネルを表示する',
      async (view) => {
        await setup(view, { isPending: false, isError: false })
        expect(screen.getByTestId('today-timeline-panel')).toBeInTheDocument()
      },
    )

    it('カレンダーの予定が取得中でもパネルを表示する', async () => {
      await setup('month', { isPending: true, isError: false })
      expect(screen.getByTestId('today-timeline-panel')).toBeInTheDocument()
    })

    it('カレンダーの予定の取得に失敗してもパネルを表示する', async () => {
      await setup('month', { isPending: false, isError: true })
      expect(screen.getByTestId('today-timeline-panel')).toBeInTheDocument()
    })

    it('パネルの予定クリックで、その予定の詳細モーダルを開く', async () => {
      await setup('month', { isPending: false, isError: false })

      fireEvent.click(
        screen.getByRole('button', { name: 'タイムラインの予定' }),
      )

      expect(screen.getByTestId('event-detail-modal')).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: '編集-timeline-event' }),
      ).toBeInTheDocument()
    })
  })
})
