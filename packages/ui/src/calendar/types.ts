export interface CalendarEvent {
  id: string
  title: string
  start: Date
  end: Date
  // True for an event added via EventSubscription (read-only); rendered
  // with a visually distinct style from the caller's own events.
  isSubscribed?: boolean
}
