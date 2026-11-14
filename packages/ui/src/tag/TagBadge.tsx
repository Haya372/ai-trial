import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'
import { type TagColorKey, tagColorVars } from './tagColorTokens'

const tagBadgeVariants = cva(
  'inline-flex max-w-full items-center rounded-full font-medium',
  {
    variants: {
      size: {
        sm: 'h-5 px-2 text-xs',
        md: 'h-6 px-2.5 text-sm',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  },
)

type TagBadgeProps = VariantProps<typeof tagBadgeVariants> & {
  name: string
  color: TagColorKey
}

export function TagBadge({ name, color, size }: TagBadgeProps) {
  const { background, foreground } = tagColorVars(color)
  return (
    <span
      data-slot="tag-badge"
      className={cn(tagBadgeVariants({ size }))}
      style={{ backgroundColor: background, color: foreground }}
      title={name}
    >
      <span className="truncate">{name}</span>
    </span>
  )
}
