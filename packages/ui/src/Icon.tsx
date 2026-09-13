import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const iconVariants = cva('inline-flex shrink-0 items-center justify-center', {
  variants: {
    size: {
      xs: '[&>svg]:h-3 [&>svg]:w-3',
      sm: '[&>svg]:h-4 [&>svg]:w-4',
      md: '[&>svg]:h-5 [&>svg]:w-5',
      lg: '[&>svg]:h-6 [&>svg]:w-6',
      xl: '[&>svg]:h-8 [&>svg]:w-8',
    },
  },
  defaultVariants: {
    size: 'md',
  },
})

type IconProps = VariantProps<typeof iconVariants> & {
  children?: React.ReactNode
}

function Icon({ size, children }: IconProps) {
  return (
    <span data-slot="icon" className={cn(iconVariants({ size }))}>
      {children}
    </span>
  )
}

export { Icon }
