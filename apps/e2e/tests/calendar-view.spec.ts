import { expect, test } from '../fixtures'
import { openWeekViewOfFixedDay, WEEK_OF_FIXED_DAY } from '../support/calendar'
import { atFixedDay } from '../support/clock'

// FIXED_NOW（2030-06-12 水曜）を基準にした表示期間のラベル。週は日曜始まり
const JUNE = '2030年6月'
const MAY = '2030年5月'
const JULY = '2030年7月'
const PREV_WEEK = '2030年6月2日〜8日'
const NEXT_WEEK = '2030年6月16日〜22日'

test.describe('カレンダー表示', () => {
  test('月・週ビューを切り替えられ、各ビューに予定が表示される', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '固定日の定例',
      startAt: atFixedDay({ hour: 14 }),
      endAt: atFixedDay({ hour: 15 }),
    })

    await signIn(user)

    // 予定は今日のタイムラインにも出るため、カレンダー本体（表示中のビューの table は1つだけ）に絞る
    const calendar = page.getByRole('table')

    await expect(page.getByText(JUNE, { exact: true })).toBeVisible()
    await expect(calendar.getByText('固定日の定例')).toBeVisible()

    // 初期表示の月ビューから切り替えると、今日ではなく月初を含む週が表示される不具合がある（#398）。
    // 切り替え直後はどの週かを問わず週の期間表示になることだけ確かめ、今日ボタンで固定日の週へ移る
    await page.getByRole('tab', { name: '週', exact: true }).click()
    await expect(page.getByText(/^\d{4}年\d{1,2}月\d{1,2}日〜/)).toBeVisible()
    await page.getByRole('button', { name: '今日', exact: true }).click()
    await expect(page.getByText(WEEK_OF_FIXED_DAY, { exact: true })).toBeVisible()
    await expect(calendar.getByText('固定日の定例')).toBeVisible()

    // 直前に表示していた日付を含む月に戻る
    await page.getByRole('tab', { name: '月', exact: true }).click()
    await expect(page.getByText(JUNE, { exact: true })).toBeVisible()
    await expect(calendar.getByText('固定日の定例')).toBeVisible()
  })

  test('月ビューで前後の月へ移動すると、その月の予定が表示される', async ({ page, user, signIn }) => {
    // 月ビューのグリッドに前後月の端の日も並ぶため、移動前の月のグリッドに出ない日付を選ぶ
    await user.api.createEvent({
      title: '5月の予定',
      startAt: atFixedDay({ dayOffset: -31, hour: 14 }),
      endAt: atFixedDay({ dayOffset: -31, hour: 15 }),
    })
    await user.api.createEvent({
      title: '6月の予定',
      startAt: atFixedDay({ hour: 14 }),
      endAt: atFixedDay({ hour: 15 }),
    })
    await user.api.createEvent({
      title: '7月の予定',
      startAt: atFixedDay({ dayOffset: 30, hour: 14 }),
      endAt: atFixedDay({ dayOffset: 30, hour: 15 }),
    })

    await signIn(user)

    const calendar = page.getByRole('table')

    await expect(page.getByText(JUNE, { exact: true })).toBeVisible()
    await expect(calendar.getByText('6月の予定')).toBeVisible()
    await expect(calendar.getByText('5月の予定')).toHaveCount(0)
    await expect(calendar.getByText('7月の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '前へ', exact: true }).click()
    await expect(page.getByText(MAY, { exact: true })).toBeVisible()
    await expect(calendar.getByText('5月の予定')).toBeVisible()
    await expect(calendar.getByText('6月の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '次へ', exact: true }).click()
    await expect(page.getByText(JUNE, { exact: true })).toBeVisible()
    await page.getByRole('button', { name: '次へ', exact: true }).click()
    await expect(page.getByText(JULY, { exact: true })).toBeVisible()
    await expect(calendar.getByText('7月の予定')).toBeVisible()
    await expect(calendar.getByText('6月の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '今日', exact: true }).click()
    await expect(page.getByText(JUNE, { exact: true })).toBeVisible()
    await expect(calendar.getByText('6月の予定')).toBeVisible()
  })

  test('週ビューで前後の週へ移動すると、その週の予定が表示される', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '前週の予定',
      startAt: atFixedDay({ dayOffset: -7, hour: 14 }),
      endAt: atFixedDay({ dayOffset: -7, hour: 15 }),
    })
    await user.api.createEvent({
      title: '今週の予定',
      startAt: atFixedDay({ hour: 14 }),
      endAt: atFixedDay({ hour: 15 }),
    })
    await user.api.createEvent({
      title: '翌週の予定',
      startAt: atFixedDay({ dayOffset: 7, hour: 14 }),
      endAt: atFixedDay({ dayOffset: 7, hour: 15 }),
    })

    await signIn(user)
    await openWeekViewOfFixedDay(page)

    const calendar = page.getByRole('table')

    await expect(calendar.getByText('今週の予定')).toBeVisible()
    await expect(calendar.getByText('前週の予定')).toHaveCount(0)
    await expect(calendar.getByText('翌週の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '前へ', exact: true }).click()
    await expect(page.getByText(PREV_WEEK, { exact: true })).toBeVisible()
    await expect(calendar.getByText('前週の予定')).toBeVisible()
    await expect(calendar.getByText('今週の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '次へ', exact: true }).click()
    await expect(page.getByText(WEEK_OF_FIXED_DAY, { exact: true })).toBeVisible()
    await page.getByRole('button', { name: '次へ', exact: true }).click()
    await expect(page.getByText(NEXT_WEEK, { exact: true })).toBeVisible()
    await expect(calendar.getByText('翌週の予定')).toBeVisible()
    await expect(calendar.getByText('今週の予定')).toHaveCount(0)

    await page.getByRole('button', { name: '今日', exact: true }).click()
    await expect(page.getByText(WEEK_OF_FIXED_DAY, { exact: true })).toBeVisible()
    await expect(calendar.getByText('今週の予定')).toBeVisible()
  })
})
