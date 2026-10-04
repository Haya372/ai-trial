import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import SubscriptionActions from './SubscriptionActions'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return {
    ...actual,
    subscribeToShare: vi.fn(),
    deleteSubscription: vi.fn(),
  }
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

describe('SubscriptionActions', () => {
  describe('初期表示', () => {
    it('initialIsSubscribed=false のとき「自分のカレンダーに追加」ボタンが表示される', () => {
      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      expect(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      ).toBeInTheDocument()
    })

    it('initialIsSubscribed=true のとき「追加済み」ラベルと「カレンダーから削除」ボタンが表示される', () => {
      render(<SubscriptionActions token="tok1" initialIsSubscribed={true} />, {
        wrapper: createWrapper(),
      })
      expect(screen.getByText('追加済み')).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      ).toBeInTheDocument()
    })
  })

  describe('追加フロー', () => {
    it('「追加」クリックで subscribeToShare が呼ばれる', async () => {
      const { subscribeToShare } = await import('../../../api/generated')
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-1',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 201,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )

      await waitFor(() => {
        expect(subscribeToShare).toHaveBeenCalledWith('tok1')
      })
    })

    it('201 で成功したとき表示が「追加済み」+「削除」ボタンに切り替わる', async () => {
      const { subscribeToShare } = await import('../../../api/generated')
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-1',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 201,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )

      await waitFor(() => {
        expect(screen.getByText('追加済み')).toBeInTheDocument()
      })
      expect(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      ).toBeInTheDocument()
    })

    it('200 で成功したときも表示が「追加済み」に切り替わる', async () => {
      const { subscribeToShare } = await import('../../../api/generated')
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-1',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 200,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )

      await waitFor(() => {
        expect(screen.getByText('追加済み')).toBeInTheDocument()
      })
    })

    it('403 FORBIDDEN のとき Alert が表示され toast.error が呼ばれ、「追加」ボタン表示に戻る', async () => {
      const { subscribeToShare } = await import('../../../api/generated')
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: { code: 'FORBIDDEN', message: 'forbidden' },
        status: 403,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      ).toBeInTheDocument()
    })

    it('ネットワーク例外時にフォールバックメッセージが表示される', async () => {
      const { subscribeToShare } = await import('../../../api/generated')
      vi.mocked(subscribeToShare).mockRejectedValueOnce(
        new Error('network error'),
      )

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  describe('削除フロー（id を保持済みの場合）', () => {
    it('追加成功後に「削除」クリックで deleteSubscription だけが呼ばれる', async () => {
      const { subscribeToShare, deleteSubscription } = await import(
        '../../../api/generated'
      )
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-1',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 201,
        headers: new Headers(),
      } as never)
      vi.mocked(deleteSubscription).mockResolvedValueOnce({
        data: undefined,
        status: 204,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })

      // 追加してから削除
      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )
      await waitFor(() =>
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )
      fireEvent.click(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )

      await waitFor(() => {
        expect(deleteSubscription).toHaveBeenCalledWith('sub-1')
      })
      // POST は1回（追加時のみ）
      expect(subscribeToShare).toHaveBeenCalledTimes(1)
    })

    it('204 で削除成功後に「追加」ボタン表示に戻る', async () => {
      const { subscribeToShare, deleteSubscription } = await import(
        '../../../api/generated'
      )
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-1',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 201,
        headers: new Headers(),
      } as never)
      vi.mocked(deleteSubscription).mockResolvedValueOnce({
        data: undefined,
        status: 204,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={false} />, {
        wrapper: createWrapper(),
      })

      fireEvent.click(
        screen.getByRole('button', { name: '自分のカレンダーに追加' }),
      )
      await waitFor(() =>
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )
      fireEvent.click(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )

      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: '自分のカレンダーに追加' }),
        ).toBeInTheDocument()
      })
    })
  })

  describe('削除フロー（遅延 id 解決）', () => {
    it('initialIsSubscribed=true で「削除」クリックすると POST してから DELETE する', async () => {
      const { subscribeToShare, deleteSubscription } = await import(
        '../../../api/generated'
      )
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: {
          id: 'sub-2',
          eventId: 'ev-1',
          createdAt: '2026-01-01T00:00:00Z',
        },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(deleteSubscription).mockResolvedValueOnce({
        data: undefined,
        status: 204,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={true} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )

      await waitFor(() => {
        expect(subscribeToShare).toHaveBeenCalledWith('tok1')
        expect(deleteSubscription).toHaveBeenCalledWith('sub-2')
      })
    })

    it('遅延 POST が 410 で失敗したとき DELETE は呼ばれず「追加済み」表示が維持される', async () => {
      const { subscribeToShare, deleteSubscription } = await import(
        '../../../api/generated'
      )
      vi.mocked(subscribeToShare).mockResolvedValueOnce({
        data: { code: 'GONE', message: 'gone' },
        status: 410,
        headers: new Headers(),
      } as never)

      render(<SubscriptionActions token="tok1" initialIsSubscribed={true} />, {
        wrapper: createWrapper(),
      })
      fireEvent.click(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      )

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(deleteSubscription).not.toHaveBeenCalled()
      expect(screen.getByText('追加済み')).toBeInTheDocument()
    })
  })
})
