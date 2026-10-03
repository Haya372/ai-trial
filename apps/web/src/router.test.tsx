import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ErrorFallback from './components/ErrorFallback'
import { router } from './router'

describe('router', () => {
  beforeEach(() => {
    // jsdom は scrollTo を実装していないため、RouterProvider のスクロール復元処理が
    // 警告を出す。テスト出力をpristineに保つためスタブ化する。
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
  })

  it('defaultErrorComponent に ErrorFallback が設定されている', () => {
    expect(router.options.defaultErrorComponent).toBe(ErrorFallback)
  })

  it('ルートコンポーネントが例外を投げても白画面にならずフォールバックUIを表示する', async () => {
    const rootRoute = createRootRoute()
    const throwingRoute = createRoute({
      getParentRoute: () => rootRoute,
      path: '/',
      component: () => {
        throw new Error('boom')
      },
    })
    const testRouter = createRouter({
      routeTree: rootRoute.addChildren([throwingRoute]),
      defaultErrorComponent: ErrorFallback,
      history: createMemoryHistory({ initialEntries: ['/'] }),
    })

    render(<RouterProvider router={testRouter} />)

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
