import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
import EventDetailModal from './EventDetailModal'

const mockEvent: EventResponse = {
  id: 'event-1',
  title: 'デザインレビュー',
  description: null,
  startAt: new Date(2026, 8, 13, 10, 0).toISOString(), // 2026-09-13 10:00
  endAt: new Date(2026, 8, 13, 11, 0).toISOString(), // 2026-09-13 11:00
  location: null,
  url: null,
}

describe('EventDetailModal', () => {
  describe('表示制御', () => {
    it('open が false のとき何も表示しない', () => {
      render(
        <EventDetailModal
          open={false}
          event={mockEvent}
          onClose={vi.fn()}
          onEdit={vi.fn()}
        />,
      )
      expect(screen.queryByText('デザインレビュー')).not.toBeInTheDocument()
    })

    it('open が true かつ event が渡されたとき予定タイトルを表示する', () => {
      render(
        <EventDetailModal
          open={true}
          event={mockEvent}
          onClose={vi.fn()}
          onEdit={vi.fn()}
        />,
      )
      expect(screen.getByText('デザインレビュー')).toBeInTheDocument()
    })

    it('open が true かつ event が null のとき何も表示しない', () => {
      render(
        <EventDetailModal
          open={true}
          event={null}
          onClose={vi.fn()}
          onEdit={vi.fn()}
        />,
      )
      // event が null の場合はモーダル内容を表示しない
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  describe('コンテンツ', () => {
    it('開始時刻・終了時刻を表示する', () => {
      render(
        <EventDetailModal
          open={true}
          event={mockEvent}
          onClose={vi.fn()}
          onEdit={vi.fn()}
        />,
      )
      // 時刻が何らかの形で表示されること
      expect(screen.getByText(/10:00/)).toBeInTheDocument()
      expect(screen.getByText(/11:00/)).toBeInTheDocument()
    })
  })

  describe('編集', () => {
    it('編集ボタン押下でonEditをイベント付きで呼ぶ', () => {
      const onEdit = vi.fn()
      render(
        <EventDetailModal
          open={true}
          event={mockEvent}
          onClose={vi.fn()}
          onEdit={onEdit}
        />,
      )
      fireEvent.click(screen.getByRole('button', { name: '編集' }))
      expect(onEdit).toHaveBeenCalledWith(mockEvent)
    })
  })
})
