import { tagColorVars } from '../tag/tagColorTokens'
import type { CalendarEvent } from './types'

// Returns the FullCalendar EventInput color props for the given event.
// When no tags are attached, returns {} so FullCalendar keeps its default color.
export function eventColorProps(e: CalendarEvent): {
  color?: string
  contrastColor?: string
} {
  if (!e.tagColors || e.tagColors.length === 0) return {}
  const { background, foreground } = tagColorVars(e.tagColors[0])
  return { color: background, contrastColor: foreground }
}
