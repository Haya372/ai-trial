import { Input as InputPrimitive } from '@base-ui/react/input'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const inputVariants = cva(
  'flex w-full rounded-lg border bg-background text-foreground transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20',
  {
    variants: {
      size: {
        sm: 'h-7 px-2.5 text-[0.8rem]',
        md: 'h-9 px-3 text-sm',
        lg: 'h-10 px-3.5 text-base',
      },
      state: {
        default: 'border-input',
        error: 'border-destructive ring-3 ring-destructive/20',
      },
    },
    defaultVariants: {
      size: 'md',
      state: 'default',
    },
  },
)

type InputProps = Omit<InputPrimitive.Props, 'className' | 'size'> &
  VariantProps<typeof inputVariants>

function Input({ size, state, ...props }: InputProps) {
  return (
    <InputPrimitive
      data-slot="input"
      className={cn(inputVariants({ size, state }))}
      {...props}
    />
  )
}

export { Input }
