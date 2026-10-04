import { useQuery } from '@tanstack/react-query'
import { getMe } from '../../../api/generated'
import type { UserResponse } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'

export function useViewerAuthQuery(): {
  data: UserResponse | null
  isPending: boolean
} {
  const setUser = useAuthStore((s) => s.setUser)
  const clearUser = useAuthStore((s) => s.clearUser)

  const query = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const res = await getMe()
      if (res.status === 200) {
        setUser(res.data)
        return res.data
      }
      // 401 はセッション切れ: ストアの古いユーザー情報を明示的に破棄する
      if (res.status === 401) {
        clearUser()
        return null
      }
      // 5xx など認証エラーと判断できない場合は再スローし、ネットワーク障害と区別する
      throw new Error(`auth check failed: ${res.status}`)
    },
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  return {
    data: query.data ?? null,
    isPending: query.isPending,
  }
}
