import { expect, test } from '@playwright/test'

test('サインアップ後、同じ認証情報でログインし、ログアウトできる', async ({ page }) => {
  const email = `e2e-${Date.now()}@example.com`
  const password = 'Passw0rd!1234'

  await page.goto('/signup')
  await page.getByLabel('メールアドレス').fill(email)
  await page.getByLabel('パスワード').fill(password)
  await page.getByLabel('表示名').fill('E2E User')
  await page.getByRole('button', { name: '登録' }).click()
  await expect(page).toHaveURL('/')

  // 再読み込みで認証ストア（メモリ）を破棄し、ログイン画面から入り直す
  await page.goto('/login')
  await page.getByLabel('メールアドレス').fill(email)
  await page.getByLabel('パスワード').fill(password)
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL('/')

  await page.getByRole('button', { name: 'ログアウト' }).click()
  await expect(page).toHaveURL('/login')

  // サーバー側のセッションが破棄されている（リロードによるストア消失とは独立に検証する）
  const me = await page.request.get('/api/auth/me')
  expect(me.status()).toBe(401)

  // 認証必須ページには戻れない
  await page.goto('/')
  await expect(page).toHaveURL('/login')
})
