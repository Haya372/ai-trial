import { Tabs, TabsList, TabsTrigger } from '@repo/ui'
import { useTranslation } from 'react-i18next'
import { CALENDAR_VIEW, type CalendarView } from '../constants'

interface CalendarViewTabsProps {
  view: CalendarView
  onViewChange: (view: CalendarView) => void
}

export default function CalendarViewTabs({
  view,
  onViewChange,
}: CalendarViewTabsProps) {
  const { t } = useTranslation('calendar')

  return (
    <Tabs
      value={view}
      onValueChange={(value) => onViewChange(value as CalendarView)}
    >
      <TabsList>
        <TabsTrigger value={CALENDAR_VIEW.MONTH}>
          {t('viewTabs.month')}
        </TabsTrigger>
        <TabsTrigger value={CALENDAR_VIEW.WEEK}>
          {t('viewTabs.week')}
        </TabsTrigger>
      </TabsList>
    </Tabs>
  )
}
