import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const spinnerVariants = cva(
  'inline-block shrink-0 animate-spin rounded-full border-2 border-current border-r-transparent',
  {
    variants: {
      size: {
        sm: 'size-4',
        md: 'size-5',
        lg: 'size-8',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  },
)

type SpinnerProps = VariantProps<typeof spinnerVariants> & {
  // Without a label the spinner is decorative; the surrounding element must convey the busy state.
  label?: string
}

function Spinner({ size, label }: SpinnerProps) {
  const a11yProps = label
    ? { role: 'status', 'aria-label': label }
    : { 'aria-hidden': true }

  return (
    <span
      data-slot="spinner"
      className={cn(spinnerVariants({ size }))}
      {...a11yProps}
    />
  )
}

export { Spinner }
