import { Separator } from '@base-ui/react/separator'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const dividerVariants = cva('shrink-0 bg-border', {
  variants: {
    orientation: {
      horizontal: 'h-px w-full',
      vertical: 'h-full w-px',
    },
  },
  defaultVariants: {
    orientation: 'horizontal',
  },
})

type DividerProps = Omit<Separator.Props, 'className'> &
  VariantProps<typeof dividerVariants>

function Divider({ orientation = 'horizontal', ...props }: DividerProps) {
  return (
    <Separator
      data-slot="divider"
      orientation={orientation}
      className={cn(dividerVariants({ orientation }))}
      {...props}
    />
  )
}

export { Divider }
