import type { Locator, Page } from '@playwright/test'
import { expect, test } from '../fixtures'
import {
  atFixedDay,
  fixedDayInputValue,
  fixedDayRange,
  NEXT_DAY_OFFSET,
  nextDayAt,
} from '../support/clock'

const EVENT_TIME = {
  startAt: atFixedDay(nextDayAt(14)),
  endAt: atFixedDay(nextDayAt(15)),
}

/** カレンダーヘッダーの「タグ管理」からタグ管理モーダルを開く */
async function openTagManagement(page: Page): Promise<Locator> {
  await page.getByRole('button', { name: 'タグ管理', exact: true }).click()
  const modal = page.getByRole('dialog', { name: 'タグ管理' })
  await expect(modal).toBeVisible()
  return modal
}

/** 予定フォームのタグ選択欄。選択中のタグは付与順にバッジで並ぶ */
function tagSelector(form: Locator): Locator {
  return form.getByRole('group', { name: 'タグ' })
}

/** タグ選択欄のポップオーバー。フォームのダイアログの外にポータルで描画される */
function tagPopover(page: Page): Locator {
  return page.locator('[data-slot="popover-content"]')
}

function tagBadges(container: Locator): Locator {
  return container.locator('[data-slot="tag-badge"]')
}

test.describe('タグ管理', () => {
  test('新しいタグを作成すると、タグ一覧に追加される', async ({
    page,
    user,
    signIn,
  }) => {
    await signIn(user)

    const modal = await openTagManagement(page)
    await expect(modal.getByText('タグはまだありません')).toBeVisible()

    await modal.getByRole('button', { name: '新しいタグ', exact: true }).click()
    await modal.getByLabel('タグ名').fill('仕事')
    await modal.getByRole('radio', { name: '青', exact: true }).click()
    await modal.getByRole('button', { name: '保存', exact: true }).click()

    await expect(modal.getByLabel('タグ名')).toBeHidden()
    await expect(tagBadges(modal)).toHaveText(['仕事'])

    // 一覧のバッジからは色キーを読み取れないため、選んだ色で登録されたかは API で確かめる
    const tags = await user.api.listTags()
    expect(tags.map(({ name, color }) => ({ name, color }))).toEqual([
      { name: '仕事', color: 'blue' },
    ])
  })

  test('既存のタグと同じ名前では作成できず、エラーが表示される', async ({
    page,
    user,
    signIn,
  }) => {
    await user.api.createTag({ name: 'Work', color: 'red' })

    await signIn(user)

    const modal = await openTagManagement(page)
    await modal.getByRole('button', { name: '新しいタグ', exact: true }).click()
    // 大文字・小文字の違いだけの名前も重複として扱われる
    await modal.getByLabel('タグ名').fill('work')
    await modal.getByRole('button', { name: '保存', exact: true }).click()

    await expect(
      modal.getByText('同じ名前のタグがすでにあります'),
    ).toBeVisible()
    await expect(modal.getByLabel('タグ名')).toBeVisible()
    expect((await user.api.listTags()).map((t) => t.name)).toEqual(['Work'])
  })

  test('タグを編集すると、タグ一覧と予定の詳細に変更が反映される', async ({
    page,
    user,
    signIn,
  }) => {
    const tag = await user.api.createTag({ name: 'Work', color: 'red' })
    await user.api.createEvent({
      title: 'タグ付きの予定',
      ...EVENT_TIME,
      tagIds: [tag.id],
    })

    await signIn(user)

    const modal = await openTagManagement(page)
    await modal.getByRole('button', { name: 'Workを編集', exact: true }).click()
    await expect(modal.getByLabel('タグ名')).toHaveValue('Work')
    await expect(
      modal.getByRole('radio', { name: '赤', exact: true }),
    ).toHaveAttribute('aria-checked', 'true')

    await modal.getByLabel('タグ名').fill('仕事')
    await modal.getByRole('radio', { name: '緑', exact: true }).click()
    await modal.getByRole('button', { name: '保存', exact: true }).click()

    await expect(modal.getByLabel('タグ名')).toBeHidden()
    await expect(tagBadges(modal)).toHaveText(['仕事'])

    await page.keyboard.press('Escape')
    await expect(modal).toBeHidden()

    await page.getByRole('table').getByText('タグ付きの予定').click()
    const detail = page.getByRole('dialog', { name: 'タグ付きの予定' })
    await expect(tagBadges(detail)).toHaveText(['仕事'])

    const tags = await user.api.listTags()
    expect(tags.map(({ name, color }) => ({ name, color }))).toEqual([
      { name: '仕事', color: 'green' },
    ])
  })

  test('タグを削除すると、予定からタグが外れ、予定自体は残る', async ({
    page,
    user,
    signIn,
  }) => {
    const tag = await user.api.createTag({ name: 'Work', color: 'red' })
    await user.api.createEvent({
      title: 'タグ付きの予定',
      ...EVENT_TIME,
      tagIds: [tag.id],
    })

    await signIn(user)

    const modal = await openTagManagement(page)
    await modal.getByRole('button', { name: 'Workを削除', exact: true }).click()
    const confirm = page.getByRole('alertdialog', {
      name: 'タグ「Work」を削除しますか？',
    })
    // 付与されている予定の件数を示し、予定は消えないことを伝える
    await expect(
      confirm.getByText('1件の予定に付与されています', { exact: false }),
    ).toBeVisible()

    // キャンセルすると確認ダイアログだけが閉じ、タグは残る
    await confirm
      .getByRole('button', { name: 'キャンセル', exact: true })
      .click()
    await expect(confirm).toBeHidden()
    await expect(tagBadges(modal)).toHaveText(['Work'])

    await modal.getByRole('button', { name: 'Workを削除', exact: true }).click()
    await confirm.getByRole('button', { name: '削除する', exact: true }).click()

    await expect(confirm).toBeHidden()
    await expect(modal.getByText('タグはまだありません')).toBeVisible()

    await page.keyboard.press('Escape')
    await expect(modal).toBeHidden()

    await page.getByRole('table').getByText('タグ付きの予定').click()
    const detail = page.getByRole('dialog', { name: 'タグ付きの予定' })
    await expect(detail).toBeVisible()
    await expect(tagBadges(detail)).toHaveCount(0)

    const events = await user.api.listEvents(fixedDayRange(NEXT_DAY_OFFSET))
    expect(events.map(({ title, tags }) => ({ title, tags }))).toEqual([
      { title: 'タグ付きの予定', tags: [] },
    ])
    // 一覧の空表示は再取得前にキャッシュから外した結果でも出るため、タグ自体が消えたかは API で確かめる
    expect(await user.api.listTags()).toEqual([])
  })
})

test.describe('予定へのタグ付与', () => {
  test('予定フォームでタグを選択すると、詳細モーダルに選択した順でタグが表示される', async ({
    page,
    user,
    signIn,
  }) => {
    await user.api.createTag({ name: '仕事', color: 'blue' })
    await user.api.createTag({ name: '重要', color: 'red' })
    await user.api.createTag({ name: 'プライベート', color: 'green' })

    await signIn(user)

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('タグを付ける予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(14)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))

    // 作成順とは逆の順に選び、選んだ順が付与順になることを確かめる
    const openSelector = form.getByRole('button', {
      name: 'タグを選択',
      exact: true,
    })
    await openSelector.click()
    await page.getByRole('checkbox', { name: '重要', exact: true }).click()
    await page.getByRole('checkbox', { name: '仕事', exact: true }).click()
    await openSelector.click()
    await expect(tagPopover(page)).toBeHidden()
    await expect(tagBadges(tagSelector(form))).toHaveText(['重要', '仕事'])

    await form.getByRole('button', { name: '保存', exact: true }).click()
    await expect(form).toBeHidden()

    await page.getByRole('table').getByText('タグを付ける予定').click()
    const detail = page.getByRole('dialog', { name: 'タグを付ける予定' })
    await expect(tagBadges(detail)).toHaveText(['重要', '仕事'])
  })

  test('予定フォームから新しいタグを作成すると、そのタグが選択された状態になる', async ({
    page,
    user,
    signIn,
  }) => {
    await signIn(user)

    await page.getByRole('button', { name: '新規作成', exact: true }).click()
    const form = page.getByRole('dialog', { name: '予定を作成' })
    await form.getByLabel('タイトル').fill('新しいタグを付ける予定')
    await form.getByLabel('開始日時').fill(fixedDayInputValue(nextDayAt(14)))
    await form.getByLabel('終了日時').fill(fixedDayInputValue(nextDayAt(15)))

    const openSelector = form.getByRole('button', {
      name: 'タグを選択',
      exact: true,
    })
    await openSelector.click()
    const popover = tagPopover(page)
    await expect(popover.getByText('タグはまだありません')).toBeVisible()
    await popover
      .getByRole('button', { name: '新しいタグを作成', exact: true })
      .click()
    await popover.getByLabel('タグ名').fill('会議')
    await popover.getByRole('radio', { name: '紫', exact: true }).click()
    await popover.getByRole('button', { name: '保存', exact: true }).click()

    // タグの作成でフォームを離れず、作成したタグがチェック済みで一覧に並ぶ
    await expect(
      page.getByRole('checkbox', { name: '会議', exact: true }),
    ).toBeChecked()
    await expect(form).toBeVisible()
    await openSelector.click()
    await expect(popover).toBeHidden()
    await expect(tagBadges(tagSelector(form))).toHaveText(['会議'])

    await form.getByRole('button', { name: '保存', exact: true }).click()
    await expect(form).toBeHidden()

    await page.getByRole('table').getByText('新しいタグを付ける予定').click()
    const detail = page.getByRole('dialog', { name: '新しいタグを付ける予定' })
    await expect(tagBadges(detail)).toHaveText(['会議'])
  })

  test('予定の編集でタグを外すと、詳細モーダルからそのタグが消える', async ({
    page,
    user,
    signIn,
  }) => {
    const work = await user.api.createTag({ name: '仕事', color: 'blue' })
    const important = await user.api.createTag({ name: '重要', color: 'red' })
    await user.api.createEvent({
      title: 'タグを外す予定',
      ...EVENT_TIME,
      tagIds: [work.id, important.id],
    })

    await signIn(user)

    const calendar = page.getByRole('table')
    await calendar.getByText('タグを外す予定').click()
    const detail = page.getByRole('dialog', { name: 'タグを外す予定' })
    await expect(tagBadges(detail)).toHaveText(['仕事', '重要'])
    await detail.getByRole('button', { name: '編集', exact: true }).click()

    const form = page.getByRole('dialog', { name: '予定を編集' })
    await expect(tagBadges(tagSelector(form))).toHaveText(['仕事', '重要'])
    await form.getByRole('button', { name: '仕事を外す', exact: true }).click()
    await expect(tagBadges(tagSelector(form))).toHaveText(['重要'])
    await form.getByRole('button', { name: '保存', exact: true }).click()
    await expect(form).toBeHidden()

    await calendar.getByText('タグを外す予定').click()
    await expect(tagBadges(detail)).toHaveText(['重要'])

    // 予定から外してもタグ自体は残る
    expect((await user.api.listTags()).map((t) => t.name)).toEqual([
      '仕事',
      '重要',
    ])
  })
})
