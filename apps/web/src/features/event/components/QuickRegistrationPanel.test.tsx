import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import QuickRegistrationPanel from './QuickRegistrationPanel'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, createEvent: vi.fn() }
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

const start = new Date(2026, 8, 22, 14, 30)
const anchor = { x: 10, y: 20 }

function renderPanel(
  props: Partial<React.ComponentProps<typeof QuickRegistrationPanel>> = {},
) {
  const onClose = vi.fn()
  const onEditDetail = vi.fn()
  render(
    <QuickRegistrationPanel
      open={true}
      anchor={anchor}
      start={start}
      onClose={onClose}
      onEditDetail={onEditDetail}
      {...props}
    />,
    { wrapper: createWrapper() },
  )
  return { onClose, onEditDetail }
}

describe('QuickRegistrationPanel', () => {
  describe('表示', () => {
    it('タイトル入力欄を表示する', () => {
      renderPanel()
      expect(screen.getByLabelText('タイトル')).toBeInTheDocument()
    })

    it('タイトル入力欄に自動フォーカスが当たる', async () => {
      renderPanel()
      await waitFor(() => {
        expect(screen.getByLabelText('タイトル')).toHaveFocus()
      })
    })

    it('openがfalseのとき何も表示しない', () => {
      renderPanel({ open: false })
      expect(screen.queryByLabelText('タイトル')).not.toBeInTheDocument()
    })
  })

  describe('バリデーション', () => {
    it('タイトルが空の場合はエラーを表示し登録しない', async () => {
      const { createEvent } = await import('../../../api/generated')
      renderPanel()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('タイトルを入力してください'),
        ).toBeInTheDocument()
      })
      expect(createEvent).not.toHaveBeenCalled()
    })
  })

  describe('登録', () => {
    it('保存するとクリック日時を開始、1時間後を終了として予定を作成する', async () => {
      const { createEvent } = await import('../../../api/generated')
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: {
          id: 'new-event',
          title: '仮予定',
          description: null,
          location: null,
          url: null,
          startAt: start.toISOString(),
          endAt: new Date(start.getTime() + 60 * 60 * 1000).toISOString(),
        },
        status: 201,
        headers: new Headers(),
      } as never)

      renderPanel()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '仮予定' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(createEvent).toHaveBeenCalledWith({
          title: '仮予定',
          startAt: start.toISOString(),
          endAt: new Date(start.getTime() + 60 * 60 * 1000).toISOString(),
          description: null,
          location: null,
          url: null,
        })
      })
      expect(mockToastSuccess).toHaveBeenCalled()
    })

    it('登録完了後は「詳細を編集」リンクを表示し、パネルは閉じない', async () => {
      const { createEvent } = await import('../../../api/generated')
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: {
          id: 'new-event',
          title: '仮予定',
          description: null,
          location: null,
          url: null,
          startAt: start.toISOString(),
          endAt: new Date(start.getTime() + 60 * 60 * 1000).toISOString(),
        },
        status: 201,
        headers: new Headers(),
      } as never)

      const { onClose } = renderPanel()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '仮予定' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: '詳細を編集' }),
        ).toBeInTheDocument()
      })
      expect(onClose).not.toHaveBeenCalled()
    })

    it('「詳細を編集」をクリックすると登録した予定でonEditDetailを呼ぶ', async () => {
      const { createEvent } = await import('../../../api/generated')
      const createdEvent = {
        id: 'new-event',
        title: '仮予定',
        description: null,
        location: null,
        url: null,
        startAt: start.toISOString(),
        endAt: new Date(start.getTime() + 60 * 60 * 1000).toISOString(),
      }
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: createdEvent,
        status: 201,
        headers: new Headers(),
      } as never)

      const { onEditDetail } = renderPanel()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '仮予定' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      const editLink = await screen.findByRole('button', {
        name: '詳細を編集',
      })
      fireEvent.click(editLink)

      expect(onEditDetail).toHaveBeenCalledWith(createdEvent)
    })

    it('登録に失敗した場合はエラートーストを表示しパネルを閉じない', async () => {
      const { createEvent } = await import('../../../api/generated')
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: { code: 'VALIDATION_ERROR', message: 'Validation error' },
        status: 400,
        headers: new Headers(),
      } as never)

      const { onClose } = renderPanel()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '仮予定' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '入力内容を確認してください',
        )
      })
      expect(onClose).not.toHaveBeenCalled()
      expect(
        screen.queryByRole('button', { name: '詳細を編集' }),
      ).not.toBeInTheDocument()
    })
  })
})
