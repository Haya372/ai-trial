import { TAG_COLOR_KEYS, type TagColorKey } from './tagColorTokens'

const STEP_BY_KEY: Record<string, number> = {
  ArrowRight: 1,
  ArrowDown: 1,
  ArrowLeft: -1,
  ArrowUp: -1,
}

// Follows the WAI-ARIA radio group pattern: arrows move and wrap, Home/End jump to the ends.
export function tagColorForKey(
  current: TagColorKey,
  key: string,
): TagColorKey | undefined {
  if (key === 'Home') return TAG_COLOR_KEYS[0]
  if (key === 'End') return TAG_COLOR_KEYS[TAG_COLOR_KEYS.length - 1]

  const step = STEP_BY_KEY[key]
  if (step === undefined) return undefined

  const count = TAG_COLOR_KEYS.length
  const index = TAG_COLOR_KEYS.indexOf(current)
  return TAG_COLOR_KEYS[(index + step + count) % count]
}
