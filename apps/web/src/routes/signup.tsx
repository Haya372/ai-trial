import { createFileRoute, useNavigate } from '@tanstack/react-router'
import SignupPage from '../features/auth/pages/SignupPage'

export const Route = createFileRoute('/signup')({
  component: SignupRoute,
})

function SignupRoute() {
  const navigate = useNavigate()
  return <SignupPage onSuccess={() => navigate({ to: '/' })} />
}
