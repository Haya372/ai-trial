import type { Page } from '@playwright/test'
import { expect, type TestUser, test } from '../fixtures'
import { atFixedDay, fixedDayInputValue, fixedDayLabel, fixedDayRange, NEXT_DAY_OFFSET, nextDayAt } from '../support/clock'

const SHARED_EVENT = {
  title: '共有する予定',
  startAt: atFixedDay(nextDayAt(14)),
  endAt: atFixedDay(nextDayAt(15)),
  location: '会議室A',
  description: '共有リンクで見せるメモ',
}

/** 共有元の予定を作成し、その共有リンクのパス（`/share/{token}`）を返す */
async function createSharedEvent(owner: TestUser): Promise<string> {
  const event = await owner.api.createEvent(SHARED_EVENT)
  return owner.api.createShare(event.id)
}

/** 共有ページを開き、予定の詳細が表示されるまで待つ */
async function openSharePage(page: Page, url: string): Promise<void> {
  await page.goto(url)
  await expect(page.getByRole('heading', { name: SHARED_EVENT.title })).toBeVisible()
}

test.describe('予定共有', () => {
  test('予定の詳細から共有リンクを生成すると、未ログインでも共有ページで予定を閲覧できる', async ({
    page,
    anonymousPage,
    user,
    signIn,
  }) => {
    await user.api.createEvent(SHARED_EVENT)

    await signIn(user)

    await page.getByRole('table').getByText(SHARED_EVENT.title).click()
    await page
      .getByRole('dialog', { name: SHARED_EVENT.title })
      .getByRole('button', { name: '共有リンクを生成', exact: true })
      .click()

    const dialog = page.getByRole('dialog', { name: '共有リンクを生成' })
    // 生成前に、メモも共有されることが明示される
    await expect(dialog.getByText('メモ欄の内容もリンクで共有されます', { exact: false })).toBeVisible()
    // 有効期限には予定の終了日時があらかじめ入っている
    await expect(dialog.getByLabel('有効期限')).toHaveValue(fixedDayInputValue(nextDayAt(15)))
    await dialog.getByRole('button', { name: '生成', exact: true }).click()

    const shareUrlInput = dialog.getByRole('textbox')
    await expect(shareUrlInput).toHaveValue(/\/share\/[^/]+$/)
    const shareUrl = await shareUrlInput.inputValue()

    await openSharePage(anonymousPage, shareUrl)
    await expect(anonymousPage.getByText(fixedDayLabel(nextDayAt(14)))).toBeVisible()
    await expect(anonymousPage.getByText(fixedDayLabel(nextDayAt(15)))).toBeVisible()
    await expect(anonymousPage.getByText(SHARED_EVENT.location)).toBeVisible()
    await expect(anonymousPage.getByText(SHARED_EVENT.description)).toBeVisible()
    // 編集・削除の操作は、予定の詳細と同じ描画で出し分けられるため、詳細の表示後に判定できる。
    // 「自分のカレンダーに追加」は認証確認の後に非同期で出し分けられ、未ログインでは描画完了の目印がなく
    // 表示されないことを安定して判定できないため、SharePage の単体テストで確かめる
    for (const name of ['編集', '削除']) {
      await expect(anonymousPage.getByRole('button', { name, exact: true })).toHaveCount(0)
    }
  })

  test('別のユーザーが共有ページから自分のカレンダーに追加すると、閲覧専用で表示される', async ({
    page,
    user: owner,
    createUser,
    signIn,
  }) => {
    const sharePath = await createSharedEvent(owner)
    const viewer = await createUser()

    // 共有ページはログイン画面を経ずに開くため、先にログインしてセッション Cookie を持たせる
    await signIn(viewer)
    await openSharePage(page, sharePath)
    await page.getByRole('button', { name: '自分のカレンダーに追加', exact: true }).click()
    await expect(page.getByText('追加済み', { exact: true })).toBeVisible()

    // page.goto でアプリの認証状態が失われるため、ログインし直してカレンダーを開く
    await signIn(viewer)

    const calendar = page.getByRole('table')
    await calendar.getByText(SHARED_EVENT.title).click()
    const detail = page.getByRole('dialog', { name: SHARED_EVENT.title })
    await expect(detail.getByText('閲覧専用', { exact: true })).toBeVisible()
    await expect(detail.getByText(fixedDayLabel(nextDayAt(14)))).toBeVisible()
    await expect(detail.getByText(SHARED_EVENT.location)).toBeVisible()
    await expect(detail.getByRole('button', { name: 'カレンダーから削除', exact: true })).toBeVisible()
    for (const name of ['編集', '削除', '共有リンクを生成']) {
      await expect(detail.getByRole('button', { name, exact: true })).toHaveCount(0)
    }
  })

  test('追加した予定をカレンダーから削除すると、カレンダーから消え、共有元の予定は残る', async ({
    page,
    user: owner,
    createUser,
    signIn,
  }) => {
    const sharePath = await createSharedEvent(owner)
    const viewer = await createUser()
    await viewer.api.subscribe(sharePath)
    await viewer.api.createEvent({
      title: '残す予定',
      startAt: atFixedDay(nextDayAt(16)),
      endAt: atFixedDay(nextDayAt(17)),
    })

    await signIn(viewer)

    const calendar = page.getByRole('table')
    await calendar.getByText(SHARED_EVENT.title).click()
    const detail = page.getByRole('dialog', { name: SHARED_EVENT.title })
    await detail.getByRole('button', { name: 'カレンダーから削除', exact: true }).click()
    const confirm = page.getByRole('alertdialog', { name: 'この予定をカレンダーから削除しますか？' })
    await confirm.getByRole('button', { name: 'カレンダーから削除', exact: true }).click()

    await expect(confirm).toBeHidden()
    await expect(detail).toBeHidden()
    await expect(calendar.getByText(SHARED_EVENT.title)).toHaveCount(0)
    await expect(calendar.getByText('残す予定')).toBeVisible()

    // 取り消しは共有元に影響しない
    const ownerEvents = await owner.api.listEvents(fixedDayRange(NEXT_DAY_OFFSET))
    expect(ownerEvents.map((e) => e.title)).toEqual([SHARED_EVENT.title])
  })
})
