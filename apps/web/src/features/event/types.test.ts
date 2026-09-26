import { describe, expect, it } from 'vitest'
import { eventFormSchema, quickRegistrationSchema } from './types'

const validData = {
  title: '定例ミーティング',
  startAt: '2026-09-22T10:00',
  endAt: '2026-09-22T11:00',
  description: '',
  location: '',
  url: '',
}

describe('eventFormSchema', () => {
  it('必須項目が揃っていれば成功する', () => {
    const result = eventFormSchema.safeParse(validData)
    expect(result.success).toBe(true)
  })

  it('タイトルが空の場合はエラーになる', () => {
    const result = eventFormSchema.safeParse({ ...validData, title: '' })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path[0] === 'title')).toBe(true)
    }
  })

  it('終了日時が開始日時と同時刻の場合はエラーになる', () => {
    const result = eventFormSchema.safeParse({
      ...validData,
      startAt: '2026-09-22T10:00',
      endAt: '2026-09-22T10:00',
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path[0] === 'endAt')).toBe(true)
    }
  })

  it('終了日時が開始日時より前の場合はエラーになる', () => {
    const result = eventFormSchema.safeParse({
      ...validData,
      startAt: '2026-09-22T10:00',
      endAt: '2026-09-22T09:00',
    })
    expect(result.success).toBe(false)
  })

  it('終了日時が開始日時より後であれば成功する', () => {
    const result = eventFormSchema.safeParse({
      ...validData,
      startAt: '2026-09-22T10:00',
      endAt: '2026-09-22T10:01',
    })
    expect(result.success).toBe(true)
  })

  it('メモ・場所・URLは任意項目で省略できる', () => {
    const result = eventFormSchema.safeParse({
      title: '定例ミーティング',
      startAt: '2026-09-22T10:00',
      endAt: '2026-09-22T11:00',
    })
    expect(result.success).toBe(true)
  })
})

describe('quickRegistrationSchema', () => {
  it('タイトルがあれば成功する', () => {
    const result = quickRegistrationSchema.safeParse({ title: '仮予定' })
    expect(result.success).toBe(true)
  })

  it('タイトルが空の場合はエラーになる', () => {
    const result = quickRegistrationSchema.safeParse({ title: '' })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe('タイトルを入力してください')
    }
  })
})
