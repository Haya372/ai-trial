import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { AxiosError } from 'axios'
import { describe, expect, it, vi } from 'vitest'
import LoginPage from './LoginPage'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, login: vi.fn() }
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

function makeAxiosError(code: string) {
  return new AxiosError(
    'Request failed',
    'ERR_BAD_REQUEST',
    undefined,
    undefined,
    {
      data: { code, message: 'error' },
      status: 401,
      statusText: '',
      headers: {},
      config: {} as never,
    },
  )
}

describe('LoginPage', () => {
  it('renders email and password fields with Japanese labels', () => {
    render(<LoginPage />)
    expect(screen.getByLabelText(/メールアドレス/)).toBeInTheDocument()
    expect(screen.getByLabelText(/パスワード/)).toBeInTheDocument()
  })

  it('shows validation error when email is empty and form submitted', async () => {
    render(<LoginPage />)
    fireEvent.click(screen.getByRole('button', { name: /ログイン/ }))
    await waitFor(() => {
      expect(
        screen.getByText(/メールアドレスを入力してください/),
      ).toBeInTheDocument()
    })
  })

  it('shows validation error when password is empty', async () => {
    render(<LoginPage />)
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

  it('calls toast.error with unauthorized message when API returns UNAUTHORIZED', async () => {
    const { login } = await import('../../../api/generated')
    vi.mocked(login).mockRejectedValueOnce(makeAxiosError('UNAUTHORIZED'))
    render(<LoginPage />)
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
})
