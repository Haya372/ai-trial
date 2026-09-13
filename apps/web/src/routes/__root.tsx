import { Toaster } from '@repo/ui'
import { createRootRoute, Outlet } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: () => (
    <>
      <Outlet />
      <Toaster />
    </>
  ),
})
