import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TagResponse } from '../../../api/generated'
import {
  createTag,
  deleteTag,
  listTags,
  updateTag,
} from '../../../api/generated'
import TagManagementButton from './TagManagementButton'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return {
    ...actual,
    listTags: vi.fn(),
    createTag: vi.fn(),
    updateTag: vi.fn(),
    deleteTag: vi.fn(),
  }
})

const work: TagResponse = {
  id: 'tag-1',
  name: '仕事',
  color: 'blue',
  eventCount: 3,
}
const hobby: TagResponse = {
  id: 'tag-2',
  name: '趣味',
  color: 'green',
  eventCount: 0,
}

function response(status: number, data: unknown = {}) {
  return { status, data, headers: new Headers() } as never
}

function mockTags(tags: TagResponse[]) {
  vi.mocked(listTags).mockResolvedValue(response(200, { tags }))
}

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  render(<TagManagementButton />, { wrapper })
  return { invalidateQueries }
}

function invalidatedKeys(spy: ReturnType<typeof setup>['invalidateQueries']) {
  return spy.mock.calls.map(([filters]) => filters?.queryKey)
}

async function openModal() {
  fireEvent.click(screen.getByRole('button', { name: 'タグ管理' }))
  return screen.findByRole('dialog', { name: 'タグ管理' })
}

beforeEach(() => {
  vi.mocked(listTags).mockReset()
  vi.mocked(createTag).mockReset()
  vi.mocked(updateTag).mockReset()
  vi.mocked(deleteTag).mockReset()
})

describe('TagManagementButton', () => {
  describe('モーダルを開く', () => {
    it('ボタンを押すまでは GET /tags を呼ばない', () => {
      mockTags([work])
      setup()
      expect(listTags).not.toHaveBeenCalled()
    })

    it('ボタンを押すと GET /tags を呼び、タグ一覧を表示する', async () => {
      mockTags([work, hobby])
      setup()
      const dialog = await openModal()

      expect(listTags).toHaveBeenCalledTimes(1)
      expect(await within(dialog).findByText('仕事')).toBeInTheDocument()
      expect(within(dialog).getByText('趣味')).toBeInTheDocument()
    })

    it('タグがないときは空の旨を表示する', async () => {
      mockTags([])
      setup()
      const dialog = await openModal()
      expect(
        await within(dialog).findByText('タグはまだありません'),
      ).toBeInTheDocument()
    })

    it('一覧の取得に失敗したらモーダル内にエラーを表示する', async () => {
      vi.mocked(listTags).mockResolvedValue(response(500))
      setup()
      const dialog = await openModal()
      expect(await within(dialog).findByRole('alert')).toHaveTextContent(
        'タグを読み込めませんでした',
      )
    })
  })

  describe('タグの作成', () => {
    it('タグ名と色を指定して作成し、タグ一覧のキャッシュを無効化する', async () => {
      mockTags([work])
      vi.mocked(createTag).mockResolvedValue(
        response(201, { id: 'tag-3', name: '家族', color: 'pink' }),
      )
      const { invalidateQueries } = setup()
      const dialog = await openModal()
      await within(dialog).findByText('仕事')

      fireEvent.click(
        within(dialog).getByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: '  家族 ' },
      })
      fireEvent.click(within(dialog).getByRole('radio', { name: 'ピンク' }))
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      await waitFor(() =>
        expect(createTag).toHaveBeenCalledWith({ name: '家族', color: 'pink' }),
      )
      await waitFor(() =>
        expect(invalidatedKeys(invalidateQueries)).toEqual([['tags']]),
      )
      expect(within(dialog).queryByLabelText('タグ名')).not.toBeInTheDocument()
    })

    it('カラーパレットに12色を SPEC-008 の表示順で並べる', async () => {
      mockTags([])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )

      const palette = within(dialog).getByRole('radiogroup', { name: '色' })
      expect(
        within(palette)
          .getAllByRole('radio')
          .map((radio) => radio.getAttribute('aria-label')),
      ).toEqual([
        '赤',
        'オレンジ',
        '黄',
        '緑',
        'ティール',
        '水色',
        '青',
        '藍',
        '紫',
        'ピンク',
        '茶',
        'グレー',
      ])
    })

    it('矢印キーで色を選び直せる', async () => {
      mockTags([])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )

      fireEvent.click(within(dialog).getByRole('radio', { name: '青' }))
      fireEvent.keyDown(within(dialog).getByRole('radio', { name: '青' }), {
        key: 'ArrowRight',
      })

      const indigo = within(dialog).getByRole('radio', { name: '藍' })
      expect(indigo).toHaveAttribute('aria-checked', 'true')
      expect(indigo).toHaveFocus()
    })

    it('タグ名が空のときはインラインでエラーを表示し、保存しない', async () => {
      mockTags([])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(
        await within(dialog).findByText('タグ名を入力してください'),
      ).toBeInTheDocument()
      expect(within(dialog).getByLabelText('タグ名')).toHaveAttribute(
        'aria-invalid',
        'true',
      )
      expect(createTag).not.toHaveBeenCalled()
    })

    it('タグ名が21文字以上のときはインラインでエラーを表示し、保存しない', async () => {
      mockTags([])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: 'a'.repeat(21) },
      })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(
        await within(dialog).findByText('タグ名は20文字以内で入力してください'),
      ).toBeInTheDocument()
      expect(createTag).not.toHaveBeenCalled()
    })

    it('409 TAG_NAME_TAKEN のときはタグ名欄に重複のエラーを表示する', async () => {
      mockTags([work])
      vi.mocked(createTag).mockResolvedValue(
        response(409, { code: 'TAG_NAME_TAKEN', message: 'taken' }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: '仕事' },
      })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(
        await within(dialog).findByText('同じ名前のタグがすでにあります'),
      ).toBeInTheDocument()
      expect(within(dialog).getByLabelText('タグ名')).toHaveAttribute(
        'aria-invalid',
        'true',
      )
    })

    it('409 TAG_LIMIT_EXCEEDED のときは上限に達した旨を表示する', async () => {
      mockTags([work])
      vi.mocked(createTag).mockResolvedValue(
        response(409, { code: 'TAG_LIMIT_EXCEEDED', message: 'limit' }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: '家族' },
      })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(
        await within(dialog).findByText(/タグは50件まで作成できます/),
      ).toBeInTheDocument()
      expect(
        within(dialog).getByRole('button', { name: '新しいタグ' }),
      ).toBeDisabled()
    })

    it('その他のエラーのときはモーダル内にエラーを表示し、モーダルを閉じない', async () => {
      mockTags([work])
      vi.mocked(createTag).mockResolvedValue(
        response(500, { code: 'INTERNAL_ERROR', message: 'boom' }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: '家族' },
      })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(await within(dialog).findByRole('alert')).toHaveTextContent(
        'タグの保存に失敗しました',
      )
      expect(
        screen.getByRole('dialog', { name: 'タグ管理' }),
      ).toBeInTheDocument()
      expect(within(dialog).getByLabelText('タグ名')).toHaveValue('家族')
    })

    it('通信自体が失敗したときもモーダル内にエラーを表示する', async () => {
      mockTags([work])
      vi.mocked(createTag).mockRejectedValue(new TypeError('Failed to fetch'))
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '新しいタグ' }),
      )
      fireEvent.change(within(dialog).getByLabelText('タグ名'), {
        target: { value: '家族' },
      })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(await within(dialog).findByRole('alert')).toHaveTextContent(
        'タグの保存に失敗しました',
      )
    })
  })

  describe('タグの上限', () => {
    it('タグが50件のときは「新しいタグ」を無効にし、上限に達した旨を表示する', async () => {
      mockTags(
        Array.from({ length: 50 }, (_, i) => ({
          id: `tag-${i}`,
          name: `タグ${i}`,
          color: 'gray' as const,
          eventCount: 0,
        })),
      )
      setup()
      const dialog = await openModal()
      await within(dialog).findByText('タグ0')

      expect(
        within(dialog).getByRole('button', { name: '新しいタグ' }),
      ).toBeDisabled()
      expect(
        within(dialog).getByText(/タグは50件まで作成できます/),
      ).toBeInTheDocument()
    })

    it('タグが49件のときは「新しいタグ」を押せる', async () => {
      mockTags(
        Array.from({ length: 49 }, (_, i) => ({
          id: `tag-${i}`,
          name: `タグ${i}`,
          color: 'gray' as const,
          eventCount: 0,
        })),
      )
      setup()
      const dialog = await openModal()
      await within(dialog).findByText('タグ0')

      expect(
        within(dialog).getByRole('button', { name: '新しいタグ' }),
      ).toBeEnabled()
      expect(
        within(dialog).queryByText(/タグは50件まで作成できます/),
      ).not.toBeInTheDocument()
    })
  })

  describe('タグの編集', () => {
    it('既存タグの名前と色を変更し、タグ一覧と予定一覧のキャッシュを無効化する', async () => {
      mockTags([work])
      vi.mocked(updateTag).mockResolvedValue(
        response(200, { id: 'tag-1', name: '業務', color: 'indigo' }),
      )
      const { invalidateQueries } = setup()
      const dialog = await openModal()

      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を編集' }),
      )
      const nameInput = within(dialog).getByLabelText('タグ名')
      expect(nameInput).toHaveValue('仕事')
      expect(within(dialog).getByRole('radio', { name: '青' })).toHaveAttribute(
        'aria-checked',
        'true',
      )

      fireEvent.change(nameInput, { target: { value: '業務' } })
      fireEvent.click(within(dialog).getByRole('radio', { name: '藍' }))
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      await waitFor(() =>
        expect(updateTag).toHaveBeenCalledWith('tag-1', {
          name: '業務',
          color: 'indigo',
        }),
      )
      await waitFor(() =>
        expect(invalidatedKeys(invalidateQueries)).toEqual(
          expect.arrayContaining([['tags'], ['events']]),
        ),
      )
    })

    it('キャンセルすると変更せずに一覧に戻る', async () => {
      mockTags([work])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を編集' }),
      )
      fireEvent.click(
        within(dialog).getByRole('button', { name: 'キャンセル' }),
      )

      expect(within(dialog).queryByLabelText('タグ名')).not.toBeInTheDocument()
      expect(within(dialog).getByText('仕事')).toBeInTheDocument()
      expect(updateTag).not.toHaveBeenCalled()
    })

    it('409 TAG_NAME_TAKEN のときはタグ名欄に重複のエラーを表示する', async () => {
      mockTags([work, hobby])
      vi.mocked(updateTag).mockResolvedValue(
        response(409, { code: 'TAG_NAME_TAKEN', message: 'taken' }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を編集' }),
      )
      const nameInput = within(dialog).getByLabelText('タグ名')
      fireEvent.change(nameInput, { target: { value: '趣味' } })
      fireEvent.click(within(dialog).getByRole('button', { name: '保存' }))

      expect(
        await within(dialog).findByText('同じ名前のタグがすでにあります'),
      ).toBeInTheDocument()
    })
  })

  describe('タグの削除', () => {
    it('確認ダイアログに付与件数を表示し、削除するとタグ一覧と予定一覧のキャッシュを無効化する', async () => {
      mockTags([work])
      vi.mocked(deleteTag).mockResolvedValue(response(204))
      const { invalidateQueries } = setup()
      const dialog = await openModal()

      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を削除' }),
      )
      const confirm = await screen.findByRole('alertdialog')
      expect(confirm).toHaveTextContent('タグ「仕事」を削除しますか？')
      expect(confirm).toHaveTextContent('3件の予定に付与されています')

      fireEvent.click(within(confirm).getByRole('button', { name: '削除する' }))

      await waitFor(() => expect(deleteTag).toHaveBeenCalledWith('tag-1'))
      await waitFor(() =>
        expect(invalidatedKeys(invalidateQueries)).toEqual(
          expect.arrayContaining([['tags'], ['events']]),
        ),
      )
      await waitFor(() =>
        expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument(),
      )
    })

    it('付与件数が0件でも件数を表示する', async () => {
      mockTags([hobby])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '趣味を削除' }),
      )
      expect(await screen.findByRole('alertdialog')).toHaveTextContent(
        '0件の予定に付与されています',
      )
    })

    it('キャンセルすると削除しない', async () => {
      mockTags([work])
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を削除' }),
      )
      const confirm = await screen.findByRole('alertdialog')
      fireEvent.click(
        within(confirm).getByRole('button', { name: 'キャンセル' }),
      )

      await waitFor(() =>
        expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument(),
      )
      expect(deleteTag).not.toHaveBeenCalled()
    })

    // Closing it mid-request would let a stale completion close another tag's dialog.
    it('削除中は Escape で確認ダイアログを閉じられない', async () => {
      mockTags([work])
      let resolveDelete: (value: never) => void = () => {}
      vi.mocked(deleteTag).mockReturnValue(
        new Promise((resolve) => {
          resolveDelete = resolve
        }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を削除' }),
      )
      const confirm = await screen.findByRole('alertdialog')
      fireEvent.click(within(confirm).getByRole('button', { name: '削除する' }))
      await waitFor(() =>
        expect(
          within(confirm).getByRole('button', { name: '削除する' }),
        ).toHaveAttribute('aria-busy', 'true'),
      )

      fireEvent.keyDown(confirm, { key: 'Escape' })
      await new Promise((resolve) => setTimeout(resolve, 50))
      expect(screen.getByRole('alertdialog')).toBeInTheDocument()

      resolveDelete(response(204))
      await waitFor(() =>
        expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument(),
      )
    })

    it('削除に失敗したらモーダル内にエラーを表示し、モーダルを閉じない', async () => {
      mockTags([work])
      vi.mocked(deleteTag).mockResolvedValue(
        response(500, { code: 'INTERNAL_ERROR', message: 'boom' }),
      )
      setup()
      const dialog = await openModal()
      fireEvent.click(
        await within(dialog).findByRole('button', { name: '仕事を削除' }),
      )
      const confirm = await screen.findByRole('alertdialog')
      fireEvent.click(within(confirm).getByRole('button', { name: '削除する' }))

      expect(await within(dialog).findByRole('alert')).toHaveTextContent(
        'タグの削除に失敗しました',
      )
      expect(
        screen.getByRole('dialog', { name: 'タグ管理' }),
      ).toBeInTheDocument()
      expect(within(dialog).getByText('仕事')).toBeInTheDocument()
    })
  })
})
