import { expect, test } from '@playwright/test'

test('サインアップ後、同じ認証情報でログインできる', async ({ page }) => {
  const email = `e2e-${Date.now()}@example.com`
  const password = 'password1234'

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
})
