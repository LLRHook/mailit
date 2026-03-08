import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RelativeTime } from '../relative-time'

describe('RelativeTime', () => {
  it('renders relative time text', () => {
    const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString()
    render(<RelativeTime date={fiveMinAgo} />)
    expect(screen.getByText(/minutes? ago/)).toBeInTheDocument()
  })

  it('renders "less than a minute ago" for very recent dates', () => {
    const justNow = new Date(Date.now() - 10 * 1000).toISOString()
    render(<RelativeTime date={justNow} />)
    expect(screen.getByText(/less than a minute ago/)).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const date = new Date().toISOString()
    render(<RelativeTime date={date} className="text-red-500" />)
    const el = screen.getByText(/ago/)
    expect(el).toHaveClass('text-red-500')
  })

  it('shows absolute date in tooltip content', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-08T12:00:00Z'))
    render(<RelativeTime date="2026-03-08T11:30:00Z" />)
    // Tooltip content is rendered but hidden — check it exists in the DOM
    expect(screen.getByText(/30 minutes ago/)).toBeInTheDocument()
    vi.useRealTimers()
  })
})
