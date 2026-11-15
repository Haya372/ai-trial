import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useEventsQuery } from '../../../hooks/useEventsQuery'
import { useCalendarStore } from '../../../stores/calendarStore'

// このファイルはCalendarPage.test.tsxと異なりcalendarStoreをモックしない。
// 月/週の切り替えや前後移動で実際にビュー・日付が変わっても、
// 絞り込み状態が維持されることを検証したいため。

vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    MonthCalendar: vi.fn(() => <div data-testid="month-calendar" />),
    WeekCalendar: vi.fn(() => <div data-testid="week-calendar" />),
  }
})

vi.mock('../../../hooks/useEventsQuery', () => ({
  useEventsQuery: vi.fn(),
}))

vi.mock('../../tag/components/TagFilterButton', () => ({
  default: vi.fn(({ onChange }) => (
    <button onClick={() => onChange('tag-1')}>タグ1で絞り込む</button>
  )),
}))

vi.mock('../../tag/components/TagManagementButton', () => ({
  default: vi.fn(() => null),
}))

vi.mock('../../timeline/components/TodayTimelinePanel', () => ({
  default: vi.fn(() => null),
}))

import CalendarPage from './CalendarPage'

function createWrapper() {
  const queryClient = new QueryClient()
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

function latestQueryArgs() {
  const call = vi.mocked(useEventsQuery).mock.calls.at(-1)
  if (!call) throw new Error('useEventsQuery was not called')
  const [startAt, endAt, tagId] = call
  return { startAt: startAt.toISOString(), endAt: endAt.toISOString(), tagId }
}

beforeEach(() => {
  useCalendarStore.setState({
    view: 'month',
    currentDate: new Date(2026, 8, 13),
  })
  vi.mocked(useEventsQuery).mockClear()
  vi.mocked(useEventsQuery).mockReturnValue({
    data: { events: [] },
    isPending: false,
    isError: false,
  } as never)
})

describe('CalendarPage タグ絞り込みの維持', () => {
  it('月/週を切り替えても絞り込み中のタグで予定を取得する', () => {
    render(<CalendarPage />, { wrapper: createWrapper() })
    fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))
    const monthRange = latestQueryArgs()

    fireEvent.click(screen.getByRole('tab', { name: '週' }))

    const weekRange = latestQueryArgs()
    expect(screen.getByTestId('week-calendar')).toBeInTheDocument()
    expect(weekRange.startAt).not.toBe(monthRange.startAt)
    expect(weekRange.tagId).toBe('tag-1')
  })

  it('前後に移動しても絞り込み中のタグで予定を取得する', () => {
    render(<CalendarPage />, { wrapper: createWrapper() })
    fireEvent.click(screen.getByRole('button', { name: 'タグ1で絞り込む' }))
    const before = latestQueryArgs()

    fireEvent.click(screen.getByRole('button', { name: '次へ' }))
    const afterNext = latestQueryArgs()
    expect(afterNext.startAt).not.toBe(before.startAt)
    expect(afterNext.tagId).toBe('tag-1')

    fireEvent.click(screen.getByRole('button', { name: '前へ' }))
    const afterPrev = latestQueryArgs()
    expect(afterPrev.startAt).toBe(before.startAt)
    expect(afterPrev.tagId).toBe('tag-1')
  })
})
