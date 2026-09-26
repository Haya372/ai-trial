import { Tabs as TabsPrimitive } from '@base-ui/react/tabs'
import { cn } from 'cn'

// ─── Tabs (Root) ──────────────────────────────────────────────────────────────

type TabsProps = Omit<TabsPrimitive.Root.Props, 'className'>

function Tabs({ ...props }: TabsProps) {
  return (
    <TabsPrimitive.Root
      data-slot="tabs"
      className={cn('flex flex-col')}
      {...props}
    />
  )
}

// ─── TabsList ─────────────────────────────────────────────────────────────────

type TabsListProps = Omit<TabsPrimitive.List.Props, 'className'>

function TabsList({ ...props }: TabsListProps) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn(
        'inline-flex items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground',
      )}
      {...props}
    />
  )
}

// ─── TabsTrigger ──────────────────────────────────────────────────────────────

type TabsTriggerProps = Omit<TabsPrimitive.Tab.Props, 'className'>

function TabsTrigger({ ...props }: TabsTriggerProps) {
  return (
    <TabsPrimitive.Tab
      data-slot="tabs-trigger"
      className={cn(
        'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium transition-all',
        'text-muted-foreground outline-none',
        'disabled:pointer-events-none disabled:opacity-50',
        'focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1',
        'data-[active]:bg-background data-[active]:text-foreground data-[active]:shadow-sm',
      )}
      {...props}
    />
  )
}

// ─── TabsContent (Panel) ──────────────────────────────────────────────────────

type TabsContentProps = Omit<TabsPrimitive.Panel.Props, 'className'>

function TabsContent({ ...props }: TabsContentProps) {
  return (
    <TabsPrimitive.Panel
      data-slot="tabs-content"
      className={cn('mt-2 text-foreground outline-none')}
      {...props}
    />
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent }
