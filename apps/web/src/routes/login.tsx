import { createFileRoute, useNavigate } from '@tanstack/react-router'
import LoginPage from '../pages/LoginPage'

export const Route = createFileRoute('/login')({
  component: LoginRoute,
})

function LoginRoute() {
  const navigate = useNavigate()
  return <LoginPage onSuccess={() => navigate({ to: '/' })} />
}
