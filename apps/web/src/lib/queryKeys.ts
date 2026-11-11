export const eventsKeys = {
  all: ['events'] as const,
  list: (startAt: string, endAt: string) =>
    [...eventsKeys.all, startAt, endAt] as const,
}

export const sharesKeys = {
  all: ['shares'] as const,
  detail: (token: string) => [...sharesKeys.all, token] as const,
}
