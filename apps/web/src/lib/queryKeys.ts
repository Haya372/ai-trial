export const eventsKeys = {
  all: ['events'] as const,
  list: (startAt: string, endAt: string, tagId?: string) =>
    [...eventsKeys.all, startAt, endAt, tagId ?? null] as const,
}

export const sharesKeys = {
  all: ['shares'] as const,
  detail: (token: string) => [...sharesKeys.all, token] as const,
}

export const tagsKeys = {
  all: ['tags'] as const,
}
