'use client'

import { Popover as PopoverPrimitive } from '@base-ui/react/popover'
import { cn } from 'cn'

// Re-export primitive parts
const Popover = PopoverPrimitive.Root

// PopoverContent: Portal + Positioner + Popup
type PopoverContentProps = Omit<
  PopoverPrimitive.Positioner.Props,
  'className'
> & {
  popupProps?: Omit<PopoverPrimitive.Popup.Props, 'className'>
}

function PopoverContent({
  children,
  side,
  sideOffset = 8,
  align,
  alignOffset,
  popupProps,
  ...positionerProps
}: PopoverContentProps) {
  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Positioner
        side={side}
        sideOffset={sideOffset}
        align={align}
        alignOffset={alignOffset}
        {...positionerProps}
      >
        <PopoverPrimitive.Popup
          data-slot="popover-content"
          className={cn(
            'w-72 rounded-md border border-border bg-popover p-4 text-popover-foreground shadow-md',
            'data-[ending-style]:opacity-0 data-[starting-style]:opacity-0',
            'transition-opacity duration-150',
          )}
          {...popupProps}
        >
          {children}
        </PopoverPrimitive.Popup>
      </PopoverPrimitive.Positioner>
    </PopoverPrimitive.Portal>
  )
}

export { Popover, PopoverContent }
