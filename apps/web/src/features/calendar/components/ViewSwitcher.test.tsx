import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import ViewSwitcher from './ViewSwitcher'

describe('ViewSwitcher', () => {
  describe('表示', () => {
    it('「月」「週」のタブを表示する', () => {
      render(<ViewSwitcher view="month" onViewChange={vi.fn()} />)
      expect(screen.getByRole('tab', { name: /月/ })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: /週/ })).toBeInTheDocument()
    })

    it('view が "month" のとき「月」タブが選択状態になる', () => {
      render(<ViewSwitcher view="month" onViewChange={vi.fn()} />)
      const monthTab = screen.getByRole('tab', { name: /月/ })
      // aria-selected で選択状態を確認
      expect(monthTab).toHaveAttribute('aria-selected', 'true')
    })

    it('view が "week" のとき「週」タブが選択状態になる', () => {
      render(<ViewSwitcher view="week" onViewChange={vi.fn()} />)
      const weekTab = screen.getByRole('tab', { name: /週/ })
      expect(weekTab).toHaveAttribute('aria-selected', 'true')
    })
  })

  describe('操作', () => {
    it('「週」タブをクリックすると onViewChange("week") が呼ばれる', () => {
      const onViewChange = vi.fn()
      render(<ViewSwitcher view="month" onViewChange={onViewChange} />)
      fireEvent.click(screen.getByRole('tab', { name: /週/ }))
      expect(onViewChange).toHaveBeenCalledWith('week')
    })

    it('「月」タブをクリックすると onViewChange("month") が呼ばれる', () => {
      const onViewChange = vi.fn()
      render(<ViewSwitcher view="week" onViewChange={onViewChange} />)
      fireEvent.click(screen.getByRole('tab', { name: /月/ }))
      expect(onViewChange).toHaveBeenCalledWith('month')
    })
  })
})
