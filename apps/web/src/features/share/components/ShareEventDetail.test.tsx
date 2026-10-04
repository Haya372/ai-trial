import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { ShareResponse } from '../../../api/generated'
import ShareEventDetail from './ShareEventDetail'

const baseEvent: ShareResponse = {
  title: 'チームMTG',
  startAt: new Date(2026, 9, 8, 10, 0, 0).toISOString(),
  endAt: new Date(2026, 9, 8, 11, 0, 0).toISOString(),
  location: '会議室A',
  url: 'https://meet.example.com',
  description: '資料は事前にSlackに投稿',
  isOwnEvent: false,
  isSubscribed: false,
}

describe('ShareEventDetail', () => {
  describe('必須フィールドの表示', () => {
    it('予定タイトルが表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(screen.getByText('チームMTG')).toBeInTheDocument()
    })

    it('開始日時ラベルが表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(screen.getByText(/開始/)).toBeInTheDocument()
    })

    it('終了日時ラベルが表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(screen.getByText(/終了/)).toBeInTheDocument()
    })
  })

  describe('オプションフィールドの表示', () => {
    it('location が非 null のとき場所が表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(screen.getByText('会議室A')).toBeInTheDocument()
    })

    it('location が null のとき場所は表示されない', () => {
      render(<ShareEventDetail event={{ ...baseEvent, location: null }} />)
      expect(screen.queryByText(/場所/)).not.toBeInTheDocument()
    })

    it('url が非 null のとき URL リンクが表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      const link = screen.getByRole('link', {
        name: 'https://meet.example.com',
      })
      expect(link).toBeInTheDocument()
      expect(link).toHaveAttribute('href', 'https://meet.example.com')
      expect(link).toHaveAttribute('target', '_blank')
      expect(link).toHaveAttribute('rel', 'noopener noreferrer')
    })

    it('url が null のとき URL は表示されない', () => {
      render(<ShareEventDetail event={{ ...baseEvent, url: null }} />)
      expect(screen.queryByText(/URL/)).not.toBeInTheDocument()
    })

    it('description が非 null のときメモが表示される', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(screen.getByText('資料は事前にSlackに投稿')).toBeInTheDocument()
    })

    it('description が null のときメモは表示されない', () => {
      render(<ShareEventDetail event={{ ...baseEvent, description: null }} />)
      expect(screen.queryByText(/メモ/)).not.toBeInTheDocument()
    })
  })

  describe('編集・削除導線の非表示', () => {
    it('「編集」ボタンが存在しない', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(
        screen.queryByRole('button', { name: /編集/ }),
      ).not.toBeInTheDocument()
    })

    it('「削除」ボタンが存在しない', () => {
      render(<ShareEventDetail event={baseEvent} />)
      expect(
        screen.queryByRole('button', { name: /削除/ }),
      ).not.toBeInTheDocument()
    })
  })
})
