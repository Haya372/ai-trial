import type { CalendarEvent } from './types'

// Visually distinguishes a subscribed (read-only, added via EventSubscription)
// event from the viewer's own events: a dashed border and reduced opacity,
// similar to how most calendar apps mute events from other calendars.
export function eventClassNames(e: CalendarEvent): string[] {
  return e.isSubscribed ? ['opacity-70', 'border', 'border-dashed'] : []
}
