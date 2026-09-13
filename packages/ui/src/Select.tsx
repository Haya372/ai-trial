import { Select as SelectPrimitive } from '@base-ui/react/select'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from 'cn'

// ─── Trigger ─────────────────────────────────────────────────────────────────

const selectTriggerVariants = cva(
  'flex w-full items-center justify-between rounded-lg border border-input bg-background text-foreground transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 data-[placeholder]:text-muted-foreground',
  {
    variants: {
      size: {
        sm: 'h-7 px-2.5 text-[0.8rem]',
        md: 'h-9 px-3 text-sm',
        lg: 'h-10 px-3.5 text-base',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  },
)

type SelectTriggerProps = Omit<SelectPrimitive.Trigger.Props, 'className'> &
  VariantProps<typeof selectTriggerVariants>

function SelectTrigger({ size, children, ...props }: SelectTriggerProps) {
  return (
    <SelectPrimitive.Trigger
      data-slot="select-trigger"
      className={selectTriggerVariants({ size })}
      {...props}
    >
      {children}
      <SelectPrimitive.Icon className="ml-auto shrink-0 opacity-50">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </SelectPrimitive.Icon>
    </SelectPrimitive.Trigger>
  )
}

// ─── Content ──────────────────────────────────────────────────────────────────

type SelectContentProps = Omit<SelectPrimitive.Positioner.Props, 'className'>

function SelectContent({ children, ...props }: SelectContentProps) {
  return (
    <SelectPrimitive.Portal>
      <SelectPrimitive.Positioner
        data-slot="select-positioner"
        className="z-50"
        {...props}
      >
        <SelectPrimitive.Popup
          data-slot="select-popup"
          className={cn(
            'min-w-[var(--trigger-width)] rounded-lg border border-border bg-background shadow-md outline-none',
            'data-[ending-style]:animate-out data-[ending-style]:fade-out-0 data-[ending-style]:zoom-out-95',
            'data-[starting-style]:animate-in data-[starting-style]:fade-in-0 data-[starting-style]:zoom-in-95',
            'py-1',
          )}
        >
          {children}
        </SelectPrimitive.Popup>
      </SelectPrimitive.Positioner>
    </SelectPrimitive.Portal>
  )
}

// ─── Item ─────────────────────────────────────────────────────────────────────

type SelectItemProps = Omit<SelectPrimitive.Item.Props, 'className'>

function SelectItem({ children, ...props }: SelectItemProps) {
  return (
    <SelectPrimitive.Item
      data-slot="select-item"
      className={cn(
        'relative flex cursor-default select-none items-center gap-2 px-3 py-1.5 text-sm outline-none',
        'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
        'data-[highlighted]:bg-muted data-[highlighted]:text-foreground',
        'pr-8',
      )}
      {...props}
    >
      <span className="absolute right-3 flex size-4 items-center justify-center">
        <SelectPrimitive.ItemIndicator>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </SelectPrimitive.ItemIndicator>
      </span>
      {children}
    </SelectPrimitive.Item>
  )
}

// ─── Re-exports ───────────────────────────────────────────────────────────────

const Select = SelectPrimitive.Root
const SelectValue = SelectPrimitive.Value
const SelectItemText = SelectPrimitive.ItemText

export {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
  SelectItemText,
}
