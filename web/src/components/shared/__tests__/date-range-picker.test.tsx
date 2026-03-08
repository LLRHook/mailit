import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { DateRangePicker } from '../date-range-picker'

describe('DateRangePicker', () => {
  it('renders placeholder when no value', () => {
    render(<DateRangePicker />)
    expect(screen.getByText('Select date range')).toBeInTheDocument()
  })

  it('renders formatted date range when both from and to provided', () => {
    const value = { from: new Date(2026, 0, 1), to: new Date(2026, 0, 31) }
    render(<DateRangePicker value={value} />)
    expect(screen.getByText(/Jan 1, 2026/)).toBeInTheDocument()
    expect(screen.getByText(/Jan 31, 2026/)).toBeInTheDocument()
  })

  it('renders only from date when to is not set', () => {
    const value = { from: new Date(2026, 2, 15) }
    render(<DateRangePicker value={value} />)
    expect(screen.getByText('Mar 15, 2026')).toBeInTheDocument()
  })

  it('opens popover on button click', () => {
    render(<DateRangePicker />)
    fireEvent.click(screen.getByText('Select date range'))
    // Calendar renders month navigation when open
    expect(document.querySelector('[role="grid"]')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<DateRangePicker className="custom-class" />)
    const button = screen.getByText('Select date range').closest('button')
    expect(button).toHaveClass('custom-class')
  })
})
