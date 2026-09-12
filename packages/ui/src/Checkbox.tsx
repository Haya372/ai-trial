import { Checkbox as CheckboxPrimitive } from '@base-ui/react/checkbox'
import { cn } from 'cn'

type CheckboxProps = Omit<CheckboxPrimitive.Root.Props, 'className'>

function Checkbox({ ...props }: CheckboxProps) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        'peer inline-flex size-4 shrink-0 cursor-pointer items-center justify-center rounded border border-input bg-background transition-colors outline-none',
        'focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50',
        'disabled:pointer-events-none disabled:opacity-50',
        'data-[checked]:border-primary data-[checked]:bg-primary',
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator
        keepMounted
        className={cn(
          'flex items-center justify-center text-primary-foreground',
          'data-[unchecked]:hidden',
        )}
      >
        <svg viewBox="0 0 14 14" className="size-3" aria-hidden="true">
          <path
            d="M2 7L5.5 10.5L12 4"
            stroke="currentColor"
            strokeWidth="2"
            fill="none"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  )
}

export { Checkbox }
