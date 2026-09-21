type CalendarView = 'month' | 'week'

interface ViewSwitcherProps {
  view: CalendarView
  onViewChange: (view: CalendarView) => void
}

export default function ViewSwitcher({
  view,
  onViewChange,
}: ViewSwitcherProps) {
  return (
    <div role="tablist" className="flex gap-1 rounded-lg bg-muted p-1">
      <button
        role="tab"
        aria-selected={view === 'month'}
        onClick={() => onViewChange('month')}
        className={`rounded-md px-3 py-1 text-sm font-medium transition-all ${
          view === 'month'
            ? 'bg-background text-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground'
        }`}
      >
        月
      </button>
      <button
        role="tab"
        aria-selected={view === 'week'}
        onClick={() => onViewChange('week')}
        className={`rounded-md px-3 py-1 text-sm font-medium transition-all ${
          view === 'week'
            ? 'bg-background text-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground'
        }`}
      >
        週
      </button>
    </div>
  )
}
