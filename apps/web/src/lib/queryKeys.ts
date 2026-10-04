export const eventsKeys = {
  all: ['events'] as const,
  list: (startDate: string, endDate: string) =>
    [...eventsKeys.all, startDate, endDate] as const,
}

export const sharesKeys = {
  all: ['shares'] as const,
  detail: (token: string) => [...sharesKeys.all, token] as const,
}
