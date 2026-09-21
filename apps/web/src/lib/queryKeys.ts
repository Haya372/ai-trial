export const eventsKeys = {
  all: ['events'] as const,
  list: (startDate: string, endDate: string) =>
    [...eventsKeys.all, startDate, endDate] as const,
}
