import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatusBadge } from '../status-badge'

describe('StatusBadge', () => {
  it('renders the status text', () => {
    render(<StatusBadge status="delivered" />)
    expect(screen.getByText('delivered')).toBeInTheDocument()
  })

  it('applies green variant for delivered status', () => {
    render(<StatusBadge status="delivered" />)
    const badge = screen.getByText('delivered')
    expect(badge).toHaveClass('text-emerald-400')
  })

  it('applies red variant for failed status', () => {
    render(<StatusBadge status="failed" />)
    const badge = screen.getByText('failed')
    expect(badge).toHaveClass('text-red-400')
  })

  it('applies yellow variant for pending status', () => {
    render(<StatusBadge status="pending" />)
    const badge = screen.getByText('pending')
    expect(badge).toHaveClass('text-yellow-400')
  })

  it('applies blue variant for opened status', () => {
    render(<StatusBadge status="opened" />)
    const badge = screen.getByText('opened')
    expect(badge).toHaveClass('text-blue-400')
  })

  it('applies fallback variant for unknown status', () => {
    render(<StatusBadge status="unknown" />)
    const badge = screen.getByText('unknown')
    expect(badge).toHaveClass('text-zinc-400')
  })
})
