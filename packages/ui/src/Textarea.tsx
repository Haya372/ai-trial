import type * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const textareaVariants = cva(
  'flex w-full rounded-lg border bg-background px-3 py-2 text-sm text-foreground transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20',
  {
    variants: {
      state: {
        default: 'border-input',
        error: 'border-destructive ring-3 ring-destructive/20',
      },
    },
    defaultVariants: {
      state: 'default',
    },
  },
)

type TextareaProps = Omit<
  React.TextareaHTMLAttributes<HTMLTextAreaElement>,
  'className'
> &
  VariantProps<typeof textareaVariants>

function Textarea({ state, ...props }: TextareaProps) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(textareaVariants({ state }))}
      {...props}
    />
  )
}

export { Textarea }
