import { useQuery } from '@tanstack/react-query'
import { getEvents } from '../api/generated'

// YYYY-MM-DD 形式にフォーマットする
function formatDate(date: Date): string {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

export function useEventsQuery(startDate: Date, endDate: Date) {
  return useQuery({
    queryKey: ['events', formatDate(startDate), formatDate(endDate)],
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
