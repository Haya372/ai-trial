import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TagResponse } from '../../../api/generated'
import { listTags } from '../../../api/generated'
import { tagsKeys } from '../../../lib/queryKeys'
import TagFilterButton from './TagFilterButton'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, listTags: vi.fn() }
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

function setup(selectedTagId: string | null = null) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const onChange = vi.fn()
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  render(
    <TagFilterButton selectedTagId={selectedTagId} onChange={onChange} />,
    {
      wrapper,
    },
  )
  return { onChange, queryClient }
}

async function openPopover(
  trigger: HTMLElement = screen.getByRole('button', { name: '絞り込み' }),
) {
  fireEvent.click(trigger)
  return screen.findByRole('dialog')
}

beforeEach(() => {
  vi.mocked(listTags).mockReset()
})

describe('TagFilterButton', () => {
  it('未絞り込みのとき「絞り込み」ボタンを表示し、タグ一覧は取得しない', () => {
    mockTags([work])
    setup()

    expect(screen.getByRole('button', { name: '絞り込み' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '絞り込みを解除' })).toBeNull()
    expect(listTags).not.toHaveBeenCalled()
  })

  it('クリックするとタグ一覧と「すべて表示」を表示する', async () => {
    mockTags([work, hobby])
    setup()

    await openPopover()

    expect(
      await screen.findByRole('button', { name: '仕事' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '趣味' })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'すべて表示' }),
    ).toBeInTheDocument()
  })

  it('未絞り込みのときは「すべて表示」を押下状態にする', async () => {
    mockTags([work])
    setup()

    await openPopover()

    expect(
      await screen.findByRole('button', { name: 'すべて表示' }),
    ).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: '仕事' })).toHaveAttribute(
      'aria-pressed',
      'false',
    )
  })

  it('タグをクリックするとそのタグのIDで onChange を呼び、一覧を閉じる', async () => {
    mockTags([work, hobby])
    const { onChange } = setup()

    await openPopover()
    fireEvent.click(await screen.findByRole('button', { name: '趣味' }))

    expect(onChange).toHaveBeenCalledWith('tag-2')
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('取得中は読み込み中を表示する', async () => {
    vi.mocked(listTags).mockImplementation(() => new Promise(() => {}))
    setup()

    await openPopover()

    expect(await screen.findByText('読み込み中...')).toBeInTheDocument()
  })

  it('取得に失敗するとエラーを表示する', async () => {
    vi.mocked(listTags).mockResolvedValue(response(500))
    setup()

    await openPopover()

    expect(
      await screen.findByText('タグを読み込めませんでした'),
    ).toBeInTheDocument()
  })

  it('再取得に失敗しても、取得済みのタグ一覧があれば一覧を表示する', async () => {
    vi.mocked(listTags)
      .mockResolvedValueOnce(response(200, { tags: [work] }))
      .mockResolvedValue(response(500))
    const { queryClient } = setup()
    await openPopover()
    await screen.findByRole('button', { name: '仕事' })

    await act(() => queryClient.refetchQueries({ queryKey: tagsKeys.all }))
    // Query results reach components through notifyManager's batched timer.
    await act(() => new Promise((resolve) => setTimeout(resolve, 0)))

    expect(queryClient.getQueryState(tagsKeys.all)?.status).toBe('error')
    expect(screen.getByRole('button', { name: '仕事' })).toBeInTheDocument()
    expect(screen.queryByText('タグを読み込めませんでした')).toBeNull()
  })

  describe('タグが0件', () => {
    it('タグがない旨と「タグ管理」への導線を表示し、クリックでタグ管理モーダルを開く', async () => {
      mockTags([])
      setup()

      await openPopover()
      expect(
        await screen.findByText('タグはまだありません'),
      ).toBeInTheDocument()
      fireEvent.click(screen.getByRole('button', { name: 'タグ管理' }))

      expect(
        await screen.findByRole('dialog', { name: 'タグ管理' }),
      ).toBeInTheDocument()
    })
  })

  describe('絞り込み中', () => {
    it('ボタンに選択中のタグ名を表示し、一覧では選択中のタグを押下状態にする', async () => {
      mockTags([work, hobby])
      setup('tag-1')

      const trigger = await screen.findByRole('button', {
        name: '絞り込み: 仕事',
      })
      expect(trigger).toHaveTextContent('仕事')
      await openPopover(trigger)

      expect(screen.getByRole('button', { name: '仕事' })).toHaveAttribute(
        'aria-pressed',
        'true',
      )
      expect(
        screen.getByRole('button', { name: 'すべて表示' }),
      ).toHaveAttribute('aria-pressed', 'false')
      expect(screen.getByRole('button', { name: '趣味' })).toHaveAttribute(
        'aria-pressed',
        'false',
      )
    })

    it('別のタグを選ぶとそのタグのIDで onChange を呼ぶ', async () => {
      mockTags([work, hobby])
      const { onChange } = setup('tag-1')

      await openPopover(
        await screen.findByRole('button', { name: '絞り込み: 仕事' }),
      )
      fireEvent.click(await screen.findByRole('button', { name: '趣味' }))

      expect(onChange).toHaveBeenCalledWith('tag-2')
    })

    it('「すべて表示」で onChange(null) を呼ぶ', async () => {
      mockTags([work, hobby])
      const { onChange } = setup('tag-1')

      await openPopover(
        await screen.findByRole('button', { name: '絞り込み: 仕事' }),
      )
      fireEvent.click(await screen.findByRole('button', { name: 'すべて表示' }))

      expect(onChange).toHaveBeenCalledWith(null)
    })

    it('解除（×）で onChange(null) を呼ぶ', async () => {
      mockTags([work])
      const { onChange } = setup('tag-1')

      fireEvent.click(screen.getByRole('button', { name: '絞り込みを解除' }))

      expect(onChange).toHaveBeenCalledWith(null)
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    it('取得したタグ一覧に選択中のタグが含まれなければ onChange(null) を呼ぶ', async () => {
      mockTags([hobby])
      const { onChange } = setup('tag-1')

      await waitFor(() => expect(onChange).toHaveBeenCalledWith(null))
    })

    it('タグ一覧が更新されて選択中のタグがなくなったら onChange(null) を呼ぶ', async () => {
      mockTags([work, hobby])
      const { onChange, queryClient } = setup('tag-1')
      await screen.findByRole('button', { name: '絞り込み: 仕事' })

      queryClient.setQueryData(tagsKeys.all, [hobby])

      await waitFor(() => expect(onChange).toHaveBeenCalledWith(null))
    })

    it('タグ一覧が未取得のうちは解除しない', () => {
      mockTags([work])
      const { onChange } = setup('tag-1')

      expect(onChange).not.toHaveBeenCalled()
    })
  })
})
