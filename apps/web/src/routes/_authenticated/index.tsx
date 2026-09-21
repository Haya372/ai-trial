import { createFileRoute } from '@tanstack/react-router'
import CalendarPage from '../../features/calendar/pages/CalendarPage'

export const Route = createFileRoute('/_authenticated/')({
  component: CalendarPage,
})
