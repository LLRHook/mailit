import { describe, it, expect } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { CopyButton } from '../copy-button'

describe('CopyButton', () => {
  it('renders a button', () => {
    render(<CopyButton value="test-value" />)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('copies value to clipboard on click', async () => {
    render(<CopyButton value="copy-me" />)
    const button = screen.getByRole('button')
    await act(async () => {
      button.click()
    })
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('copy-me')
  })

  it('accepts optional className', () => {
    render(<CopyButton value="test" className="extra-class" />)
    const button = screen.getByRole('button')
    expect(button).toHaveClass('extra-class')
  })
})
