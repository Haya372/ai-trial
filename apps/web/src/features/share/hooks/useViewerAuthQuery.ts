import { useQuery } from '@tanstack/react-query'
import { getMe } from '../../../api/generated'
import type { UserResponse } from '../../../api/generated'
import { useAuthStore } from '../../../stores/auth'

export function useViewerAuthQuery(): {
  data: UserResponse | null
  isPending: boolean
} {
  const setUser = useAuthStore((s) => s.setUser)

  const query = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      try {
        const res = await getMe()
        if (res.status === 200) {
          setUser(res.data)
          return res.data
        }
        // 401 や 500 は null として扱い、公開閲覧を止めない
        return null
      } catch {
        return null
      }
    },
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  return {
    data: query.data ?? null,
    isPending: query.isPending,
  }
}
