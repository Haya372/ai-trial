import {
  type APIRequestContext,
  type APIResponse,
  expect,
} from '@playwright/test'
import {
  type CreateEventRequest,
  type CreateEventShareRequest,
  type CreateEventShareResponse,
  type CreateTagRequest,
  type EventResponse,
  type EventsListResponse,
  type GetEventsParams,
  getCreateEventShareUrl,
  getCreateEventUrl,
  getCreateTagUrl,
  getGetEventsUrl,
  getListTagsUrl,
  getSubscribeToShareUrl,
  type TagResponse,
  type TagsListResponse,
} from './generated/api'

/** ログイン済みユーザーのセッション Cookie を持つ APIRequestContext で、UI を経ずにテストデータを用意する */
export class UserApi {
  constructor(private readonly request: APIRequestContext) {}

  async createEvent(input: CreateEventRequest): Promise<EventResponse> {
    const res = await this.request.post(getCreateEventUrl(), { data: input })
    await expectStatus(res, 201)
    return res.json()
  }

  /** 共有リンクを発行し、`/share/{token}` 形式のパスを返す。有効期限は省略時に予定の終了日時になる */
  async createShare(eventId: string): Promise<string> {
    const data: CreateEventShareRequest = {}
    const res = await this.request.post(getCreateEventShareUrl(eventId), {
      data,
    })
    await expectStatus(res, 201)
    const body: CreateEventShareResponse = await res.json()
    return body.url
  }

  /** 共有リンクのパス（`/share/{token}`）から、共有された予定を自分のカレンダーに追加する */
  async subscribe(sharePath: string): Promise<void> {
    const token = sharePath.replace(/^\/share\//, '')
    const res = await this.request.post(getSubscribeToShareUrl(token))
    await expectStatus(res, 201)
  }

  async createTag(input: CreateTagRequest): Promise<TagResponse> {
    const res = await this.request.post(getCreateTagUrl(), { data: input })
    await expectStatus(res, 201)
    return res.json()
  }

  /** 画面表示に頼らず、サーバーに登録されているタグを確かめる */
  async listTags(): Promise<TagResponse[]> {
    const res = await this.request.get(getListTagsUrl())
    await expectStatus(res, 200)
    const body: TagsListResponse = await res.json()
    return body.tags
  }

  /** 画面表示に頼らず、サーバーに登録されている予定を確かめる */
  async listEvents(range: GetEventsParams): Promise<EventResponse[]> {
    const res = await this.request.get(getGetEventsUrl(range))
    await expectStatus(res, 200)
    const body: EventsListResponse = await res.json()
    return body.events
  }
}

export async function expectStatus(
  res: APIResponse,
  status: number,
): Promise<void> {
  // 失敗時にレスポンス本文を出し、バリデーションエラー等の原因を追えるようにする
  expect(res.status(), `${res.url()}: ${await res.text()}`).toBe(status)
}
