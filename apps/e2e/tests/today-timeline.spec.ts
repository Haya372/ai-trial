import type { Page } from '@playwright/test'
import { expect, test } from '../fixtures'
import { atFixedDay, type FixedDayTime } from '../support/clock'

/** 今日のタイムラインパネル（section の aria-label で region になる） */
const timelineOf = (page: Page) => page.getByRole('region', { name: '今日のタイムライン' })

/**
 * FIXED_NOW の日付の指定時刻まで時計を進めて止める。
 * 時計は install 後も実時間で進むため、現在時刻に依存する表示を確かめる前に止めて結果を固定する。
 * 途中のタイマーも発火するので、分の境界ごとの再計算もこれで走る
 */
async function pauseAtFixedDay(page: Page, time: FixedDayTime): Promise<void> {
  await page.clock.pauseAt(new Date(atFixedDay(time)))
}

test.describe('今日のタイムライン', () => {
  test('今日の予定だけが時系列で並び、次の予定と空き時間が表示される', async ({ page, user, signIn }) => {
    // 並び順を確かめるため、開始日時の順とは異なる順で作成する
    await user.api.createEvent({
      title: '設計レビュー',
      startAt: atFixedDay({ hour: 10, minute: 30 }),
      endAt: atFixedDay({ hour: 12 }),
    })
    await user.api.createEvent({
      title: '朝会',
      startAt: atFixedDay({ hour: 9 }),
      endAt: atFixedDay({ hour: 10, minute: 15 }),
    })
    await user.api.createEvent({
      title: '前日からの作業',
      startAt: atFixedDay({ dayOffset: -1, hour: 22 }),
      endAt: atFixedDay({ hour: 8 }),
    })
    await user.api.createEvent({
      title: '前日の予定',
      startAt: atFixedDay({ dayOffset: -1, hour: 14 }),
      endAt: atFixedDay({ dayOffset: -1, hour: 15 }),
    })
    await user.api.createEvent({
      title: '翌日の予定',
      startAt: atFixedDay({ dayOffset: 1, hour: 11 }),
      endAt: atFixedDay({ dayOffset: 1, hour: 12 }),
    })

    await signIn(user)
    await pauseAtFixedDay(page, { hour: 10, minute: 5 })

    const timeline = timelineOf(page)
    const items = [
      '前日から〜08:00前日からの作業',
      '60分の空き',
      '09:00〜10:15朝会',
      // 開催中の朝会の後、まだ始まっていない設計レビューの前に現在時刻を置く
      '10:05',
      '15分の空き',
      '10:30〜12:00設計レビュー',
    ]

    await expect(timeline.getByText('設計レビュー', { exact: true }).first()).toBeVisible()
    await expect(timeline.getByText('あと25分', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('listitem')).toHaveText(items)

    // カレンダーの表示期間を移動しても、パネルは今日の予定を表示し続ける
    await page.getByRole('button', { name: '次へ', exact: true }).click()
    await expect(page.getByText('2030年7月', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('listitem')).toHaveText(items)
  })

  test('時刻が進むと、次の予定までの時間と現在時刻の位置が更新される', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '朝会',
      startAt: atFixedDay({ hour: 9 }),
      endAt: atFixedDay({ hour: 10, minute: 15 }),
    })
    await user.api.createEvent({
      title: '設計レビュー',
      startAt: atFixedDay({ hour: 10, minute: 30 }),
      endAt: atFixedDay({ hour: 11 }),
    })

    await signIn(user)

    const timeline = timelineOf(page)

    await pauseAtFixedDay(page, { hour: 10, minute: 5 })
    await expect(timeline.getByText('あと25分', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('listitem')).toHaveText([
      '09:00〜10:15朝会',
      '10:05',
      '15分の空き',
      '10:30〜11:00設計レビュー',
    ])

    // 朝会が終わり、空き時間に入ると現在時刻は空き時間の後に移る
    await pauseAtFixedDay(page, { hour: 10, minute: 20 })
    await expect(timeline.getByText('あと10分', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('listitem')).toHaveText([
      '09:00〜10:15朝会',
      '15分の空き',
      '10:20',
      '10:30〜11:00設計レビュー',
    ])

    // 設計レビューが始まると開催中の予定になり、今日の残りの予定はなくなる
    await pauseAtFixedDay(page, { hour: 10, minute: 30 })
    await expect(timeline.getByText('今日の予定はもうありません', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('listitem')).toHaveText([
      '09:00〜10:15朝会',
      '15分の空き',
      '10:30〜11:00設計レビュー',
      '10:30',
    ])
  })

  test('パネルの予定から詳細を開いて削除すると、パネルからも消える', async ({ page, user, signIn }) => {
    await user.api.createEvent({
      title: '設計レビュー',
      startAt: atFixedDay({ hour: 14 }),
      endAt: atFixedDay({ hour: 15 }),
      location: '会議室A',
    })

    await signIn(user)

    const timeline = timelineOf(page)

    await timeline.getByRole('button', { name: /設計レビュー/ }).click()
    const detail = page.getByRole('dialog', { name: '設計レビュー' })
    await expect(detail.getByText('会議室A')).toBeVisible()

    await detail.getByRole('button', { name: '削除', exact: true }).click()
    const confirm = page.getByRole('alertdialog', { name: 'この予定を削除しますか？' })
    await confirm.getByRole('button', { name: '削除', exact: true }).click()

    await expect(detail).toBeHidden()
    await expect(timeline.getByText('今日の予定はありません', { exact: true })).toBeVisible()
    await expect(timeline.getByRole('button')).toHaveCount(0)
    await expect(page.getByRole('table').getByText('設計レビュー')).toHaveCount(0)
  })
})
