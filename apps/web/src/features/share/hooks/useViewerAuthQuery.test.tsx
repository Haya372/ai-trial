import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useAuthStore } from '../../../stores/auth'
import { useViewerAuthQuery } from './useViewerAuthQuery'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, getMe: vi.fn() }
})

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  displayName: 'Test User',
}

beforeEach(() => {
  vi.restoreAllMocks()
  useAuthStore.setState({ user: null })
})

afterEach(() => {
  useAuthStore.setState({ user: null })
})

describe('useViewerAuthQuery', () => {
  it('200 のとき data にユーザーが返り、ストアにユーザーが保存される', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockResolvedValueOnce({
      data: mockUser,
      status: 200,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.data).toEqual(mockUser)
    expect(useAuthStore.getState().user).toEqual(mockUser)
  })

  it('401 のとき data が null になり、ストアのユーザーが消去される', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockResolvedValueOnce({
      data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
      status: 401,
      headers: new Headers(),
    } as never)

    // 401 前にストアに古いユーザー情報があった状態を再現
    useAuthStore.setState({ user: mockUser })

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.data).toBeNull()
    // セッション切れ時に古いユーザー情報が残り続けないこと
    expect(useAuthStore.getState().user).toBeNull()
  })

  it('ネットワークエラー時はストアのユーザー情報が保持される', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockRejectedValueOnce(new Error('Network error'))

    // ネットワークエラー前にログイン済み状態を再現
    useAuthStore.setState({ user: mockUser })

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    // ネットワークエラーはセッション切れではないのでストアを消去しない
    expect(useAuthStore.getState().user).toEqual(mockUser)
  })

  it('200 のとき isError は false になる', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockResolvedValueOnce({
      data: mockUser,
      status: 200,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.isError).toBe(false)
  })

  it('401 のとき isError は false になる（未ログインは確認失敗ではない）', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockResolvedValueOnce({
      data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
      status: 401,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isPending).toBe(false))
    expect(result.current.isError).toBe(false)
  })

  it('5xx のとき isError が true になる', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockResolvedValueOnce({
      data: { code: 'INTERNAL_ERROR', message: 'internal error' },
      status: 500,
      headers: new Headers(),
    } as never)

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.data).toBeNull()
  })

  it('ネットワークエラーのとき isError が true になる', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
  })

  it('refetch で確認をやり直し、成功すれば isError が false に戻る', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe)
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce({
        data: mockUser,
        status: 200,
        headers: new Headers(),
      } as never)

    const { result } = renderHook(() => useViewerAuthQuery(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
    result.current.refetch()
    await waitFor(() => expect(result.current.data).toEqual(mockUser))
    expect(result.current.isError).toBe(false)
  })

  it('共有ページを開き直すと、前回の結果を使わずにログイン状態を確認し直す', async () => {
    const { getMe } = await import('../../../api/generated')
    vi.mocked(getMe)
      .mockResolvedValueOnce({
        data: { code: 'UNAUTHORIZED', message: 'unauthorized' },
        status: 401,
        headers: new Headers(),
      } as never)
      .mockResolvedValueOnce({
        data: mockUser,
        status: 200,
        headers: new Headers(),
      } as never)
    // 同じ QueryClient のまま、別の画面でログインしてから戻ってきた状態を再現する
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )

    const first = renderHook(() => useViewerAuthQuery(), { wrapper })
    await waitFor(() => expect(first.result.current.isPending).toBe(false))
    expect(first.result.current.data).toBeNull()
    first.unmount()
    // 実際の画面遷移では、共有ページを離れてから戻るまでに破棄が済んでいる
    await waitFor(() =>
      expect(
        queryClient.getQueryCache().find({ queryKey: ['auth', 'me'] }),
      ).toBeUndefined(),
    )

    const second = renderHook(() => useViewerAuthQuery(), { wrapper })
    await waitFor(() => expect(second.result.current.data).toEqual(mockUser))
  })
})
