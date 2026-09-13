import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const textVariants = cva('', {
  variants: {
    variant: {
      h1: 'scroll-m-20 text-3xl font-bold leading-tight tracking-tight',
      h2: 'scroll-m-20 text-2xl font-semibold leading-tight tracking-tight',
      h3: 'scroll-m-20 text-xl font-semibold leading-tight tracking-tight',
      body: 'text-base leading-normal',
      caption: 'text-sm leading-normal text-muted-foreground',
      code: 'font-mono text-sm rounded-md bg-muted px-1.5 py-0.5',
    },
  },
  defaultVariants: {
    variant: 'body',
  },
})

const elementMap = {
  h1: 'h1',
  h2: 'h2',
  h3: 'h3',
  body: 'p',
  caption: 'span',
  code: 'code',
} as const

type TextVariant = NonNullable<VariantProps<typeof textVariants>['variant']>

type TextProps = {
  variant?: TextVariant
  children?: React.ReactNode
}

function Text({ variant = 'body', children }: TextProps) {
  const Element = elementMap[variant]
  return <Element className={cn(textVariants({ variant }))}>{children}</Element>
}

export { Text }
