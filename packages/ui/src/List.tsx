import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

const listVariants = cva('', {
  variants: {
    variant: {
      unordered: 'list-disc pl-6',
      ordered: 'list-decimal pl-6',
      none: 'list-none pl-0',
    },
  },
  defaultVariants: {
    variant: 'unordered',
  },
})

type ListProps = Omit<React.HTMLAttributes<HTMLUListElement>, 'className'> &
  VariantProps<typeof listVariants>

function List({ variant = 'unordered', ...props }: ListProps) {
  const classes = cn(listVariants({ variant }), 'space-y-1')

  if (variant === 'ordered') {
    return <ol data-slot="list" className={classes} {...props} />
  }

  return <ul data-slot="list" className={classes} {...props} />
}

type ListItemProps = Omit<React.LiHTMLAttributes<HTMLLIElement>, 'className'>

function ListItem({ ...props }: ListItemProps) {
  return (
    <li
      data-slot="list-item"
      className={cn('text-sm leading-relaxed')}
      {...props}
    />
  )
}

export { List, ListItem }
