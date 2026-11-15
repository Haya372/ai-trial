import { expect, test } from '../fixtures'
import { atFixedDay, fixedDayInputValue, fixedDayLabel } from '../support/clock'

// FIXED_NOW の翌日。今日のタイムラインに出ない日にして、カレンダー本体だけに予定が表示されるようにする
const nextDayAt = (hour: number, minute = 0) => ({ dayOffset: 1, hour, minute })

test.describe('予定管理', () => {
  test('フォームから予定を作成すると、カレンダーに表示される', async ({ page, user, signIn }) => {
    await signIn(user)

    const calendar = page.getByRole('table')

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('E2Eで作成した予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(14)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))
    await form.getByLabel('場所').fill('会議室A')
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form).toBeHidden()
    await expect(calendar.getByText('E2Eで作成した予定')).toBeVisible()

    // カレンダーには日時・場所が出ないため、入力値どおりに登録されたかは詳細モーダルで確かめる
    await calendar.getByText('E2Eで作成した予定').click()
    const detail = page.getByRole('dialog', { name: 'E2Eで作成した予定' })
    await expect(detail.getByText(fixedDayLabel(nextDayAt(14)))).toBeVisible()
    await expect(detail.getByText(fixedDayLabel(nextDayAt(15)))).toBeVisible()
    await expect(detail.getByText('会議室A')).toBeVisible()
  })

  test('作成フォームをキャンセルすると、予定は登録されない', async ({ page, user, signIn }) => {
    await signIn(user)

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('キャンセルした予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(14)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))
    await form.getByRole('button', { name: 'キャンセル', exact: true }).click()

    await expect(form).toBeHidden()
    await expect(page.getByRole('table')).toBeVisible()
    await expect(page.getByText('キャンセルした予定')).toHaveCount(0)

    // 開き直したフォームに前回の入力が残っていない
    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    await expect(form.getByLabel('タイトル')).toHaveValue('')
  })

  test('タイトルが空の場合は、保存されずエラーが表示される', async ({ page, user, signIn }) => {
    await signIn(user)

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(14)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form.getByText('タイトルを入力してください')).toBeVisible()
    await expect(form).toBeVisible()
  })

  test('終了日時が開始日時以前の場合は、保存されずエラーが表示される', async ({ page, user, signIn }) => {
    await signIn(user)

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('日時を直す予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(15)))
    // 同時刻も不可という境界値で確かめる
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form.getByText('終了日時は開始日時より後に設定してください')).toBeVisible()
    await expect(form).toBeVisible()

    // エラーを解消すれば保存できる
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(16)))
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form).toBeHidden()
    await expect(page.getByRole('table').getByText('日時を直す予定')).toBeVisible()
  })

  test('詳細モーダルから予定を編集すると、変更が反映される', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '編集前の予定',
      startAt: atFixedDay(nextDayAt(14)),
      endAt: atFixedDay(nextDayAt(15)),
      location: '会議室A',
    })

    await signIn(user)

    const calendar = page.getByRole('table')

    await calendar.getByText('編集前の予定').click()
    await page
      .getByRole('dialog', { name: '編集前の予定' })
      .getByRole('button', { name: '編集', exact: true })
      .click()

    const form = page.getByRole('dialog', { name: '予定を編集' })
    await expect(form.getByLabel('タイトル')).toHaveValue('編集前の予定')
    await expect(form.getByLabel('開始日時')).toHaveValue(fixedDayInputValue(nextDayAt(14)))
    await expect(form.getByLabel('終了日時')).toHaveValue(fixedDayInputValue(nextDayAt(15)))
    await expect(form.getByLabel('場所')).toHaveValue('会議室A')

    await form.getByLabel('タイトル').fill('編集後の予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(16)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(17, 30)))
    await form.getByLabel('場所').fill('会議室B')
    await form.getByRole('button', { name: '保存', exact: true }).click()

    await expect(form).toBeHidden()
    await expect(calendar.getByText('編集後の予定')).toBeVisible()
    await expect(calendar.getByText('編集前の予定')).toHaveCount(0)

    await calendar.getByText('編集後の予定').click()
    const detail = page.getByRole('dialog', { name: '編集後の予定' })
    await expect(detail.getByText(fixedDayLabel(nextDayAt(16)))).toBeVisible()
    await expect(detail.getByText(fixedDayLabel(nextDayAt(17, 30)))).toBeVisible()
    await expect(detail.getByText('会議室B')).toBeVisible()
  })

  test('予定を削除すると、カレンダーから消える', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '削除する予定',
      startAt: atFixedDay(nextDayAt(14)),
      endAt: atFixedDay(nextDayAt(15)),
    })
    await user.api.createEvent({
      title: '残す予定',
      startAt: atFixedDay(nextDayAt(16)),
      endAt: atFixedDay(nextDayAt(17)),
    })

    await signIn(user)

    const calendar = page.getByRole('table')

    await calendar.getByText('削除する予定').click()
    const detail = page.getByRole('dialog', { name: '削除する予定' })
    await detail.getByRole('button', { name: '削除', exact: true }).click()
    const confirm = page.getByRole('alertdialog', { name: 'この予定を削除しますか？' })

    // キャンセルすると確認ダイアログだけが閉じ、詳細モーダルに戻る
    await confirm.getByRole('button', { name: 'キャンセル', exact: true }).click()
    await expect(confirm).toBeHidden()
    await expect(detail).toBeVisible()

    await detail.getByRole('button', { name: '削除', exact: true }).click()
    await confirm.getByRole('button', { name: '削除', exact: true }).click()

    await expect(confirm).toBeHidden()
    await expect(detail).toBeHidden()
    await expect(calendar.getByText('削除する予定')).toHaveCount(0)
    await expect(calendar.getByText('残す予定')).toBeVisible()
  })
})
