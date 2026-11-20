import type { Page } from '@playwright/test'
import { expect, test } from '../fixtures'
import { fixedDayInputValue, fixedDayLabel } from '../support/clock'

// FIXED_NOW（2030-06-12 水曜）の翌日。今日のタイムラインに出ない日にして、カレンダー本体だけに予定が表示されるようにする
const NEXT_DAY = '2030-06-13'
const nextDayAt = (hour: number, minute = 0) => ({ dayOffset: 1, hour, minute })

/** 週ビューを開き、FIXED_NOW を含む週を表示する */
async function openWeekView(page: Page) {
  await page.getByRole('tab', { name: '週', exact: true }).click()
  // 切り替え直後は月初を含む週が表示される（#398）ため、今日ボタンで固定日の週へ移る
  await page.getByRole('button', { name: '今日', exact: true }).click()
  await expect(page.getByText('2030年6月9日〜15日', { exact: true })).toBeVisible()
}

/**
 * 週ビューの指定日・時刻の時間枠をクリックする。
 * FullCalendar はクリック座標から日時を判定するため、日付の列と時刻の行が交わる位置をクリックする
 */
async function clickTimeSlot(page: Page, date: string, time: string) {
  const slot = page.locator(`[data-time="${time}"]`).first()
  await slot.scrollIntoViewIfNeeded()
  const slotBox = await slot.boundingBox()
  // 終日枠と時間枠の両方に同じ data-date のセルがあるが、列の横位置はどちらも同じ
  const dayBox = await page.locator(`[role="cell"][data-date="${date}"]`).first().boundingBox()
  if (!slotBox || !dayBox) throw new Error(`時間枠が見つからない: ${date} ${time}`)
  await page.mouse.click(dayBox.x + dayBox.width / 2, slotBox.y + slotBox.height / 2)
}

test.describe('クイック登録', () => {
  test('時間枠をクリックしてタイトルを入力しEnterを押すと、1時間の予定が登録される', async ({
    page,
    user,
    signIn,
  }) => {
    await signIn(user)
    await openWeekView(page)

    const calendar = page.getByRole('table')

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog')
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
    await openWeekView(page)

    await clickTimeSlot(page, NEXT_DAY, '10:30:00')
    const panel = page.getByRole('dialog')
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
    await openWeekView(page)

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog')
    await panel.getByLabel('タイトル').press('Enter')

    await expect(panel.getByText('タイトルを入力してください')).toBeVisible()
    await expect(panel.getByLabel('タイトル')).toBeVisible()

    // エラーを直せばそのまま登録できる
    await panel.getByLabel('タイトル').fill('入力し直した予定')
    await panel.getByLabel('タイトル').press('Enter')

    await expect(panel.getByText('予定を登録しました')).toBeVisible()
    await expect(page.getByRole('table').getByText('入力し直した予定')).toBeVisible()
  })

  test('保存前にEscで閉じると、予定は登録されない', async ({ page, user, signIn }) => {
    await signIn(user)
    await openWeekView(page)

    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    const panel = page.getByRole('dialog')
    await panel.getByLabel('タイトル').fill('破棄する予定')
    await page.keyboard.press('Escape')

    await expect(panel).toBeHidden()
    await expect(page.getByText('破棄する予定')).toHaveCount(0)

    // 開き直したパネルに前回の入力が残っていない
    await clickTimeSlot(page, NEXT_DAY, '14:00:00')
    await expect(panel.getByLabel('タイトル')).toHaveValue('')
  })
})
