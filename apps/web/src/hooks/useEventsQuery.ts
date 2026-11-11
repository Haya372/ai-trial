import { useQuery } from '@tanstack/react-query'
import { getEvents } from '../api/generated'
import { eventsKeys } from '../lib/queryKeys'

// 検索範囲 [startAt, endAt) をUTCで送る（ADR-036）
export function useEventsQuery(startAt: Date, endAt: Date) {
  const startAtIso = startAt.toISOString()
  const endAtIso = endAt.toISOString()
  return useQuery({
    queryKey: eventsKeys.list(startAtIso, endAtIso),
    queryFn: async () => {
      const res = await getEvents({ startAt: startAtIso, endAt: endAtIso })
      if (res.status !== 200) {
        throw res
      }
      return res.data
    },
    staleTime: 5 * 60 * 1000, // 5分キャッシュ
  })
}
