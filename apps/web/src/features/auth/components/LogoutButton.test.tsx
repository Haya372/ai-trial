import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { logout } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'
import LogoutButton from './LogoutButton'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, logout: vi.fn() }
})

const { mockNavigate } = vi.hoisted(() => ({ mockNavigate: vi.fn() }))
vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return { ...actual, useNavigate: () => mockNavigate }
})

const { mockToastError } = vi.hoisted(() => ({ mockToastError: vi.fn() }))
vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    toast: Object.assign(vi.fn(), { error: mockToastError, success: vi.fn() }),
  }
})

const user = {
  id: '00000000-0000-4000-8000-000000000000',
  email: 'test@example.com',
  displayName: 'Test',
}

describe('LogoutButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ user: user as never })
  })

  it('calls logout API, clears auth store and navigates to /login', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 204 } as never)
    render(<LogoutButton />)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' })
    })
    expect(logout).toHaveBeenCalledTimes(1)
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('clears auth state when the session is already invalid (401)', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 401 } as never)
    render(<LogoutButton />)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' })
    })
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('keeps auth state and shows error toast when the server fails (500)', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 500 } as never)
    render(<LogoutButton />)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('ログアウトに失敗しました')
    })
    expect(useAuthStore.getState().user).not.toBeNull()
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('keeps auth state and shows error toast on network error', async () => {
    vi.mocked(logout).mockRejectedValue(new Error('network'))
    render(<LogoutButton />)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('ログアウトに失敗しました')
    })
    expect(useAuthStore.getState().user).not.toBeNull()
  })
})
