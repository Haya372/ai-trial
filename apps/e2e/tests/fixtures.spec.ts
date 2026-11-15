import { expect, test } from '../fixtures'
import { FIXED_NOW, atFixedDay } from '../support/clock'

test('ブラウザの現在時刻とタイムゾーンが固定されている', async ({ page }) => {
  await page.goto('/login')

  const now = await page.evaluate(() => ({
    epochMs: Date.now(),
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  }))

  expect(now.timeZone).toBe('Asia/Tokyo')
  // install した時刻から実時間で進むため、数秒の誤差を許容する
  expect(now.epochMs - FIXED_NOW.getTime()).toBeGreaterThanOrEqual(0)
  expect(now.epochMs - FIXED_NOW.getTime()).toBeLessThan(60_000)
})

test('API経由で作成した予定が、ログイン後のカレンダーに表示される', async ({ page, user, signIn }) => {
  await user.api.createEvent({
    title: '固定日の定例',
    startAt: atFixedDay({ hour: 14 }),
    endAt: atFixedDay({ hour: 15 }),
  })

  await signIn(user)

  await expect(page.getByText('固定日の定例')).toBeVisible()
})

test('別のユーザーには、他のユーザーが作成した予定が表示されない', async ({
  page,
  user,
  createUser,
  signIn,
}) => {
  await user.api.createEvent({
    title: '1人目だけの予定',
    startAt: atFixedDay({ hour: 9 }),
    endAt: atFixedDay({ hour: 10 }),
  })
  const other = await createUser()
  await other.api.createEvent({
    title: '2人目の予定',
    startAt: atFixedDay({ hour: 11 }),
    endAt: atFixedDay({ hour: 12 }),
  })

  await signIn(other)

  await expect(page.getByText('2人目の予定')).toBeVisible()
  await expect(page.getByText('1人目だけの予定')).toHaveCount(0)
})
