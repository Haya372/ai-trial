import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
import EventFormModal from './EventFormModal'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, createEvent: vi.fn(), updateEvent: vi.fn() }
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

const existingEvent: EventResponse = {
  id: 'event-1',
  title: 'デザインレビュー',
  description: 'メモ',
  startAt: '2026-09-22T01:00:00.000Z',
  endAt: '2026-09-22T02:00:00.000Z',
  location: '会議室A',
  url: 'https://example.com',
}

function renderModal(
  props: Partial<React.ComponentProps<typeof EventFormModal>> = {},
) {
  const onClose = vi.fn()
  render(
    <EventFormModal
      open={true}
      mode="create"
      event={null}
      onClose={onClose}
      {...props}
    />,
    { wrapper: createWrapper() },
  )
  return { onClose }
}

describe('EventFormModal', () => {
  describe('表示', () => {
    it('createモードのとき「予定を作成」を表示する', () => {
      renderModal({ mode: 'create', event: null })
      expect(screen.getByText('予定を作成')).toBeInTheDocument()
    })

    it('editモードのとき「予定を編集」と既存の値を表示する', () => {
      renderModal({ mode: 'edit', event: existingEvent })
      expect(screen.getByText('予定を編集')).toBeInTheDocument()
      expect(screen.getByDisplayValue('デザインレビュー')).toBeInTheDocument()
      expect(screen.getByDisplayValue('メモ')).toBeInTheDocument()
      expect(screen.getByDisplayValue('会議室A')).toBeInTheDocument()
      expect(
        screen.getByDisplayValue('https://example.com'),
      ).toBeInTheDocument()
    })

    it('openがfalseのとき何も表示しない', () => {
      renderModal({ open: false })
      expect(screen.queryByText('予定を作成')).not.toBeInTheDocument()
    })
  })

  describe('バリデーション', () => {
    it('タイトルが空の場合はエラーを表示し送信しない', async () => {
      const { createEvent } = await import('../../../api/generated')
      renderModal()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('タイトルを入力してください'),
        ).toBeInTheDocument()
      })
      expect(createEvent).not.toHaveBeenCalled()
    })

    it('終了日時が開始日時以前の場合はエラーを表示する', async () => {
      renderModal()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: 'テスト予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T09:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('終了日時は開始日時より後に設定してください'),
        ).toBeInTheDocument()
      })
    })
  })

  describe('送信', () => {
    it('createモードで保存するとcreateEventを呼び、成功後にonCloseを呼ぶ', async () => {
      const { createEvent } = await import('../../../api/generated')
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: { ...existingEvent, id: 'new-event' },
        status: 201,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({ mode: 'create', event: null })
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '新しい予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T11:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(createEvent).toHaveBeenCalledWith(
          expect.objectContaining({ title: '新しい予定' }),
        )
      })
      await waitFor(() => expect(onClose).toHaveBeenCalled())
      expect(mockToastSuccess).toHaveBeenCalled()
    })

    it('editモードで保存するとupdateEventをidつきで呼ぶ', async () => {
      const { updateEvent } = await import('../../../api/generated')
      vi.mocked(updateEvent).mockResolvedValueOnce({
        data: existingEvent,
        status: 200,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({
        mode: 'edit',
        event: existingEvent,
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(updateEvent).toHaveBeenCalledWith(
          'event-1',
          expect.objectContaining({ title: 'デザインレビュー' }),
        )
      })
      await waitFor(() => expect(onClose).toHaveBeenCalled())
    })

    it('保存失敗時はエラートーストを表示しモーダルを閉じない', async () => {
      const { createEvent } = await import('../../../api/generated')
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: { code: 'VALIDATION_ERROR', message: 'Validation error' },
        status: 400,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({ mode: 'create', event: null })
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '新しい予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T11:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '入力内容を確認してください',
        )
      })
      expect(onClose).not.toHaveBeenCalled()
    })
  })

  describe('キャンセル', () => {
    it('キャンセルボタンでonCloseを呼ぶ', () => {
      const { onClose } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }))
      expect(onClose).toHaveBeenCalled()
    })
  })
})
