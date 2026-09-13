import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { useAuthStore } from '../stores/auth'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: () => {
    const user = useAuthStore.getState().user
    if (!user) {
      throw redirect({ to: '/login' })
    }
  },
  component: () => <Outlet />,
})
