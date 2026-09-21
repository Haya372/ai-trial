import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import CalendarNav from './CalendarNav'

describe('CalendarNav', () => {
  const defaultProps = {
    view: 'month' as const,
    currentDate: new Date(2026, 8, 13), // 2026年9月13日
    onPrev: vi.fn(),
    onNext: vi.fn(),
    onToday: vi.fn(),
  }

  describe('期間ラベル', () => {
    it('月ビューのとき "2026年9月" を表示する', () => {
      render(<CalendarNav {...defaultProps} />)
      expect(screen.getByText('2026年9月')).toBeInTheDocument()
    })

    it('週ビューのとき週の範囲を表示する', () => {
      render(
        <CalendarNav
          {...defaultProps}
          view="week"
          currentDate={new Date(2026, 8, 13)} // 2026-09-13 (日曜始まりの週: 9/13-9/19)
        />,
      )
      // 週の範囲が表示されること（フォーマットは実装次第）
      expect(screen.getByText(/2026年9月/)).toBeInTheDocument()
    })
  })

  describe('ナビゲーションボタン', () => {
    it('「前へ」ボタンをクリックすると onPrev が呼ばれる', () => {
      const onPrev = vi.fn()
      render(<CalendarNav {...defaultProps} onPrev={onPrev} />)
      fireEvent.click(screen.getByRole('button', { name: /前へ/ }))
      expect(onPrev).toHaveBeenCalledOnce()
    })

    it('「次へ」ボタンをクリックすると onNext が呼ばれる', () => {
      const onNext = vi.fn()
      render(<CalendarNav {...defaultProps} onNext={onNext} />)
      fireEvent.click(screen.getByRole('button', { name: /次へ/ }))
      expect(onNext).toHaveBeenCalledOnce()
    })

    it('「今日」ボタンをクリックすると onToday が呼ばれる', () => {
      const onToday = vi.fn()
      render(<CalendarNav {...defaultProps} onToday={onToday} />)
      fireEvent.click(screen.getByRole('button', { name: /今日/ }))
      expect(onToday).toHaveBeenCalledOnce()
    })
  })
})
