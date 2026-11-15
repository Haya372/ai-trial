import { tagColorVars } from '../tag/tagColorTokens'
import type { TagColorKey } from '../tag/tagColorTokens'

interface CalendarEventContentProps {
  timeText: string
  title: string
  tagColors?: TagColorKey[]
  timeClass?: string
  titleClass?: string
}

export function CalendarEventContent({
  timeText,
  title,
  tagColors,
  timeClass,
  titleClass,
}: CalendarEventContentProps) {
  // Dots show the 2nd and later tag colors; the 1st is already the event background.
  const dotColors = tagColors && tagColors.length > 1 ? tagColors.slice(1) : []
  // Ring color matches the event background's foreground so dots stay visible
  // even when a dot color coincides with the event background.
  const ringColor =
    tagColors && tagColors.length > 0
      ? tagColorVars(tagColors[0]).foreground
      : undefined

  return (
    // Fragment, not a wrapper: FullCalendar's theme styles the direct children of
    // the event inner (gap, column stacking in week view, title truncation).
    <>
      {timeText && <div className={timeClass}>{timeText}</div>}
      <div className={titleClass}>{title}</div>
      {dotColors.length > 0 && (
        <div
          data-slot="calendar-event-tag-dots"
          aria-hidden="true"
          // One row tall plus 1px padding so the dot ring isn't clipped; wrapped dots fall outside and are hidden.
          className="flex h-2.5 min-w-0 flex-wrap gap-0.5 overflow-hidden p-px"
        >
          {dotColors.map((color, i) => (
            <span
              // Same color may appear twice (duplicate colors are allowed), so pair it with the position.
              key={`${color}-${i}`}
              className="inline-block shrink-0 size-2 rounded-full"
              style={{
                backgroundColor: tagColorVars(color).background,
                boxShadow: `0 0 0 1px ${ringColor}`,
              }}
            />
          ))}
        </div>
      )}
    </>
  )
}
