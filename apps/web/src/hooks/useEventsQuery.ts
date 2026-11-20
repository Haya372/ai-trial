import { useQuery } from '@tanstack/react-query'
import { getEvents } from '../api/generated'
import { eventsKeys } from '../lib/queryKeys'

// 検索範囲 [startAt, endAt) をUTCで送る（ADR-036）
export function useEventsQuery(
  startAt: Date,
  endAt: Date,
  tagId: string | null = null,
) {
  const startAtIso = startAt.toISOString()
  const endAtIso = endAt.toISOString()
  return useQuery({
    queryKey: eventsKeys.list(startAtIso, endAtIso, tagId),
    queryFn: async () => {
      const res = await getEvents({
        startAt: startAtIso,
        endAt: endAtIso,
        ...(tagId !== null && { tagId }),
      })
      if (res.status !== 200) {
        throw res
      }
      return res.data
    },
    // タグの絞り込みを切り替えても、取得中にカレンダーがローディング表示へ
    // 入れ替わらないよう前のデータを残す。期間が変わったときは別の期間の
    // 予定を見せてしまうため残さない
    placeholderData: (previousData, previousQuery) => {
      const prev = previousQuery && eventsKeys.parseList(previousQuery.queryKey)
      return prev?.startAt === startAtIso && prev.endAt === endAtIso
        ? previousData
        : undefined
    },
    staleTime: 5 * 60 * 1000, // 5分キャッシュ
  })
}
