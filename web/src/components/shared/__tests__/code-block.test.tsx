import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { CodeBlock } from '../code-block'

describe('CodeBlock', () => {
  it('renders the code content', () => {
    render(<CodeBlock code="console.log('hello')" />)
    expect(screen.getByText("console.log('hello')")).toBeInTheDocument()
  })

  it('renders code inside a pre element', () => {
    render(<CodeBlock code="const x = 1" />)
    const pre = screen.getByText('const x = 1')
    expect(pre.tagName).toBe('PRE')
  })

  it('includes a copy button', () => {
    render(<CodeBlock code="some code" />)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })
})
