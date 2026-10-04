import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import SharePage from './SharePage'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return {
    ...actual,
    getShareByToken: vi.fn(),
    getMe: vi.fn(),
    subscribeToShare: vi.fn(),
    deleteSubscription: vi.fn(),
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

const shareData = {
  title: 'チームMTG',
  startAt: new Date(2026, 9, 8, 10, 0, 0).toISOString(),
  endAt: new Date(2026, 9, 8, 11, 0, 0).toISOString(),
  location: '会議室A',
  url: null,
  description: null,
  isOwnEvent: false,
  isSubscribed: false,
}

describe('SharePage', () => {
  describe('ローディング表示', () => {
    it('取得中はローディングが表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockReturnValueOnce(new Promise(() => {}))
      vi.mocked(getMe).mockReturnValueOnce(new Promise(() => {}))

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })
      expect(screen.getByText('読み込み中…')).toBeInTheDocument()
    })
  })

  describe('正常系', () => {
    it('200 のとき予定タイトルが表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: shareData,
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
    })

    it('description が null のときメモは表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, description: null },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(screen.queryByText(/メモ/)).not.toBeInTheDocument()
    })

    it('location が null のとき場所は表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, location: null },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(screen.queryByText(/場所/)).not.toBeInTheDocument()
    })

    it('url が null のとき URL は表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, url: null },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(screen.queryByText(/URL/)).not.toBeInTheDocument()
    })
  })

  describe('エラーモード', () => {
    it('410 Gone のとき期限切れメッセージが表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { code: 'GONE', message: 'gone' },
        status: 410,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(
          screen.getByText('この共有リンクは期限切れです'),
        ).toBeInTheDocument()
      })
    })

    it('410 Gone のとき ShareEventDetail は表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { code: 'GONE', message: 'gone' },
        status: 410,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(
          screen.getByText('この共有リンクは期限切れです'),
        ).toBeInTheDocument()
      })
      expect(screen.queryByText('チームMTG')).not.toBeInTheDocument()
    })

    it('404 のとき「共有リンクが見つかりません」が表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { code: 'NOT_FOUND', message: 'not found' },
        status: 404,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(
          screen.getByText('共有リンクが見つかりません'),
        ).toBeInTheDocument()
      })
    })

    it('500 のとき汎用エラーが表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { code: 'INTERNAL_ERROR', message: 'internal error' },
        status: 500,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('エラーが発生しました')).toBeInTheDocument()
      })
    })
  })

  describe('編集・削除導線の非表示', () => {
    it('「編集」ボタンが存在しない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: shareData,
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('button', { name: /編集/ }),
      ).not.toBeInTheDocument()
    })

    it('「削除」ボタンが存在しない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: shareData,
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('button', { name: /削除/ }),
      ).not.toBeInTheDocument()
    })
  })

  describe('SubscriptionActions の表示条件', () => {
    it('getMe が 401 のとき SubscriptionActions は表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, isOwnEvent: false, isSubscribed: false },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('button', { name: '自分のカレンダーに追加' }),
      ).not.toBeInTheDocument()
    })

    it('getMe が 200 かつ isOwnEvent=true のとき SubscriptionActions は表示されない', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, isOwnEvent: true, isSubscribed: false },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
        status: 200,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('チームMTG')).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('button', { name: '自分のカレンダーに追加' }),
      ).not.toBeInTheDocument()
    })

    it('getMe が 200 かつ isOwnEvent=false かつ isSubscribed=false のとき「追加」ボタンが表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, isOwnEvent: false, isSubscribed: false },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
        status: 200,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: '自分のカレンダーに追加' }),
        ).toBeInTheDocument()
      })
    })

    it('getMe が 200 かつ isOwnEvent=false かつ isSubscribed=true のとき「追加済み」+「削除」が表示される', async () => {
      const { getShareByToken, getMe } = await import('../../../api/generated')
      vi.mocked(getShareByToken).mockResolvedValueOnce({
        data: { ...shareData, isOwnEvent: false, isSubscribed: true },
        status: 200,
        headers: new Headers(),
      } as never)
      vi.mocked(getMe).mockResolvedValueOnce({
        data: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
        status: 200,
        headers: new Headers(),
      } as never)

      render(<SharePage token="tok1" />, { wrapper: createWrapper() })

      await waitFor(() => {
        expect(screen.getByText('追加済み')).toBeInTheDocument()
      })
      expect(
        screen.getByRole('button', { name: 'カレンダーから削除' }),
      ).toBeInTheDocument()
    })
  })
})
