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
    <div>
      {timeText && <span className={timeClass}>{timeText}</span>}
      <span className={titleClass}>{title}</span>
      {dotColors.length > 0 && (
        <div
          data-slot="calendar-event-tag-dots"
          aria-hidden="true"
          className="flex h-2 flex-wrap overflow-hidden gap-0.5"
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
    </div>
  )
}
