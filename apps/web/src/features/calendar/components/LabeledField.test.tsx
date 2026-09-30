import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import LabeledField from './LabeledField'

describe('LabeledField', () => {
  describe('表示', () => {
    it('ラベルをコロン区切りで表示する', () => {
      render(<LabeledField label="開始">値</LabeledField>)
      expect(screen.getByText('開始:', { exact: false })).toBeInTheDocument()
    })

    it('children が文字列の場合そのまま表示する', () => {
      render(<LabeledField label="メモ">設計方針をレビューする</LabeledField>)
      expect(screen.getByText('設計方針をレビューする')).toBeInTheDocument()
    })

    it('children が要素の場合そのまま表示する', () => {
      render(
        <LabeledField label="URL">
          <a href="https://example.com/agenda">https://example.com/agenda</a>
        </LabeledField>,
      )
      expect(
        screen.getByRole('link', { name: 'https://example.com/agenda' }),
      ).toHaveAttribute('href', 'https://example.com/agenda')
    })
  })
})
