import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('App', () => {
  it('renders design system showcase headings', () => {
    render(<App />)
    expect(
      screen.getByRole('heading', { name: /typography/i }),
    ).toBeInTheDocument()
  })
})
