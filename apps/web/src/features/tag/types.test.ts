import i18n from 'i18next'
import { describe, expect, it } from 'vitest'
import { createTagFormSchema } from './types'

const schema = createTagFormSchema(i18n.getFixedT('ja', 'tag'))

function nameError(name: string): string | undefined {
  const result = schema.safeParse({ name, color: 'red' })
  return result.success ? undefined : result.error.issues[0]?.message
}

describe('createTagFormSchema', () => {
  it('前後の空白を除去したタグ名を返す', () => {
    expect(schema.parse({ name: '  仕事  ', color: 'blue' })).toEqual({
      name: '仕事',
      color: 'blue',
    })
  })

  it('空のタグ名はエラーになる', () => {
    expect(nameError('')).toBe('タグ名を入力してください')
  })

  it('空白だけのタグ名はエラーになる', () => {
    expect(nameError('   ')).toBe('タグ名を入力してください')
  })

  it('20文字のタグ名は通る', () => {
    expect(nameError('あ'.repeat(20))).toBeUndefined()
  })

  it('21文字のタグ名はエラーになる', () => {
    expect(nameError('あ'.repeat(21))).toBe(
      'タグ名は20文字以内で入力してください',
    )
  })

  it('サロゲートペアの文字も1文字として数える', () => {
    expect(nameError('😀'.repeat(20))).toBeUndefined()
  })

  it('12色以外の色はエラーになる', () => {
    expect(schema.safeParse({ name: '仕事', color: 'black' }).success).toBe(
      false,
    )
  })
})
