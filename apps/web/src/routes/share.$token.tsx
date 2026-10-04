import { createFileRoute } from '@tanstack/react-router'
import SharePage from '../features/share/pages/SharePage'

export const Route = createFileRoute('/share/$token')({
  component: ShareRoute,
})

function ShareRoute() {
  const { token } = Route.useParams()
  return <SharePage token={token} />
}
