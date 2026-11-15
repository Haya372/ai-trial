import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { deleteTag, type TagResponse } from '../../../api/generated'
import { eventsKeys, tagsKeys } from '../../../lib/queryKeys'

export function useTagDelete(onDeleted: () => void, onFailed: () => void) {
  const queryClient = useQueryClient()
  const [target, setTarget] = useState<TagResponse | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const confirmDelete = async () => {
    if (!target) return
    setIsDeleting(true)
    try {
      const res = await deleteTag(target.id)
      if (res.status !== 204) {
        onFailed()
        return
      }
      // Drop the tag from the cache before refetching, so a failed refetch still
      // lets the calendar see it is gone and clear a filter that uses it.
      queryClient.setQueryData<TagResponse[]>(tagsKeys.all, (tags) =>
        tags?.filter((tag) => tag.id !== target.id),
      )
      await queryClient.invalidateQueries({ queryKey: tagsKeys.all })
      // Deleting detaches the tag from every event, which changes their calendar colors.
      // Lists filtered by the deleted tag are skipped: the calendar clears that filter
      // because the tag is gone from the cache, so refetching them would only fetch
      // an empty result.
      await queryClient.invalidateQueries({
        queryKey: eventsKeys.all,
        predicate: (query) =>
          eventsKeys.parseList(query.queryKey)?.tagId !== target.id,
      })
      onDeleted()
    } catch {
      onFailed()
    } finally {
      setIsDeleting(false)
      setTarget(null)
    }
  }

  return {
    target,
    requestDelete: setTarget,
    cancelDelete: () => setTarget(null),
    confirmDelete,
    isDeleting,
  }
}
