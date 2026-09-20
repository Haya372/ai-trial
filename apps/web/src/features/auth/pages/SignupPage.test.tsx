import { fireEvent, render, screen, waitFor } from '@testing-library/react'
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

describe('SignupPage', () => {
  it('renders email, password, and displayName fields with Japanese labels', () => {
    render(<SignupPage />)
    expect(screen.getByLabelText(/メールアドレス/)).toBeInTheDocument()
    expect(screen.getByLabelText(/パスワード/)).toBeInTheDocument()
    expect(screen.getByLabelText(/表示名/)).toBeInTheDocument()
  })

  it('shows validation error for too-short password', async () => {
    render(<SignupPage />)
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

  it('calls toast.error with conflict message when API returns CONFLICT', async () => {
    const { signup } = await import('../../../api/generated')
    vi.mocked(signup).mockResolvedValueOnce({
      data: { code: 'CONFLICT', message: 'Conflict' },
      status: 409,
      headers: new Headers(),
    } as never)
    render(<SignupPage />)
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
    render(<SignupPage onSuccess={onSuccess} />)
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
    render(<SignupPage />)
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
})
