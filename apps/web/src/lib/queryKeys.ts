import type { QueryKey } from '@tanstack/react-query'

export const eventsKeys = {
  all: ['events'] as const,
  // The 'list' marker keeps parseList from mistaking other events keys of the
  // same length (e.g. a future detail key) for list keys.
  list: (startAt: string, endAt: string, tagId: string | null = null) =>
    [...eventsKeys.all, 'list', startAt, endAt, tagId] as const,
  // Callers that inspect cached keys (e.g. predicates) go through this so they
  // do not depend on the positions inside list()'s array.
  parseList: (queryKey: QueryKey) => {
    if (
      queryKey.length !== 5 ||
      queryKey[0] !== eventsKeys.all[0] ||
      queryKey[1] !== 'list'
    ) {
      return undefined
    }
    const [, , startAt, endAt, tagId] = queryKey
    return { startAt, endAt, tagId }
  },
}

export const sharesKeys = {
  all: ['shares'] as const,
  detail: (token: string) => [...sharesKeys.all, token] as const,
}

export const tagsKeys = {
  all: ['tags'] as const,
}
