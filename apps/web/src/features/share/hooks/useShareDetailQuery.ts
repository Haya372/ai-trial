import { useQuery } from '@tanstack/react-query'
import { getShareByToken } from '../../../api/generated'
import { toShareDetailError } from '../utils'
import type { ShareDetailError } from '../utils'

const sharesKeys = {
  all: ['shares'] as const,
  detail: (token: string) => [...sharesKeys.all, token] as const,
}

export function useShareDetailQuery(token: string) {
  return useQuery({
    queryKey: sharesKeys.detail(token),
    queryFn: async () => {
      const res = await getShareByToken(token)
      if (res.status === 200) {
        return res.data
      }
      throw toShareDetailError(res.status, res.data)
    },
    retry: false,
    staleTime: 0,
  })
}

export type { ShareDetailError }
