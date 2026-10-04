import { useQuery } from '@tanstack/react-query'
import { getShareByToken } from '../../../api/generated'
import { sharesKeys } from '../../../lib/queryKeys'
import { toShareDetailError } from '../utils'
import type { ShareDetailError } from '../utils'

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
