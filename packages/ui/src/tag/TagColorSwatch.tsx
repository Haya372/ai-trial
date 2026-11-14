import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'
import { type TagColorKey, tagColorVars } from './tagColorTokens'

const swatchVariants = cva('inline-block shrink-0 rounded-full', {
  variants: {
    size: {
      sm: 'size-2',
      md: 'size-3',
      lg: 'size-6',
    },
  },
  defaultVariants: {
    size: 'md',
  },
})

type TagColorSwatchProps = VariantProps<typeof swatchVariants> & {
  color: TagColorKey
}

// Decorative only: the tag name or the palette's aria-label carries the meaning.
export function TagColorSwatch({ color, size }: TagColorSwatchProps) {
  return (
    <span
      data-slot="tag-color-swatch"
      aria-hidden="true"
      className={cn(swatchVariants({ size }))}
      style={{ backgroundColor: tagColorVars(color).background }}
    />
  )
}
