import { type APIRequestContext, type APIResponse, expect } from '@playwright/test'

export type CreateEventInput = {
  title: string
  startAt: string
  endAt: string
  description?: string | null
  location?: string | null
  url?: string | null
  tagIds?: string[]
}

export type TagColor =
  | 'red'
  | 'orange'
  | 'yellow'
  | 'green'
  | 'teal'
  | 'cyan'
  | 'blue'
  | 'indigo'
  | 'purple'
  | 'pink'
  | 'brown'
  | 'gray'

/** POST /tags・GET /tags が返すタグ（TagResponse） */
export type TagResponse = {
  id: string
  name: string
  color: TagColor
  eventCount: number
}

/** POST /events・GET /events が返す予定（EventResponse） */
export type EventResponse = Omit<CreateEventInput, 'tagIds'> & {
  id: string
  isSubscribed: boolean
  subscriptionId: string | null
  tags: { id: string; name: string; color: TagColor }[]
}

/** ログイン済みユーザーのセッション Cookie を持つ APIRequestContext で、UI を経ずにテストデータを用意する */
export class UserApi {
  constructor(private readonly request: APIRequestContext) {}

  async createEvent(input: CreateEventInput): Promise<EventResponse> {
    const res = await this.request.post('/api/events', { data: input })
    await expectStatus(res, 201)
    return res.json()
  }

  /** 共有リンクを発行し、`/share/{token}` 形式のパスを返す。有効期限は省略時に予定の終了日時になる */
  async createShare(eventId: string): Promise<string> {
    const res = await this.request.post(`/api/events/${eventId}/shares`, { data: {} })
    await expectStatus(res, 201)
    return (await res.json()).url
  }

  /** 共有リンクのパス（`/share/{token}`）から、共有された予定を自分のカレンダーに追加する */
  async subscribe(sharePath: string): Promise<void> {
    const token = sharePath.replace(/^\/share\//, '')
    const res = await this.request.post(`/api/shares/${token}/subscriptions`)
    await expectStatus(res, 201)
  }

  async createTag(input: { name: string; color: TagColor }): Promise<TagResponse> {
    const res = await this.request.post('/api/tags', { data: input })
    await expectStatus(res, 201)
    return res.json()
  }

  /** 画面表示に頼らず、サーバーに登録されているタグを確かめる */
  async listTags(): Promise<TagResponse[]> {
    const res = await this.request.get('/api/tags')
    await expectStatus(res, 200)
    return (await res.json()).tags
  }

  /** 画面表示に頼らず、サーバーに登録されている予定を確かめる */
  async listEvents(range: { startAt: string; endAt: string }): Promise<EventResponse[]> {
    const res = await this.request.get('/api/events', { params: range })
    await expectStatus(res, 200)
    return (await res.json()).events
  }
}

export async function expectStatus(res: APIResponse, status: number): Promise<void> {
  // 失敗時にレスポンス本文を出し、バリデーションエラー等の原因を追えるようにする
  expect(res.status(), `${res.url()}: ${await res.text()}`).toBe(status)
}
