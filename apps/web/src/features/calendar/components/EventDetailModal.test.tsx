import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
import EventDetailModal from './EventDetailModal'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, deleteEvent: vi.fn() }
})

const { mockToastError, mockToastSuccess } = vi.hoisted(() => ({
  mockToastError: vi.fn(),
  mockToastSuccess: vi.fn(),
}))

vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    toast: Object.assign(vi.fn(), {
      error: mockToastError,
      success: mockToastSuccess,
    }),
  }
})

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

const mockEvent: EventResponse = {
  id: 'event-1',
  title: 'デザインレビュー',
  description: null,
  startAt: new Date(2026, 8, 13, 10, 0).toISOString(), // 2026-09-13 10:00
  endAt: new Date(2026, 8, 13, 11, 0).toISOString(), // 2026-09-13 11:00
  location: null,
  url: null,
  isSubscribed: false,
}

const mockEventWithDetails: EventResponse = {
  ...mockEvent,
  description: '設計方針をレビューする',
  location: '会議室A',
  url: 'https://example.com/agenda',
}

const mockSubscribedEvent: EventResponse = {
  ...mockEvent,
  isSubscribed: true,
}

function renderModal(
  props: Partial<React.ComponentProps<typeof EventDetailModal>> = {},
) {
  const onClose = vi.fn()
  const onEdit = vi.fn()
  const onShare = vi.fn()
  render(
    <EventDetailModal
      open={true}
      event={mockEvent}
      onClose={onClose}
      onEdit={onEdit}
      onShare={onShare}
      {...props}
    />,
    { wrapper: createWrapper() },
  )
  return { onClose, onEdit, onShare }
}

describe('EventDetailModal', () => {
  describe('表示制御', () => {
    it('open が false のとき何も表示しない', () => {
      renderModal({ open: false })
      expect(screen.queryByText('デザインレビュー')).not.toBeInTheDocument()
    })

    it('open が true かつ event が渡されたとき予定タイトルを表示する', () => {
      renderModal()
      expect(screen.getByText('デザインレビュー')).toBeInTheDocument()
    })

    it('open が true かつ event が null のとき何も表示しない', () => {
      renderModal({ event: null })
      // event が null の場合はモーダル内容を表示しない
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  describe('コンテンツ', () => {
    it('開始時刻・終了時刻を表示する', () => {
      renderModal()
      // 時刻が何らかの形で表示されること
      expect(screen.getByText(/10:00/)).toBeInTheDocument()
      expect(screen.getByText(/11:00/)).toBeInTheDocument()
    })

    it('メモ・場所・URLがある場合はそれぞれ表示する', () => {
      renderModal({ event: mockEventWithDetails })
      expect(screen.getByText('設計方針をレビューする')).toBeInTheDocument()
      expect(screen.getByText('会議室A')).toBeInTheDocument()
      expect(
        screen.getByRole('link', { name: 'https://example.com/agenda' }),
      ).toHaveAttribute('href', 'https://example.com/agenda')
    })

    it('メモ・場所・URLがnullの場合はそれぞれ表示しない', () => {
      renderModal()
      expect(screen.queryByText('メモ')).not.toBeInTheDocument()
      expect(screen.queryByText('場所')).not.toBeInTheDocument()
      expect(screen.queryByText('URL')).not.toBeInTheDocument()
    })

    it('開始・終了ラベルをコロン区切りで表示する', () => {
      renderModal()
      expect(screen.getByText('開始:', { exact: false })).toBeInTheDocument()
      expect(screen.getByText('終了:', { exact: false })).toBeInTheDocument()
    })
  })

  describe('編集', () => {
    it('編集ボタン押下でonEditをイベント付きで呼ぶ', () => {
      const { onEdit } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: '編集' }))
      expect(onEdit).toHaveBeenCalledWith(mockEvent)
    })
  })

  describe('共有', () => {
    it('「共有リンクを生成」ボタンが表示される', () => {
      renderModal()
      expect(
        screen.getByRole('button', { name: '共有リンクを生成' }),
      ).toBeInTheDocument()
    })

    it('「共有リンクを生成」ボタン押下でonShareをイベント付きで呼ぶ', () => {
      const { onShare } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: '共有リンクを生成' }))
      expect(onShare).toHaveBeenCalledWith(mockEvent)
    })
  })

  describe('購読中の予定（isSubscribed）', () => {
    it('isSubscribedがtrueのとき「閲覧専用」を表示する', () => {
      renderModal({ event: mockSubscribedEvent })
      expect(screen.getByText('閲覧専用')).toBeInTheDocument()
    })

    it('isSubscribedがfalseのとき「閲覧専用」を表示しない', () => {
      renderModal()
      expect(screen.queryByText('閲覧専用')).not.toBeInTheDocument()
    })

    it('isSubscribedがtrueのとき削除・編集・共有ボタンを表示しない', () => {
      renderModal({ event: mockSubscribedEvent })
      expect(
        screen.queryByRole('button', { name: '削除' }),
      ).not.toBeInTheDocument()
      expect(
        screen.queryByRole('button', { name: '編集' }),
      ).not.toBeInTheDocument()
      expect(
        screen.queryByRole('button', { name: '共有リンクを生成' }),
      ).not.toBeInTheDocument()
    })
  })

  describe('削除', () => {
    it('削除ボタン押下で確認ダイアログを表示する', () => {
      renderModal()
      fireEvent.click(screen.getByRole('button', { name: '削除' }))
      expect(screen.getByText('この予定を削除しますか？')).toBeInTheDocument()
    })

    it('確認ダイアログで「キャンセル」を押すと確認ダイアログを閉じ、詳細モーダルは開いたままになる', () => {
      const { onClose } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: '削除' }))
      fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }))
      expect(
        screen.queryByText('この予定を削除しますか？'),
      ).not.toBeInTheDocument()
      expect(screen.getByText('デザインレビュー')).toBeInTheDocument()
      expect(onClose).not.toHaveBeenCalled()
    })

    it('確認ダイアログで「削除」を押すとdeleteEventを呼び、成功後にonCloseを呼ぶ', async () => {
      const { deleteEvent } = await import('../../../api/generated')
      vi.mocked(deleteEvent).mockResolvedValueOnce({
        data: undefined,
        status: 204,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: '削除' }))
      const confirmDialog = screen.getByRole('alertdialog')
      fireEvent.click(
        within(confirmDialog).getByRole('button', { name: '削除' }),
      )

      await waitFor(() => {
        expect(deleteEvent).toHaveBeenCalledWith('event-1')
      })
      await waitFor(() => expect(onClose).toHaveBeenCalled())
      expect(mockToastSuccess).toHaveBeenCalled()
    })

    it('削除中は確認ダイアログのキャンセルボタンが無効化される', async () => {
      const { deleteEvent } = await import('../../../api/generated')
      let resolveDelete: (
        value: Awaited<ReturnType<typeof deleteEvent>>,
      ) => void = () => {}
      vi.mocked(deleteEvent).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveDelete = resolve
          }),
      )

      renderModal()
      fireEvent.click(screen.getByRole('button', { name: '削除' }))
      fireEvent.click(
        within(screen.getByRole('alertdialog')).getByRole('button', {
          name: '削除',
        }),
      )

      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: 'キャンセル' }),
        ).toBeDisabled()
      })

      resolveDelete({
        data: undefined,
        status: 204,
        headers: new Headers(),
      } as never)
    })

    it('削除失敗時はエラートーストを表示し、モーダルを閉じない', async () => {
      const { deleteEvent } = await import('../../../api/generated')
      vi.mocked(deleteEvent).mockResolvedValueOnce({
        data: { code: 'INTERNAL_ERROR', message: 'error' },
        status: 500,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: '削除' }))
      fireEvent.click(
        within(screen.getByRole('alertdialog')).getByRole('button', {
          name: '削除',
        }),
      )

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          'サーバーエラーが発生しました。しばらく経ってから再試行してください',
        )
      })
      expect(onClose).not.toHaveBeenCalled()
      expect(screen.getByText('デザインレビュー')).toBeInTheDocument()
    })
  })
})
