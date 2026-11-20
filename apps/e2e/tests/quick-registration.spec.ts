import type { Page } from '@playwright/test'
import { expect, type TestUser, test } from '../fixtures'
import { openWeekViewOfFixedDay } from '../support/calendar'
import {
  fixedDayDateString,
  fixedDayInputValue,
  fixedDayLabel,
  fixedDayRange,
  NEXT_DAY_OFFSET,
  nextDayAt,
} from '../support/clock'

const NEXT_DAY = fixedDayDateString(NEXT_DAY_OFFSET)

/**
 * 週ビューの指定日・時刻の時間枠をクリックする。
 * FullCalendar はクリック座標から日時を判定するため、日付の列と時刻の行が交わる位置をクリックする
 */
async function clickTimeSlot(page: Page, date: string, time: string) {
  // 終日枠の行にも同じ data-date のセルがあるため、時刻の行（data-time）を含む時間枠の行に絞る
  const timedRow = page.getByRole('row').filter({ has: page.locator('[data-time]') })
  const slot = timedRow.locator(`[data-time="${time}"]`).first()
  const dayColumn = timedRow.locator(`[role="cell"][data-date="${date}"]`)
  await slot.scrollIntoViewIfNeeded()
  const [slotBox, columnBox] = await Promise.all([slot.boundingBox(), dayColumn.boundingBox()])
  if (!slotBox || !columnBox) throw new Error(`時間枠が見つからない: ${date} ${time}`)
  // 列内の相対位置で指定し、計測後にスクロールしてもずれず、クリック前の待機も効くようにする
  await dayColumn.click({
    position: { x: columnBox.width / 2, y: slotBox.y - columnBox.y + slotBox.height / 2 },
  })
}

/** FIXED_NOW の翌日に登録されている予定のタイトル */
async function nextDayEventTitles(user: TestUser): Promise<string[]> {
  const events = await user.api.listEvents(fixedDayRange(NEXT_DAY_OFFSET))
  return events.map((e) => e.title)
}

test.describe('クイック登録', () => {
  test('時間枠をクリックしてタイトルを入力しEnterを押すと、1時間の予定が登録される', async ({
    page,
    user,
    signIn,
  }) => {
    await signIn(user)
    await openWeekViewOfFixedDay(page)

    const calendar = page.getByRole('table')

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog', { name: 'クイック登録', exact: true })
    const titleInput = panel.getByLabel('タイトル')
    // タイトル入力欄に自動でフォーカスが当たる
    await expect(titleInput).toBeFocused()
    await titleInput.fill('クイック登録した予定')
    await titleInput.press('Enter')

    await expect(panel.getByText('予定を登録しました')).toBeVisible()
    await expect(calendar.getByText('クイック登録した予定')).toBeVisible()

    await page.keyboard.press('Escape')
    await expect(panel).toBeHidden()

    // クリックした時刻が開始、その1時間後が終了になっていることを詳細モーダルで確かめる
    await calendar.getByText('クイック登録した予定').click()
    const detail = page.getByRole('dialog', { name: 'クイック登録した予定' })
    await expect(detail.getByText(fixedDayLabel(nextDayAt(14)))).toBeVisible()
    await expect(detail.getByText(fixedDayLabel(nextDayAt(15)))).toBeVisible()
  })

  test('保存ボタンで登録した後、「詳細を編集」から編集フォームを開ける', async ({ page, user, signIn }) => {
    await signIn(user)
    await openWeekViewOfFixedDay(page)

    await clickTimeSlot(page, NEXT_DAY, '10:30:00')
    const panel = page.getByRole('dialog', { name: 'クイック登録', exact: true })
    await panel.getByLabel('タイトル').fill('あとで詳細を書く予定')
    await panel.getByRole('button', { name: '保存', exact: true }).click()

    await panel.getByRole('button', { name: '詳細を編集', exact: true }).click()

    const form = page.getByRole('dialog', { name: '予定を編集' })
    await expect(form.getByLabel('タイトル')).toHaveValue('あとで詳細を書く予定')
    await expect(form.getByLabel('開始日時')).toHaveValue(fixedDayInputValue(nextDayAt(10, 30)))
    await expect(form.getByLabel('終了日時')).toHaveValue(fixedDayInputValue(nextDayAt(11, 30)))

    await form.getByLabel('場所').fill('会議室A')
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form).toBeHidden()
    await page.getByRole('table').getByText('あとで詳細を書く予定').click()
    await expect(
      page.getByRole('dialog', { name: 'あとで詳細を書く予定' }).getByText('会議室A'),
    ).toBeVisible()
  })

  test('タイトルが空の場合は、登録されずパネル内にエラーが表示される', async ({ page, user, signIn }) => {
    await signIn(user)
    await openWeekViewOfFixedDay(page)

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog', { name: 'クイック登録', exact: true })
    await panel.getByLabel('タイトル').press('Enter')

    await expect(panel.getByText('タイトルを入力してください')).toBeVisible()
    await expect(panel.getByLabel('タイトル')).toBeVisible()

    // エラーを直せばそのまま登録できる
    await panel.getByLabel('タイトル').fill('入力し直した予定')
    await panel.getByLabel('タイトル').press('Enter')

    await expect(panel.getByText('予定を登録しました')).toBeVisible()
    await expect(page.getByRole('table').getByText('入力し直した予定')).toBeVisible()
    // 空タイトルでの送信が登録されていないことを、サーバー側の件数で確かめる
    expect(await nextDayEventTitles(user)).toEqual(['入力し直した予定'])
  })

  test('保存前にEscで閉じると、予定は登録されない', async ({ page, user, signIn }) => {
    await signIn(user)
    await openWeekViewOfFixedDay(page)

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog', { name: 'クイック登録', exact: true })
    await panel.getByLabel('タイトル').fill('破棄する予定')
    await page.keyboard.press('Escape')

    await expect(panel).toBeHidden()

    // 画面に出ていないことは即座に成立してしまうため、続けて登録を完了させてから
    // サーバー側に破棄した予定が残っていないことを確かめる
    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    await panel.getByLabel('タイトル').fill('残す予定')
    await panel.getByLabel('タイトル').press('Enter')
    await expect(panel.getByText('予定を登録しました')).toBeVisible()

    expect(await nextDayEventTitles(user)).toEqual(['残す予定'])
  })
})
