import { createFileRoute } from '@tanstack/react-router'
import { useAuthStore } from '../../stores/auth'

export const Route = createFileRoute('/_authenticated/')({
  component: CalendarPlaceholder,
})

function CalendarPlaceholder() {
  const user = useAuthStore((s) => s.user)
  return (
    <div className="p-8">
      <h1>Calendar</h1>
      <p>Logged in as {user?.email}</p>
    </div>
  )
}
