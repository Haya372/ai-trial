import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { CALENDAR_VIEW } from '../constants'
import CalendarNavigation from './CalendarNavigation'

describe('CalendarNavigation', () => {
  const defaultProps = {
    view: CALENDAR_VIEW.MONTH,
    currentDate: new Date(2026, 8, 13), // 2026年9月13日
    onPrev: vi.fn(),
    onNext: vi.fn(),
    onToday: vi.fn(),
  }

  describe('期間ラベル', () => {
    it('月ビューのとき "2026年9月" を表示する', () => {
      render(<CalendarNavigation {...defaultProps} />)
      expect(screen.getByText('2026年9月')).toBeInTheDocument()
    })

    it('週ビューのとき週の範囲を表示する', () => {
      render(
        <CalendarNavigation
          {...defaultProps}
          view={CALENDAR_VIEW.WEEK}
          currentDate={new Date(2026, 8, 13)} // 2026-09-13 (日曜始まりの週: 9/13-9/19)
        />,
      )
      // 週の範囲が表示されること（フォーマットは実装次第）
      expect(screen.getByText(/2026年9月/)).toBeInTheDocument()
    })

    it('週が年をまたぐとき終了日の年も表示する', () => {
      render(
        <CalendarNavigation
          {...defaultProps}
          view={CALENDAR_VIEW.WEEK}
          currentDate={new Date(2026, 11, 30)} // 2026-12-30 (水): 週は12/27〜2027/1/2
        />,
      )
      expect(screen.getByText(/2026年12月27日/)).toBeInTheDocument()
      expect(screen.getByText(/2027年1月2日/)).toBeInTheDocument()
    })
  })

  describe('ナビゲーションボタン', () => {
    it('「前へ」ボタンをクリックすると onPrev が呼ばれる', () => {
      const onPrev = vi.fn()
      render(<CalendarNavigation {...defaultProps} onPrev={onPrev} />)
      fireEvent.click(screen.getByRole('button', { name: /前へ/ }))
      expect(onPrev).toHaveBeenCalledOnce()
    })

    it('「次へ」ボタンをクリックすると onNext が呼ばれる', () => {
      const onNext = vi.fn()
      render(<CalendarNavigation {...defaultProps} onNext={onNext} />)
      fireEvent.click(screen.getByRole('button', { name: /次へ/ }))
      expect(onNext).toHaveBeenCalledOnce()
    })

    it('「今日」ボタンをクリックすると onToday が呼ばれる', () => {
      const onToday = vi.fn()
      render(<CalendarNavigation {...defaultProps} onToday={onToday} />)
      fireEvent.click(screen.getByRole('button', { name: /今日/ }))
      expect(onToday).toHaveBeenCalledOnce()
    })
  })
})
