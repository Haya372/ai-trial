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
      await queryClient.invalidateQueries({ queryKey: tagsKeys.all })
      // Deleting detaches the tag from every event, which changes their calendar colors.
      // Lists filtered by the deleted tag are skipped: the calendar clears that filter
      // once the refetched tags no longer include it, so refetching them would only
      // fetch an empty result.
      await queryClient.invalidateQueries({
        queryKey: eventsKeys.all,
        predicate: (query) => query.queryKey[3] !== target.id,
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
