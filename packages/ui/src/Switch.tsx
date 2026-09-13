import { Switch as SwitchPrimitive } from '@base-ui/react/switch'
import { cn } from 'cn'

type SwitchProps = Omit<SwitchPrimitive.Root.Props, 'className'>

function Switch({ ...props }: SwitchProps) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        'relative inline-flex h-6 w-10 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors outline-none',
        'focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:border-ring',
        'disabled:pointer-events-none disabled:opacity-50',
        'data-[checked]:bg-primary data-[unchecked]:bg-input',
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          'pointer-events-none block size-4 rounded-full bg-background shadow-sm transition-transform',
          'data-[checked]:translate-x-4 data-[unchecked]:translate-x-0',
        )}
      />
    </SwitchPrimitive.Root>
  )
}

export { Switch }
