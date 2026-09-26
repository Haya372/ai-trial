import { QueryClient } from '@tanstack/react-query'
import { describe, expect, it, vi } from 'vitest'
import { eventsKeys } from '../../lib/queryKeys'
import { runEventMutation } from './runEventMutation'

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

function createQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

describe('runEventMutation', () => {
  it('期待するstatusが返った場合、クエリを無効化し成功トーストを表示してonSuccessを呼ぶ', async () => {
    const queryClient = createQueryClient()
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
    const onSuccess = vi.fn()

    await runEventMutation({
      queryClient,
      request: Promise.resolve({ status: 204, data: undefined }),
      expectedStatus: 204,
      mode: 'delete',
      successMessage: '予定を削除しました',
      onSuccess,
    })

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: eventsKeys.all })
    expect(mockToastSuccess).toHaveBeenCalledWith('予定を削除しました')
    expect(onSuccess).toHaveBeenCalled()
  })

  it('期待しないstatusが返った場合、エラートーストを表示しonSuccessを呼ばない', async () => {
    const queryClient = createQueryClient()
    const onSuccess = vi.fn()

    await runEventMutation({
      queryClient,
      request: Promise.resolve({
        status: 500,
        data: { code: 'INTERNAL_ERROR' },
      }),
      expectedStatus: 204,
      mode: 'delete',
      successMessage: '予定を削除しました',
      onSuccess,
    })

    expect(mockToastError).toHaveBeenCalledWith(
      'サーバーエラーが発生しました。しばらく経ってから再試行してください',
    )
    expect(onSuccess).not.toHaveBeenCalled()
  })

  it('requestが例外を投げた場合、エラートーストを表示しonSuccessを呼ばない', async () => {
    const queryClient = createQueryClient()
    const onSuccess = vi.fn()

    await runEventMutation({
      queryClient,
      request: Promise.reject(new Error('network error')),
      expectedStatus: 204,
      mode: 'delete',
      successMessage: '予定を削除しました',
      onSuccess,
    })

    expect(mockToastError).toHaveBeenCalledWith('予定の削除に失敗しました')
    expect(onSuccess).not.toHaveBeenCalled()
  })
})
