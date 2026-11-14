import { cn } from 'cn'
import { type KeyboardEvent, useRef } from 'react'
import { tagColorForKey } from './tagColorNavigation'
import {
  TAG_COLOR_KEYS,
  type TagColorKey,
  tagColorVars,
} from './tagColorTokens'

interface TagColorPaletteProps {
  value: TagColorKey
  onChange: (color: TagColorKey) => void
  // Injected so the host app can localize them (apps/web uses i18next).
  colorLabels: Record<TagColorKey, string>
  'aria-label'?: string
  'aria-labelledby'?: string
  disabled?: boolean
}

export function TagColorPalette({
  value,
  onChange,
  colorLabels,
  'aria-label': ariaLabel,
  'aria-labelledby': ariaLabelledBy,
  disabled = false,
}: TagColorPaletteProps) {
  const buttonRefs = useRef(new Map<TagColorKey, HTMLButtonElement>())

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const next = tagColorForKey(value, event.key)
    if (!next) return
    event.preventDefault()
    onChange(next)
    buttonRefs.current.get(next)?.focus()
  }

  return (
    <div
      data-slot="tag-color-palette"
      role="radiogroup"
      aria-label={ariaLabel}
      aria-labelledby={ariaLabelledBy}
      aria-disabled={disabled || undefined}
      className="grid grid-cols-6 gap-2"
      onKeyDown={disabled ? undefined : handleKeyDown}
    >
      {TAG_COLOR_KEYS.map((color) => {
        const checked = color === value
        return (
          <button
            key={color}
            ref={(element) => {
              if (element) buttonRefs.current.set(color, element)
              else buttonRefs.current.delete(color)
            }}
            type="button"
            role="radio"
            aria-checked={checked}
            aria-label={colorLabels[color]}
            title={colorLabels[color]}
            // Roving tabindex: Tab enters at the selected color, arrows move within the group.
            tabIndex={checked ? 0 : -1}
            disabled={disabled}
            onClick={() => onChange(color)}
            className={cn(
              'flex size-8 cursor-pointer items-center justify-center rounded-full outline-none transition-shadow',
              'focus-visible:ring-3 focus-visible:ring-ring/50',
              'disabled:pointer-events-none disabled:opacity-50',
              checked &&
                'ring-2 ring-foreground ring-offset-2 ring-offset-background',
            )}
            style={{
              backgroundColor: tagColorVars(color).background,
              color: tagColorVars(color).foreground,
            }}
          >
            {checked && (
              <svg viewBox="0 0 14 14" className="size-4" aria-hidden="true">
                <path
                  d="M2 7L5.5 10.5L12 4"
                  stroke="currentColor"
                  strokeWidth="2"
                  fill="none"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            )}
          </button>
        )
      })}
    </div>
  )
}
