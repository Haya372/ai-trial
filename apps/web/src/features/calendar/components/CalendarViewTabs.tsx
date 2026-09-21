import { Tabs, TabsList, TabsTrigger } from '@repo/ui'
import { CALENDAR_VIEW, type CalendarView } from '../constants'

interface CalendarViewTabsProps {
  view: CalendarView
  onViewChange: (view: CalendarView) => void
}

export default function CalendarViewTabs({
  view,
  onViewChange,
}: CalendarViewTabsProps) {
  return (
    <Tabs
      value={view}
      onValueChange={(value) => onViewChange(value as CalendarView)}
    >
      <TabsList>
        <TabsTrigger value={CALENDAR_VIEW.MONTH}>月</TabsTrigger>
        <TabsTrigger value={CALENDAR_VIEW.WEEK}>週</TabsTrigger>
      </TabsList>
    </Tabs>
  )
}
