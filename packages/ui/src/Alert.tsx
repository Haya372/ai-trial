import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const alertVariants = cva(
  'relative w-full rounded-lg border p-4 [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg+div]:pl-7',
  {
    variants: {
      variant: {
        default: 'bg-background text-foreground border-border',
        info: 'bg-primary/5 text-foreground border-primary/20 [&>svg]:text-primary',
        success:
          'bg-green-50 text-green-900 border-green-200 dark:bg-green-950 dark:text-green-100 dark:border-green-800 [&>svg]:text-green-600',
        warning:
          'bg-yellow-50 text-yellow-900 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-100 dark:border-yellow-800 [&>svg]:text-yellow-600',
        destructive:
          'bg-destructive/10 text-destructive border-destructive/30 [&>svg]:text-destructive',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

type AlertProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'> &
  VariantProps<typeof alertVariants>

function Alert({ variant, ...props }: AlertProps) {
  return (
    <div
      data-slot="alert"
      role="alert"
      className={cn(alertVariants({ variant }))}
      {...props}
    />
  )
}

type AlertTitleProps = Omit<
  React.HTMLAttributes<HTMLHeadingElement>,
  'className'
>

function AlertTitle(props: AlertTitleProps) {
  return (
    <h5
      data-slot="alert-title"
      className={cn('mb-1 font-medium leading-tight tracking-tight')}
      {...props}
    />
  )
}

type AlertDescriptionProps = Omit<
  React.HTMLAttributes<HTMLParagraphElement>,
  'className'
>

function AlertDescription(props: AlertDescriptionProps) {
  return (
    <p
      data-slot="alert-description"
      className={cn('text-sm leading-normal opacity-90')}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription }
