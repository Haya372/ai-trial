import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import ErrorFallback from './ErrorFallback'

describe('ErrorFallback', () => {
  it('予期しないエラー発生時のフォールバックUIを表示する', () => {
    render(<ErrorFallback reset={vi.fn()} />)
    expect(
      screen.getByText('予期しないエラーが発生しました'),
    ).toBeInTheDocument()
  })

  it('再試行ボタンをクリックすると reset が呼ばれる', () => {
    const reset = vi.fn()
    render(<ErrorFallback reset={reset} />)
    fireEvent.click(screen.getByRole('button', { name: '再試行' }))
    expect(reset).toHaveBeenCalled()
  })

  it('console.error を呼ばない', () => {
    const consoleErrorSpy = vi
      .spyOn(console, 'error')
      .mockImplementation(() => {})
    render(<ErrorFallback reset={vi.fn()} />)
    expect(consoleErrorSpy).not.toHaveBeenCalled()
    consoleErrorSpy.mockRestore()
  })
})
