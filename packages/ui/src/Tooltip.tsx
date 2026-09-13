'use client'

import { Tooltip as TooltipPrimitive } from '@base-ui/react/tooltip'
import { cn } from 'cn'

// Re-export primitive parts
const Tooltip = TooltipPrimitive.Root
const TooltipTrigger = TooltipPrimitive.Trigger

// TooltipContent: Portal + Positioner + Popup
type TooltipContentProps = Omit<
  TooltipPrimitive.Positioner.Props,
  'className'
> & {
  popupProps?: Omit<TooltipPrimitive.Popup.Props, 'className'>
}

function TooltipContent({
  children,
  side,
  sideOffset = 6,
  align,
  alignOffset,
  popupProps,
  ...positionerProps
}: TooltipContentProps) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Positioner
        side={side}
        sideOffset={sideOffset}
        align={align}
        alignOffset={alignOffset}
        {...positionerProps}
      >
        <TooltipPrimitive.Popup
          data-slot="tooltip-content"
          className={cn(
            'rounded-md bg-foreground px-2 py-1 text-xs text-background',
            'data-[ending-style]:opacity-0 data-[starting-style]:opacity-0',
            'transition-opacity duration-150',
          )}
          {...popupProps}
        >
          {children}
        </TooltipPrimitive.Popup>
      </TooltipPrimitive.Positioner>
    </TooltipPrimitive.Portal>
  )
}

export { Tooltip, TooltipTrigger, TooltipContent }
