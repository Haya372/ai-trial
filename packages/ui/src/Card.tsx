import { cn } from 'cn'

type CardProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function Card({ ...props }: CardProps) {
  return (
    <div
      data-slot="card"
      className={cn('rounded-xl border bg-card text-card-foreground shadow-sm')}
      {...props}
    />
  )
}

type CardHeaderProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function CardHeader({ ...props }: CardHeaderProps) {
  return (
    <div
      data-slot="card-header"
      className={cn('flex flex-col space-y-1.5 p-6')}
      {...props}
    />
  )
}

type CardTitleProps = Omit<
  React.HTMLAttributes<HTMLHeadingElement>,
  'className'
>

function CardTitle({ ...props }: CardTitleProps) {
  return (
    <h3
      data-slot="card-title"
      className={cn('text-lg font-semibold leading-none tracking-tight')}
      {...props}
    />
  )
}

type CardDescriptionProps = Omit<
  React.HTMLAttributes<HTMLParagraphElement>,
  'className'
>

function CardDescription({ ...props }: CardDescriptionProps) {
  return (
    <p
      data-slot="card-description"
      className={cn('text-sm text-muted-foreground')}
      {...props}
    />
  )
}

type CardContentProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function CardContent({ ...props }: CardContentProps) {
  return <div data-slot="card-content" className={cn('p-6 pt-0')} {...props} />
}

type CardFooterProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'className'>

function CardFooter({ ...props }: CardFooterProps) {
  return (
    <div
      data-slot="card-footer"
      className={cn('flex items-center p-6 pt-0')}
      {...props}
    />
  )
}

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter }
