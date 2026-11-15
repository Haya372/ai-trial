import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import { type ReactNode, useState } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { EventTagResponse, TagResponse } from '../../../api/generated'
import { createTag, listTags } from '../../../api/generated'
import EventTagSelector from './EventTagSelector'

vi.mock('../../../api/generated', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../../api/generated')>()
  return { ...actual, listTags: vi.fn(), createTag: vi.fn() }
})

function tag(id: string, name: string): TagResponse {
  return { id, name, color: 'blue', eventCount: 0 }
}

const work = tag('tag-1', '仕事')
const hobby = tag('tag-2', '趣味')
const urgent = tag('tag-3', '至急')

function response(status: number, data: unknown = {}) {
  return { status, data, headers: new Headers() } as never
}

function mockTags(tags: TagResponse[]) {
  vi.mocked(listTags).mockResolvedValue(response(200, { tags }))
}

function setup({
  initialValue = [],
  initialTags = [],
}: {
  initialValue?: string[]
  initialTags?: EventTagResponse[]
} = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const onChange = vi.fn()
  function Harness() {
    const [value, setValue] = useState(initialValue)
    return (
      <EventTagSelector
        labelId="tag-label"
        value={value}
        onChange={(ids) => {
          onChange(ids)
          setValue(ids)
        }}
        initialTags={initialTags}
      />
    )
  }
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  render(<Harness />, { wrapper })
  return { onChange }
}

async function openList() {
  fireEvent.click(screen.getByRole('button', { name: 'タグを選択' }))
  return screen.findByRole('dialog')
}

function selectedTagNames() {
  return screen
    .queryAllByRole('button', { name: /を外す$/ })
    .map((button) => button.getAttribute('aria-label')?.replace(/を外す$/, ''))
}

beforeEach(() => {
  vi.mocked(listTags).mockReset()
  vi.mocked(createTag).mockReset()
})

describe('EventTagSelector', () => {
  it('作成済みタグを一覧表示する', async () => {
    mockTags([work, hobby])
    setup()

    const list = await openList()

    expect(
      await within(list).findByRole('checkbox', { name: '仕事' }),
    ).toBeInTheDocument()
    expect(within(list).getByRole('checkbox', { name: '趣味' })).toBeVisible()
  })

  it('チェックした順に末尾へ追加し、チェックを外すと除く', async () => {
    mockTags([work, hobby, urgent])
    const { onChange } = setup()
    const list = await openList()

    fireEvent.click(await within(list).findByRole('checkbox', { name: '至急' }))
    fireEvent.click(within(list).getByRole('checkbox', { name: '仕事' }))
    expect(onChange).toHaveBeenLastCalledWith(['tag-3', 'tag-1'])
    expect(selectedTagNames()).toEqual(['至急', '仕事'])

    fireEvent.click(within(list).getByRole('checkbox', { name: '至急' }))
    expect(onChange).toHaveBeenLastCalledWith(['tag-1'])
  })

  it('選択中のタグを × で外せる', async () => {
    mockTags([work, hobby])
    const { onChange } = setup({ initialValue: ['tag-1', 'tag-2'] })

    fireEvent.click(await screen.findByRole('button', { name: '仕事を外す' }))

    expect(onChange).toHaveBeenLastCalledWith(['tag-2'])
  })

  it('タグ一覧にない選択中のタグは initialTags の名前で表示する', async () => {
    mockTags([hobby])
    setup({
      initialValue: ['tag-deleted', 'tag-2'],
      initialTags: [
        { id: 'tag-deleted', name: '削除済み', color: 'red' },
        { id: 'tag-2', name: '趣味', color: 'green' },
      ],
    })

    await waitFor(() => expect(listTags).toHaveBeenCalled())
    expect(selectedTagNames()).toEqual(['削除済み', '趣味'])
  })

  it('5件選択済みのときは未選択のタグと「新しいタグを作成」を無効にし、上限の旨を表示する', async () => {
    const tags = Array.from({ length: 6 }, (_, i) => tag(`tag-${i}`, `T${i}`))
    mockTags(tags)
    setup({ initialValue: tags.slice(0, 5).map((t) => t.id) })

    const list = await openList()

    const unselected = await within(list).findByRole('checkbox', { name: 'T5' })
    expect(unselected).toHaveAttribute('aria-disabled', 'true')
    expect(
      within(list).getByRole('checkbox', { name: 'T0' }),
    ).not.toHaveAttribute('aria-disabled', 'true')
    expect(
      within(list).getByRole('button', { name: '新しいタグを作成' }),
    ).toBeDisabled()
    expect(
      within(list).getByText('タグは1つの予定に5個まで付与できます'),
    ).toBeInTheDocument()
  })

  it('タグが50件のときは「新しいタグを作成」を無効にする', async () => {
    mockTags(Array.from({ length: 50 }, (_, i) => tag(`tag-${i}`, `T${i}`)))
    setup()

    const list = await openList()

    await within(list).findByRole('checkbox', { name: 'T0' })
    expect(
      within(list).getByRole('button', { name: '新しいタグを作成' }),
    ).toBeDisabled()
  })

  it('「新しいタグを作成」でタグを作成すると、レスポンスの id を末尾に追加する', async () => {
    mockTags([work])
    vi.mocked(createTag).mockResolvedValue(
      response(201, {
        id: 'tag-new',
        name: '新規',
        color: 'red',
        eventCount: 0,
      }),
    )
    const { onChange } = setup({ initialValue: ['tag-1'] })
    const list = await openList()

    fireEvent.click(
      await within(list).findByRole('button', { name: '新しいタグを作成' }),
    )
    fireEvent.change(within(list).getByLabelText('タグ名'), {
      target: { value: '新規' },
    })
    fireEvent.click(within(list).getByRole('button', { name: '保存' }))

    await waitFor(() =>
      expect(onChange).toHaveBeenLastCalledWith(['tag-1', 'tag-new']),
    )
    expect(createTag).toHaveBeenCalledWith({ name: '新規', color: 'red' })
  })

  it('タグ作成が 409 TAG_LIMIT_EXCEEDED のときは一覧に戻り、上限の旨を表示する', async () => {
    mockTags([work])
    vi.mocked(createTag).mockResolvedValue(
      response(409, { code: 'TAG_LIMIT_EXCEEDED', message: 'limit' }),
    )
    setup()
    const list = await openList()

    fireEvent.click(
      await within(list).findByRole('button', { name: '新しいタグを作成' }),
    )
    fireEvent.change(within(list).getByLabelText('タグ名'), {
      target: { value: '新規' },
    })
    fireEvent.click(within(list).getByRole('button', { name: '保存' }))

    expect(
      await within(list).findByText(
        'タグは50件まで作成できます。不要なタグを削除してください',
      ),
    ).toBeInTheDocument()
    expect(
      within(list).getByRole('button', { name: '新しいタグを作成' }),
    ).toBeDisabled()
  })

  it('タグがないときはその旨を表示する', async () => {
    mockTags([])
    setup()

    const list = await openList()

    expect(
      await within(list).findByText('タグはまだありません'),
    ).toBeInTheDocument()
  })

  it('タグ一覧の取得に失敗したらエラーを表示する', async () => {
    vi.mocked(listTags).mockResolvedValue(response(500))
    setup()

    const list = await openList()

    expect(
      await within(list).findByText('タグを読み込めませんでした'),
    ).toBeInTheDocument()
  })
})
