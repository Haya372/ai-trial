import { act, renderHook } from '@testing-library/react'
import i18n from 'i18next'
import { useForm } from 'react-hook-form'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useRevalidateOnLanguageChange } from './useRevalidateOnLanguageChange'

function useTestForm() {
  const form = useForm<{ name: string }>({ defaultValues: { name: '' } })
  useRevalidateOnLanguageChange(form)
  return form
}

describe('useRevalidateOnLanguageChange', () => {
  afterEach(async () => {
    await i18n.changeLanguage('ja')
  })

  it('エラー表示中のフィールドがない場合、言語切替時に trigger を呼ばない', async () => {
    const { result } = renderHook(() => useTestForm())
    const triggerSpy = vi.spyOn(result.current, 'trigger')

    await act(async () => {
      await i18n.changeLanguage('en')
    })

    expect(triggerSpy).not.toHaveBeenCalled()
  })

  it('エラー表示中のフィールドがある場合、言語切替時に trigger を呼び直す', async () => {
    const { result } = renderHook(() => useTestForm())

    await act(async () => {
      result.current.setError('name', { type: 'manual', message: 'required' })
    })

    const triggerSpy = vi.spyOn(result.current, 'trigger')

    await act(async () => {
      await i18n.changeLanguage('en')
    })

    expect(triggerSpy).toHaveBeenCalledWith(['name'])
  })
})
