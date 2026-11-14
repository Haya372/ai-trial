import { useQuery } from '@tanstack/react-query'
import { listTags } from '../../../api/generated'
import { tagsKeys } from '../../../lib/queryKeys'

export function useTagsQuery(enabled: boolean) {
  return useQuery({
    queryKey: tagsKeys.all,
    queryFn: async () => {
      const res = await listTags()
      if (res.status !== 200) {
        throw res
      }
      return res.data.tags
    },
    enabled,
  })
}
