import { expect, type Page } from '@playwright/test'

// FIXED_NOW（2030-06-12 水曜）を含む週の期間表示。週は日曜始まり
export const WEEK_OF_FIXED_DAY = '2030年6月9日〜15日'

/** 週ビューに切り替え、FIXED_NOW を含む週を表示する */
export async function openWeekViewOfFixedDay(page: Page): Promise<void> {
  await page.getByRole('tab', { name: '週', exact: true }).click()
  // 切り替え直後は月初を含む週が表示される（#398）ため、今日ボタンで固定日の週へ移る。
  // #398 を直したら今日ボタンの操作を外し、切り替えだけで固定日の週になることを確かめる
  await page.getByRole('button', { name: '今日', exact: true }).click()
  await expect(page.getByText(WEEK_OF_FIXED_DAY, { exact: true })).toBeVisible()
}
