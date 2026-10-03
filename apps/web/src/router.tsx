import { createRouter } from '@tanstack/react-router'
import ErrorFallback from './components/ErrorFallback'
import { routeTree } from './routeTree.gen'

export const router = createRouter({
  routeTree,
  defaultErrorComponent: ErrorFallback,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
