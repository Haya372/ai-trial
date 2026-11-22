import { randomUUID } from 'node:crypto'
import {
  test as base,
  type APIRequestContext,
  expect,
  type Page,
} from '@playwright/test'
import { expectStatus, UserApi } from './support/api'
import { FIXED_NOW } from './support/clock'
import { getSignupUrl, type SignupRequest } from './support/generated/api'

export type TestUser = {
  email: string
  password: string
  displayName: string
  /** このユーザーとしてログイン済みの API クライアント */
  api: UserApi
}

type Fixtures = {
  /** テストごとに一意なメールアドレスでサインアップ済みのユーザー */
  user: TestUser
  /** 共有シナリオなどで使う、追加のユーザーをサインアップする */
  createUser: () => Promise<TestUser>
  /**
   * `page` でログイン画面から指定ユーザーとしてログインし、カレンダー画面を開く。
   * 認証状態はメモリにしか保持されず `page.goto` で失われるため、
   * テストデータを API で用意し終えてから呼ぶ
   */
  signIn: (user: TestUser) => Promise<void>
  /** `page` とは Cookie を共有しない、未ログイン状態のページ */
  anonymousPage: Page
}

const PASSWORD = 'Passw0rd!1234'

export const test = base.extend<Fixtures>({
  page: async ({ page }, use) => {
    // ページ読み込み前に install し、アプリの初期表示日も固定時刻にする
    await page.clock.install({ time: FIXED_NOW })
    await use(page)
  },

  createUser: async ({ playwright, baseURL }, use) => {
    const contexts: APIRequestContext[] = []

    await use(async () => {
      const request = await playwright.request.newContext({ baseURL })
      contexts.push(request)

      // 並列実行・再実行でも衝突しないよう、メールアドレスを UUID で一意にする
      const email = `e2e-${randomUUID()}@example.com`
      const displayName = `E2E User ${contexts.length}`
      const data: SignupRequest = { email, password: PASSWORD, displayName }
      const res = await request.post(getSignupUrl(), { data })
      await expectStatus(res, 201)

      return {
        email,
        password: PASSWORD,
        displayName,
        api: new UserApi(request),
      }
    })

    await Promise.all(contexts.map((c) => c.dispose()))
  },

  user: async ({ createUser }, use) => {
    await use(await createUser())
  },

  signIn: async ({ page }, use) => {
    await use(async (user) => {
      await page.goto('/login')
      await page.getByLabel('メールアドレス').fill(user.email)
      await page.getByLabel('パスワード').fill(user.password)
      await page.getByRole('button', { name: 'ログイン' }).click()
      await expect(page).toHaveURL('/')
    })
  },

  anonymousPage: async ({ browser }, use) => {
    const context = await browser.newContext()
    const page = await context.newPage()
    await page.clock.install({ time: FIXED_NOW })
    await use(page)
    await context.close()
  },
})

export { expect }
