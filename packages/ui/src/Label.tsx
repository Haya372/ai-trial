import * as React from 'react'
import { cn } from 'cn'

type LabelProps = Omit<React.LabelHTMLAttributes<HTMLLabelElement>, 'className'>

function Label(props: LabelProps) {
  return (
    <label
      data-slot="label"
      className={cn(
        'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
      )}
      {...props}
    />
  )
}

export { Label }
