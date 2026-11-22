import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactElement, ReactNode } from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import i18n from 'i18next'
import { describe, expect, it, vi } from 'vitest'
import LoginPage from './LoginPage'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, login: vi.fn() }
})

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return {
    ...actual,
    Link: ({ to, children }: { to: string; children: ReactNode }) => (
      <a href={to}>{children}</a>
    ),
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

describe('LoginPage', () => {
  it('renders email and password fields with Japanese labels', () => {
    renderPage(<LoginPage />)
    expect(screen.getByLabelText(/メールアドレス/)).toBeInTheDocument()
    expect(screen.getByLabelText(/パスワード/)).toBeInTheDocument()
  })

  it('shows validation error when email is empty and form submitted', async () => {
    renderPage(<LoginPage />)
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(
        screen.getByText(/メールアドレスを入力してください/),
      ).toBeInTheDocument()
    })
  })

  it('shows validation error when password is empty', async () => {
    renderPage(<LoginPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(
        screen.getByText(/パスワードを入力してください/),
      ).toBeInTheDocument()
    })
  })

  it('reserves the validation message slot even when there is no error', () => {
    const { container } = renderPage(<LoginPage />)
    expect(
      container.querySelectorAll('[data-slot="form-message"]'),
    ).toHaveLength(2)
  })

  it('renders a link to the signup page', () => {
    renderPage(<LoginPage />)
    const link = screen.getByRole('link', {
      name: /アカウントをお持ちでない方はこちら/,
    })
    expect(link).toHaveAttribute('href', '/signup')
  })

  it('updates a displayed validation error to the new language when the language changes', async () => {
    renderPage(<LoginPage />)
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(
        screen.getByText(/メールアドレスを入力してください/),
      ).toBeInTheDocument()
    })
    await i18n.changeLanguage('en')
    await waitFor(() => {
      expect(screen.getByText(/Email is required/)).toBeInTheDocument()
    })
  })

  it('送信中はログインボタンをローディング状態にする', async () => {
    const { login } = await import('../../../api/generated')
    vi.mocked(login).mockReturnValueOnce(new Promise(() => {}))
    renderPage(<LoginPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /ログイン/ })
      expect(button).toBeDisabled()
      expect(button).toHaveAttribute('aria-busy', 'true')
    })
  })

  it('calls toast.error with unauthorized message when API returns UNAUTHORIZED', async () => {
    const { login } = await import('../../../api/generated')
    vi.mocked(login).mockResolvedValueOnce({
      data: { code: 'UNAUTHORIZED', message: 'Unauthorized' },
      status: 401,
      headers: new Headers(),
    } as never)
    renderPage(<LoginPage />)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(
        'メールアドレスまたはパスワードが正しくありません',
      )
    })
  })

  it('成功したら、ログイン前に取得したクエリのキャッシュを破棄する', async () => {
    const { login } = await import('../../../api/generated')
    vi.mocked(login).mockResolvedValueOnce({
      data: { id: '1', email: 'test@example.com', displayName: 'Test' },
      status: 200,
      headers: new Headers(),
    } as never)
    const { queryClient } = renderPage(<LoginPage />)
    // 未ログインのときに取得した、閲覧者によって結果が変わるクエリ
    queryClient.setQueryData(['auth', 'me'], null)
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0)
    })
  })

  it('失敗したときは、クエリのキャッシュを残す', async () => {
    const { login } = await import('../../../api/generated')
    vi.mocked(login).mockResolvedValueOnce({
      data: { code: 'INTERNAL_ERROR', message: 'internal error' },
      status: 500,
      headers: new Headers(),
    } as never)
    const { queryClient } = renderPage(<LoginPage />)
    queryClient.setQueryData(['auth', 'me'], null)
    mockToastError.mockClear()
    fireEvent.change(screen.getByLabelText(/メールアドレス/), {
      target: { value: 'test@example.com' },
    })
    fireEvent.change(screen.getByLabelText(/パスワード/), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalled()
    })
    expect(queryClient.getQueryCache().getAll()).toHaveLength(1)
  })
})
