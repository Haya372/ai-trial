import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import EventDetailModal from './EventDetailModal'

const mockEvent = {
  id: 'event-1',
  title: 'デザインレビュー',
  start: new Date(2026, 8, 13, 10, 0), // 2026-09-13 10:00
  end: new Date(2026, 8, 13, 11, 0), // 2026-09-13 11:00
}

describe('EventDetailModal', () => {
  describe('表示制御', () => {
    it('open が false のとき何も表示しない', () => {
      render(
        <EventDetailModal open={false} event={mockEvent} onClose={vi.fn()} />,
      )
      expect(screen.queryByText('デザインレビュー')).not.toBeInTheDocument()
    })

    it('open が true かつ event が渡されたとき予定タイトルを表示する', () => {
      render(
        <EventDetailModal open={true} event={mockEvent} onClose={vi.fn()} />,
      )
      expect(screen.getByText('デザインレビュー')).toBeInTheDocument()
    })

    it('open が true かつ event が null のとき何も表示しない', () => {
      render(<EventDetailModal open={true} event={null} onClose={vi.fn()} />)
      // event が null の場合はモーダル内容を表示しない
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  describe('コンテンツ', () => {
    it('開始時刻・終了時刻を表示する', () => {
      render(
        <EventDetailModal open={true} event={mockEvent} onClose={vi.fn()} />,
      )
      // 時刻が何らかの形で表示されること
      expect(screen.getByText(/10:00/)).toBeInTheDocument()
      expect(screen.getByText(/11:00/)).toBeInTheDocument()
    })
  })
})
