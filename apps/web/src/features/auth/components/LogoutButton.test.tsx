import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
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

function renderWithClient(queryClient = new QueryClient()) {
  return render(
    <QueryClientProvider client={queryClient}>
      <LogoutButton />
    </QueryClientProvider>,
  )
}

describe('LogoutButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ user: user as never })
  })

  it('calls logout API, clears auth store and navigates to /login', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 204 } as never)
    renderWithClient()
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' })
    })
    expect(logout).toHaveBeenCalledTimes(1)
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('clears auth state when the session is already invalid (401)', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 401 } as never)
    renderWithClient()
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' })
    })
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('keeps auth state and shows error toast when the server fails (500)', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 500 } as never)
    renderWithClient()
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('ログアウトに失敗しました')
    })
    expect(useAuthStore.getState().user).not.toBeNull()
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('keeps auth state and shows error toast on network error', async () => {
    vi.mocked(logout).mockRejectedValue(new Error('network'))
    renderWithClient()
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('ログアウトに失敗しました')
    })
    expect(useAuthStore.getState().user).not.toBeNull()
  })

  it('clears the query cache so the next user cannot see stale data', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 204 } as never)
    const queryClient = new QueryClient()
    queryClient.setQueryData(['events'], [{ id: 'prev-user-event' }])
    renderWithClient(queryClient)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' })
    })
    expect(queryClient.getQueryData(['events'])).toBeUndefined()
  })

  it('keeps the query cache when logout fails', async () => {
    vi.mocked(logout).mockResolvedValue({ status: 500 } as never)
    const queryClient = new QueryClient()
    queryClient.setQueryData(['events'], [{ id: 'event' }])
    renderWithClient(queryClient)
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }))
    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalled()
    })
    expect(queryClient.getQueryData(['events'])).toBeDefined()
  })
})
