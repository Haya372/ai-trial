import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import i18n from 'i18next'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { EventResponse } from '../../../api/generated'
import ShareLinkDialog from './ShareLinkDialog'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, createEventShare: vi.fn() }
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

// 未来の日時を使うため、テスト実行時点より確実に先の値を使う
const futureEndAt = new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString()
const futureStartAt = new Date(
  Date.now() + 1 * 24 * 60 * 60 * 1000,
).toISOString()

const mockEvent: EventResponse = {
  id: 'event-1',
  title: 'テスト予定',
  description: null,
  startAt: futureStartAt,
  endAt: futureEndAt,
  location: null,
  url: null,
}

function renderDialog(
  props: Partial<React.ComponentProps<typeof ShareLinkDialog>> = {},
) {
  const onClose = vi.fn()
  render(
    <ShareLinkDialog
      open={true}
      event={mockEvent}
      onClose={onClose}
      {...props}
    />,
    { wrapper: createWrapper() },
  )
  return { onClose }
}

describe('ShareLinkDialog', () => {
  describe('表示制御', () => {
    it('open が false のとき何も表示しない', () => {
      renderDialog({ open: false })
      expect(screen.queryByText('共有リンクを生成')).not.toBeInTheDocument()
    })

    it('open が true のときタイトルが表示される', () => {
      renderDialog()
      expect(screen.getByText('共有リンクを生成')).toBeInTheDocument()
    })

    it('open が true のとき注意文言が表示される', () => {
      renderDialog()
      expect(
        screen.getByText(/メモ欄の内容もリンクで共有されます/),
      ).toBeInTheDocument()
    })

    it('open が true のとき有効期限フィールドが表示される', () => {
      renderDialog()
      expect(screen.getByLabelText('有効期限')).toBeInTheDocument()
    })

    it('open が true のとき「生成」ボタンが表示される', () => {
      renderDialog()
      expect(screen.getByRole('button', { name: '生成' })).toBeInTheDocument()
    })

    it('有効期限の初期値が event.endAt を datetime-local 形式に変換した値になっている', () => {
      renderDialog()
      const input = screen.getByLabelText('有効期限') as HTMLInputElement
      // endAt の datetime-local 形式の値が設定されている（秒・ミリ秒なし）
      expect(input.value).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/)
    })
  })

  describe('バリデーション', () => {
    it('有効期限が空のとき「有効期限を入力してください」が表示され createEventShare は呼ばれない', async () => {
      const { createEventShare } = await import('../../../api/generated')
      renderDialog()
      fireEvent.change(screen.getByLabelText('有効期限'), {
        target: { value: '' },
      })
      fireEvent.click(screen.getByRole('button', { name: '生成' }))
      await waitFor(() => {
        expect(
          screen.getByText('有効期限を入力してください'),
        ).toBeInTheDocument()
      })
      expect(createEventShare).not.toHaveBeenCalled()
    })

    it('有効期限が現在時刻以前のとき「有効期限は未来の日時を指定してください」が表示される', async () => {
      renderDialog()
      // 過去の日時を入力
      fireEvent.change(screen.getByLabelText('有効期限'), {
        target: { value: '2020-01-01T00:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '生成' }))
      await waitFor(() => {
        expect(
          screen.getByText('有効期限は未来の日時を指定してください'),
        ).toBeInTheDocument()
      })
    })

    it('有効期限が予定の startAt より前のとき「有効期限は予定の開始日時以降を指定してください」が表示される', async () => {
      // startAt が1日後・endAt が2日後のイベントで、startAt より前（かつ未来）の値を入力
      const start = new Date(Date.now() + 2 * 24 * 60 * 60 * 1000)
      const end = new Date(Date.now() + 3 * 24 * 60 * 60 * 1000)
      const event: EventResponse = {
        ...mockEvent,
        startAt: start.toISOString(),
        endAt: end.toISOString(),
      }
      renderDialog({ event })
      // startAt（2日後）より前で、現在より後（1日後）の値
      const beforeStart = new Date(Date.now() + 1 * 24 * 60 * 60 * 1000)
      const y = beforeStart.getFullYear()
      const mo = String(beforeStart.getMonth() + 1).padStart(2, '0')
      const d = String(beforeStart.getDate()).padStart(2, '0')
      const h = String(beforeStart.getHours()).padStart(2, '0')
      const mi = String(beforeStart.getMinutes()).padStart(2, '0')
      fireEvent.change(screen.getByLabelText('有効期限'), {
        target: { value: `${y}-${mo}-${d}T${h}:${mi}` },
      })
      fireEvent.click(screen.getByRole('button', { name: '生成' }))
      await waitFor(() => {
        expect(
          screen.getByText('有効期限は予定の開始日時以降を指定してください'),
        ).toBeInTheDocument()
      })
    })

    it('バリデーションエラー表示中に言語を切り替えるとメッセージが追従する', async () => {
      renderDialog()
      fireEvent.change(screen.getByLabelText('有効期限'), {
        target: { value: '' },
      })
      fireEvent.click(screen.getByRole('button', { name: '生成' }))
      await waitFor(() => {
        expect(
          screen.getByText('有効期限を入力してください'),
        ).toBeInTheDocument()
      })
      // en に切り替えても同じキーに紐づいたメッセージが表示される（en は暫定日本語）
      await i18n.changeLanguage('en')
      await waitFor(() => {
        expect(
          screen.getByText('有効期限を入力してください'),
        ).toBeInTheDocument()
      })
    })
  })

  describe('送信', () => {
    it('正常系: 201 を返すとき URL とコピーボタンが表示される', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: {
          url: '/share/abc123',
          expiresAt: futureEndAt,
        },
        status: 201,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastSuccess).toHaveBeenCalledWith(
          '共有リンクを生成しました',
        )
      })
      // window.location.origin + '/share/abc123' の完全 URL が表示される
      expect(screen.getByDisplayValue(/\/share\/abc123/)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'コピー' })).toBeInTheDocument()
    })

    it('400 VALIDATION_ERROR + BEFORE_EVENT_START: 有効期限フィールドのエラーが表示される', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: {
          code: 'VALIDATION_ERROR',
          message: 'validation failed',
          details: [
            {
              field: 'expiresAt',
              code: 'BEFORE_EVENT_START',
              message: 'expiresAt must not be before the event start',
            },
          ],
        },
        status: 400,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(
          screen.getByText('有効期限は予定の開始日時以降を指定してください'),
        ).toBeInTheDocument()
      })
    })

    it('400 VALIDATION_ERROR + NOT_IN_FUTURE: 有効期限フィールドのエラーが表示される', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: {
          code: 'VALIDATION_ERROR',
          message: 'validation failed',
          details: [
            {
              field: 'expiresAt',
              code: 'NOT_IN_FUTURE',
              message: 'expiresAt must be in the future',
            },
          ],
        },
        status: 400,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(
          screen.getByText('有効期限は未来の日時を指定してください'),
        ).toBeInTheDocument()
      })
    })

    it('400 VALIDATION_ERROR: expiresAt の detail が配列の先頭以外にあってもフィールドエラーが表示される', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: {
          code: 'VALIDATION_ERROR',
          message: 'validation failed',
          details: [
            {
              field: 'other',
              code: 'SOME_OTHER_ERROR',
              message: 'unrelated field error',
            },
            {
              field: 'expiresAt',
              code: 'NOT_IN_FUTURE',
              message: 'expiresAt must be in the future',
            },
          ],
        },
        status: 400,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(
          screen.getByText('有効期限は未来の日時を指定してください'),
        ).toBeInTheDocument()
      })
    })

    it('401: モーダル内 Alert にエラーメッセージが表示され toast.error が呼ばれる', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'not authenticated' },
        status: 401,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    it('403: モーダル内 Alert にエラーメッセージが表示され toast.error が呼ばれる', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: { code: 'FORBIDDEN', message: 'forbidden' },
        status: 403,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          'この予定を共有する権限がありません',
        )
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    it('404: モーダル内 Alert にエラーメッセージが表示され toast.error が呼ばれる', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: { code: 'NOT_FOUND', message: 'not found' },
        status: 404,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '予定が見つかりませんでした（削除された可能性があります）',
        )
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    it('500: モーダル内 Alert にエラーメッセージが表示され toast.error が呼ばれる', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: { code: 'INTERNAL_ERROR', message: 'internal error' },
        status: 500,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    it('ネットワーク例外時: フォールバックのエラーメッセージが表示される', async () => {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockRejectedValueOnce(
        new Error('network error'),
      )

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '共有リンクの生成に失敗しました',
        )
      })
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  describe('コピー', () => {
    async function renderWithResult() {
      const { createEventShare } = await import('../../../api/generated')
      vi.mocked(createEventShare).mockResolvedValueOnce({
        data: {
          url: '/share/tok123',
          expiresAt: futureEndAt,
        },
        status: 201,
        headers: new Headers(),
      } as never)

      renderDialog()
      fireEvent.click(screen.getByRole('button', { name: '生成' }))
      await waitFor(() => {
        expect(
          screen.getByRole('button', { name: 'コピー' }),
        ).toBeInTheDocument()
      })
    }

    it('「コピー」ボタン押下で navigator.clipboard.writeText が完全 URL で呼ばれる', async () => {
      const writeText = vi.fn().mockResolvedValueOnce(undefined)
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText },
        configurable: true,
      })

      await renderWithResult()
      fireEvent.click(screen.getByRole('button', { name: 'コピー' }))

      await waitFor(() => {
        expect(writeText).toHaveBeenCalledWith(
          expect.stringContaining('/share/tok123'),
        )
      })
    })

    it('clipboard 成功時: toast.success が呼ばれる', async () => {
      const writeText = vi.fn().mockResolvedValueOnce(undefined)
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText },
        configurable: true,
      })

      await renderWithResult()
      fireEvent.click(screen.getByRole('button', { name: 'コピー' }))

      await waitFor(() => {
        expect(mockToastSuccess).toHaveBeenCalledWith(
          '共有リンクをコピーしました',
        )
      })
    })

    it('clipboard 失敗時: toast.error が呼ばれる', async () => {
      const writeText = vi.fn().mockRejectedValueOnce(new Error('denied'))
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText },
        configurable: true,
      })

      await renderWithResult()
      fireEvent.click(screen.getByRole('button', { name: 'コピー' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '共有リンクのコピーに失敗しました',
        )
      })
    })
  })

  describe('競合状態（stale response）', () => {
    it('別の予定に切り替えた後に前の予定への生成リクエストが成功しても、結果が表示されない', async () => {
      const { createEventShare } = await import('../../../api/generated')
      let resolveRequest!: (value: {
        data: { url: string; expiresAt: string }
        status: 201
        headers: Headers
      }) => void
      vi.mocked(createEventShare).mockReturnValueOnce(
        new Promise((resolve) => {
          resolveRequest = resolve as never
        }) as never,
      )

      const onClose = vi.fn()
      const { rerender } = render(
        <ShareLinkDialog open={true} event={mockEvent} onClose={onClose} />,
        { wrapper: createWrapper() },
      )
      fireEvent.click(screen.getByRole('button', { name: '生成' }))

      // 予定Aへのリクエストが送信される（バリデーションの非同期処理を flush する）まで待つ
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 0))
      })

      // 予定Aへのリクエストが未解決のまま、ダイアログを閉じて別の予定Bで開き直す
      const eventB: EventResponse = { ...mockEvent, id: 'event-2' }
      rerender(
        <ShareLinkDialog open={false} event={mockEvent} onClose={onClose} />,
      )
      rerender(<ShareLinkDialog open={true} event={eventB} onClose={onClose} />)

      // 予定Aへのリクエストがここで解決する。関連する非同期処理（状態更新）が
      // 完全に片付くまで act 内でマイクロタスクを明示的に流し切る
      await act(async () => {
        resolveRequest({
          data: { url: '/share/stale-a', expiresAt: futureEndAt },
          status: 201,
          headers: new Headers(),
        })
        await new Promise((resolve) => setTimeout(resolve, 0))
      })

      // 予定Aの結果が予定Bのダイアログに紛れ込んで表示されないこと
      expect(screen.getByRole('button', { name: '生成' })).toBeInTheDocument()
      expect(screen.queryByDisplayValue(/stale-a/)).not.toBeInTheDocument()
      expect(mockToastSuccess).not.toHaveBeenCalled()
    })
  })
})
