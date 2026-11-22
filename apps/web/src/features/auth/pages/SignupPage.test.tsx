import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactElement, ReactNode } from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import i18n from 'i18next'
import { describe, expect, it, vi } from 'vitest'
import SignupPage from './SignupPage'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, signup: vi.fn() }
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

function renderPage(ui: ReactElement) {
  const queryClient = new QueryClient()
  return {
    queryClient,
    ...render(ui, {
      wrapper: ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      ),
    }),
  }
}

describe('SignupPage', () => {
  it('renders email, password, and displayName fields with Japanese labels', () => {
    renderPage(<SignupPage />)
    expect(screen.getByLabelText(/メールアドレス/)).toBeInTheDocument()
    expect(screen.getByLabelText(/パスワード/)).toBeInTheDocument()
    expect(screen.getByLabelText(/表示名/)).toBeInTheDocument()
  })

  it('shows validation error for too-short password', async () => {
    renderPage(<SignupPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'ab' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(screen.getByText(/8文字以上/)).toBeInTheDocument()
    })
  })

  it('updates a displayed validation error to the new language when the language changes', async () => {
    renderPage(<SignupPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'ab' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(screen.getByText(/8文字以上/)).toBeInTheDocument()
    })
    await i18n.changeLanguage('en')
    await waitFor(() => {
      expect(screen.getByText(/at least 8 characters/)).toBeInTheDocument()
    })
  })

  it('calls toast.error with conflict message when API returns CONFLICT', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { code: 'CONFLICT', message: 'Conflict' },
      status: 409,
      headers: new Headers(),
    } as never)
    renderPage(<SignupPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(
        'このメールアドレスはすでに使用されています',
      )
    })
  })

  it('calls toast.success on successful signup', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { id: '1', email: 'test@example.com', displayName: 'Test' },
      status: 201,
      headers: new Headers(),
    } as never)
    const onSuccess = vi.fn()
    renderPage(<SignupPage onSuccess={onSuccess} />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('アカウントを作成しました')
      expect(onSuccess).toHaveBeenCalled()
    })
  })

  it('calls signup with displayName undefined when field is left empty', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { id: '1', email: 'test@example.com', displayName: '' },
      status: 201,
      headers: new Headers(),
    } as never)
    renderPage(<SignupPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(signup).toHaveBeenCalledWith(
        expect.objectContaining({ displayName: undefined }),
      )
    })
  })

  it('成功したら、ログイン前に取得したクエリのキャッシュを破棄する', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { id: '1', email: 'test@example.com', displayName: 'Test' },
      status: 201,
      headers: new Headers(),
    } as never)
    const { queryClient } = renderPage(<SignupPage />)
    // 未ログインのときに取得した、閲覧者によって結果が変わるクエリ
    queryClient.setQueryData(['auth', 'me'], null)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0)
    })
  })

  it('失敗したときは、クエリのキャッシュを残す', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { code: 'INTERNAL_ERROR', message: 'internal error' },
      status: 500,
      headers: new Headers(),
    } as never)
    const { queryClient } = renderPage(<SignupPage />)
    queryClient.setQueryData(['auth', 'me'], null)
    mockToastError.mockClear()
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /登録/ }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalled()
    })
    expect(queryClient.getQueryCache().getAll()).toHaveLength(1)
  })
})
