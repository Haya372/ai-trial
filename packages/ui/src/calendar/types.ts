import type { TagColorKey } from '../tag/tagColorTokens'

export interface CalendarEvent {
  id: string
  title: string
  start: Date
  end: Date
  // True for an event added via EventSubscription (read-only); rendered
  // with a visually distinct style from the caller's own events.
  isSubscribed?: boolean
  // Colors of the attached tags in insertion order. The first one becomes the
  // event background; the rest are shown as dots. Omitted/empty → default color.
  tagColors?: TagColorKey[]
}
