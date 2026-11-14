import './tagColors.css'

// Mirrors the TagColor enum in packages/schema; the order is the palette order defined by SPEC-008.
export const TAG_COLOR_KEYS = [
  'red',
  'orange',
  'yellow',
  'green',
  'teal',
  'cyan',
  'blue',
  'indigo',
  'purple',
  'pink',
  'brown',
  'gray',
] as const

export type TagColorKey = (typeof TAG_COLOR_KEYS)[number]

export interface TagColorVars {
  background: string
  foreground: string
}

export function tagColorVars(key: TagColorKey): TagColorVars {
  return {
    background: `var(--tag-${key}-bg)`,
    foreground: `var(--tag-${key}-fg)`,
  }
}
