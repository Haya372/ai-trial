import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import i18n from 'i18next'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { EventResponse, TagResponse } from '../../../api/generated'
import {
  createEvent,
  createTag,
  listTags,
  updateEvent,
} from '../../../api/generated'
import EventFormModal from './EventFormModal'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return {
    ...actual,
    createEvent: vi.fn(),
    updateEvent: vi.fn(),
    listTags: vi.fn(),
    createTag: vi.fn(),
  }
})

const workTag: TagResponse = {
  id: 'tag-1',
  name: '仕事',
  color: 'blue',
  eventCount: 0,
}
const hobbyTag: TagResponse = {
  id: 'tag-2',
  name: '趣味',
  color: 'green',
  eventCount: 0,
}

beforeEach(() => {
  vi.mocked(createEvent).mockReset()
  vi.mocked(updateEvent).mockReset()
  vi.mocked(createTag).mockReset()
  vi.mocked(listTags).mockReset()
  vi.mocked(listTags).mockResolvedValue({
    data: { tags: [workTag, hobbyTag] },
    status: 200,
    headers: new Headers(),
  } as never)
  mockToastError.mockClear()
})

const { mockToastError, mockToastSuccess } = vi.hoisted(() => ({
  mockToastError: vi.fn(),
  mockToastSuccess: vi.fn(),
}))

vi.mock('@repo/ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@repo/ui')>()
  return {
    ...actual,
    toast: Object.assign(vi.fn(), {
      error: mockToastError,
      success: mockToastSuccess,
    }),
  }
})

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

const existingEvent: EventResponse = {
  id: 'event-1',
  title: 'デザインレビュー',
  description: 'メモ',
  startAt: '2026-09-22T01:00:00.000Z',
  endAt: '2026-09-22T02:00:00.000Z',
  location: '会議室A',
  url: 'https://example.com',
  isSubscribed: false,
  subscriptionId: null,
  tags: [],
  recurringSeries: null,
}

function renderModal(
  props: Partial<React.ComponentProps<typeof EventFormModal>> = {},
) {
  const onClose = vi.fn()
  render(
    <EventFormModal
      open={true}
      mode="create"
      event={null}
      onClose={onClose}
      {...props}
    />,
    { wrapper: createWrapper() },
  )
  return { onClose }
}

describe('EventFormModal', () => {
  describe('表示', () => {
    it('createモードのとき「予定を作成」を表示する', () => {
      renderModal({ mode: 'create', event: null })
      expect(screen.getByText('予定を作成')).toBeInTheDocument()
    })

    it('editモードのとき「予定を編集」と既存の値を表示する', () => {
      renderModal({ mode: 'edit', event: existingEvent })
      expect(screen.getByText('予定を編集')).toBeInTheDocument()
      expect(screen.getByDisplayValue('デザインレビュー')).toBeInTheDocument()
      expect(screen.getByDisplayValue('メモ')).toBeInTheDocument()
      expect(screen.getByDisplayValue('会議室A')).toBeInTheDocument()
      expect(
        screen.getByDisplayValue('https://example.com'),
      ).toBeInTheDocument()
    })

    it('openがfalseのとき何も表示しない', () => {
      renderModal({ open: false })
      expect(screen.queryByText('予定を作成')).not.toBeInTheDocument()
    })

    it('createモードでinitialStartが渡された場合、開始日時の初期値に設定する', () => {
      const initialStart = new Date(2026, 8, 22, 14, 30)
      renderModal({ mode: 'create', event: null, initialStart })
      expect(screen.getByLabelText('開始日時')).toHaveValue('2026-09-22T14:30')
    })
  })

  describe('バリデーション', () => {
    it('タイトルが空の場合はエラーを表示し送信しない', async () => {
      renderModal()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('タイトルを入力してください'),
        ).toBeInTheDocument()
      })
      expect(createEvent).not.toHaveBeenCalled()
    })

    it('終了日時が開始日時以前の場合はエラーを表示する', async () => {
      renderModal()
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: 'テスト予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T09:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('終了日時は開始日時より後に設定してください'),
        ).toBeInTheDocument()
      })
    })

    it('表示中のバリデーションエラーは言語切り替え後に翻訳し直される', async () => {
      renderModal()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))
      await waitFor(() => {
        expect(
          screen.getByText('タイトルを入力してください'),
        ).toBeInTheDocument()
      })
      await i18n.changeLanguage('en')
      await waitFor(() => {
        expect(screen.getByText('Title is required')).toBeInTheDocument()
      })
    })
  })

  describe('送信', () => {
    it('createモードで保存するとcreateEventを呼び、成功後にonCloseを呼ぶ', async () => {
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: { ...existingEvent, id: 'new-event' },
        status: 201,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({ mode: 'create', event: null })
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '新しい予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T11:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(createEvent).toHaveBeenCalledWith(
          expect.objectContaining({ title: '新しい予定' }),
        )
      })
      await waitFor(() => expect(onClose).toHaveBeenCalled())
      expect(mockToastSuccess).toHaveBeenCalled()
    })

    it('editモードで保存するとupdateEventをidつきで呼ぶ', async () => {
      vi.mocked(updateEvent).mockResolvedValueOnce({
        data: existingEvent,
        status: 200,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({
        mode: 'edit',
        event: existingEvent,
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(updateEvent).toHaveBeenCalledWith(
          'event-1',
          expect.objectContaining({ title: 'デザインレビュー' }),
        )
      })
      await waitFor(() => expect(onClose).toHaveBeenCalled())
    })

    it('保存失敗時はエラートーストを表示しモーダルを閉じない', async () => {
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: { code: 'VALIDATION_ERROR', message: 'Validation error' },
        status: 400,
        headers: new Headers(),
      } as never)

      const { onClose } = renderModal({ mode: 'create', event: null })
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '新しい予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T11:00' },
      })
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalledWith(
          '入力内容を確認してください',
        )
      })
      expect(onClose).not.toHaveBeenCalled()
    })
  })

  describe('タグ', () => {
    function fillRequiredFields() {
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '新しい予定' },
      })
      fireEvent.change(screen.getByLabelText('開始日時'), {
        target: { value: '2026-09-22T10:00' },
      })
      fireEvent.change(screen.getByLabelText('終了日時'), {
        target: { value: '2026-09-22T11:00' },
      })
    }

    async function openTagList() {
      fireEvent.click(screen.getByRole('button', { name: 'タグを選択' }))
      await waitFor(() => expect(screen.getAllByRole('dialog')).toHaveLength(2))
      return screen.getAllByRole('dialog')[1]
    }

    function tagGroup() {
      return screen.getByRole('group', { name: 'タグ' })
    }

    it('createモードで選択した順に tagIds を送る', async () => {
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: existingEvent,
        status: 201,
        headers: new Headers(),
      } as never)
      renderModal({ mode: 'create', event: null })
      fillRequiredFields()

      const list = await openTagList()
      fireEvent.click(
        await within(list).findByRole('checkbox', { name: '趣味' }),
      )
      fireEvent.click(within(list).getByRole('checkbox', { name: '仕事' }))
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(createEvent).toHaveBeenCalledWith(
          expect.objectContaining({ tagIds: ['tag-2', 'tag-1'] }),
        )
      })
    })

    it('editモードでは既存のタグを初期選択し、updateEvent に tagIds を送る', async () => {
      vi.mocked(updateEvent).mockResolvedValueOnce({
        data: existingEvent,
        status: 200,
        headers: new Headers(),
      } as never)
      renderModal({
        mode: 'edit',
        event: {
          ...existingEvent,
          tags: [{ id: 'tag-2', name: '趣味', color: 'green' }],
        },
      })

      expect(
        within(tagGroup()).getByRole('button', { name: '趣味を外す' }),
      ).toBeInTheDocument()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => {
        expect(updateEvent).toHaveBeenCalledWith(
          'event-1',
          expect.objectContaining({ tagIds: ['tag-2'] }),
        )
      })
    })

    it('共有から追加した予定の編集ではタグ選択欄を表示せず、tagIds を送らない', async () => {
      vi.mocked(updateEvent).mockResolvedValueOnce({
        data: existingEvent,
        status: 200,
        headers: new Headers(),
      } as never)
      renderModal({
        mode: 'edit',
        event: {
          ...existingEvent,
          isSubscribed: true,
          subscriptionId: 'sub-1',
        },
      })

      expect(
        screen.queryByRole('group', { name: 'タグ' }),
      ).not.toBeInTheDocument()
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      await waitFor(() => expect(updateEvent).toHaveBeenCalled())
      expect(vi.mocked(updateEvent).mock.calls[0][1]).not.toHaveProperty(
        'tagIds',
      )
    })

    it('タグ選択欄でタグを作成しても予定は保存しない', async () => {
      const newTag: TagResponse = {
        id: 'tag-new',
        name: '新規',
        color: 'red',
        eventCount: 0,
      }
      vi.mocked(createTag).mockImplementationOnce(async () => {
        vi.mocked(listTags).mockResolvedValue({
          data: { tags: [workTag, hobbyTag, newTag] },
          status: 200,
          headers: new Headers(),
        } as never)
        return { data: newTag, status: 201, headers: new Headers() } as never
      })
      renderModal({ mode: 'create', event: null })
      fillRequiredFields()

      const list = await openTagList()
      fireEvent.click(
        await within(list).findByRole('button', { name: '新しいタグを作成' }),
      )
      fireEvent.change(within(list).getByLabelText('タグ名'), {
        target: { value: '新規' },
      })
      fireEvent.click(within(list).getByRole('button', { name: '保存' }))

      expect(
        await within(tagGroup()).findByRole('button', { name: '新規を外す' }),
      ).toBeInTheDocument()
      expect(createEvent).not.toHaveBeenCalled()
    })

    it('削除済みのタグが選択されていたら tagIds 欄にエラーを表示し、トーストは出さない', async () => {
      vi.mocked(createEvent).mockResolvedValueOnce({
        data: {
          code: 'VALIDATION_ERROR',
          message: 'Validation error',
          details: [
            { field: 'tagIds', code: 'NOT_FOUND', message: 'not found' },
          ],
        },
        status: 400,
        headers: new Headers(),
      } as never)
      const { onClose } = renderModal({ mode: 'create', event: null })
      fillRequiredFields()

      const list = await openTagList()
      fireEvent.click(
        await within(list).findByRole('checkbox', { name: '仕事' }),
      )
      fireEvent.click(screen.getByRole('button', { name: '保存' }))

      expect(
        await screen.findByText(
          '削除されたタグが選択されています。外してから保存してください',
        ),
      ).toBeInTheDocument()
      expect(mockToastError).not.toHaveBeenCalled()
      expect(onClose).not.toHaveBeenCalled()
    })
  })

  describe('キャンセル', () => {
    it('キャンセルボタンでonCloseを呼ぶ', () => {
      const { onClose } = renderModal()
      fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }))
      expect(onClose).toHaveBeenCalled()
    })
  })

  describe('再オープン時のリセット', () => {
    it('入力後に閉じて同じmode/eventで再度開くと入力値がリセットされる', () => {
      const onClose = vi.fn()
      const { rerender } = render(
        <EventFormModal
          open={true}
          mode="create"
          event={null}
          onClose={onClose}
        />,
        { wrapper: createWrapper() },
      )
      fireEvent.change(screen.getByLabelText('タイトル'), {
        target: { value: '一時的なタイトル' },
      })
      expect(screen.getByDisplayValue('一時的なタイトル')).toBeInTheDocument()

      rerender(
        <EventFormModal
          open={false}
          mode="create"
          event={null}
          onClose={onClose}
        />,
      )
      rerender(
        <EventFormModal
          open={true}
          mode="create"
          event={null}
          onClose={onClose}
        />,
      )

      expect(
        screen.queryByDisplayValue('一時的なタイトル'),
      ).not.toBeInTheDocument()
      expect(screen.getByLabelText('タイトル')).toHaveValue('')
    })
  })
})
