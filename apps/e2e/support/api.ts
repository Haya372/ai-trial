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

export type CreatedEvent = CreateEventInput & { id: string }

/** ログイン済みユーザーのセッション Cookie を持つ APIRequestContext で、UI を経ずにテストデータを用意する */
export class UserApi {
  constructor(private readonly request: APIRequestContext) {}

  async createEvent(input: CreateEventInput): Promise<CreatedEvent> {
    const res = await this.request.post('/api/events', { data: input })
    await expectStatus(res, 201)
    return res.json()
  }
}

export async function expectStatus(res: APIResponse, status: number): Promise<void> {
  // 失敗時にレスポンス本文を出し、バリデーションエラー等の原因を追えるようにする
  expect(res.status(), `${res.url()}: ${await res.text()}`).toBe(status)
}
