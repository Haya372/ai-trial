import { useQuery } from '@tanstack/react-query'
import { getEvents } from '../api/generated'
import { pad } from '../lib/dateFormat'
import { eventsKeys } from '../lib/queryKeys'

// YYYY-MM-DD 形式にフォーマットする
function formatDate(date: Date): string {
  const y = date.getFullYear()
  const m = pad(date.getMonth() + 1)
  const d = pad(date.getDate())
  return `${y}-${m}-${d}`
}

export function useEventsQuery(startDate: Date, endDate: Date) {
  return useQuery({
    queryKey: eventsKeys.list(formatDate(startDate), formatDate(endDate)),
    queryFn: async () => {
      const res = await getEvents({
        startDate: formatDate(startDate),
        endDate: formatDate(endDate),
      })
      if (res.status !== 200) {
        throw res
      }
      return res.data
    },
    staleTime: 5 * 60 * 1000, // 5分キャッシュ
  })
}
