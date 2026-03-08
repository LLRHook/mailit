import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Mail } from 'lucide-react'
import { StatCard } from '../stat-card'

describe('StatCard', () => {
  it('renders title and value', () => {
    render(<StatCard title="Emails Sent" value={1234} icon={Mail} />)
    expect(screen.getByText('Emails Sent')).toBeInTheDocument()
    expect(screen.getByText('1234')).toBeInTheDocument()
  })

  it('renders string value', () => {
    render(<StatCard title="Rate" value="99.5%" icon={Mail} />)
    expect(screen.getByText('99.5%')).toBeInTheDocument()
  })

  it('renders change with up trend', () => {
    render(<StatCard title="Sent" value={100} icon={Mail} change="+12%" trend="up" />)
    const changeEl = screen.getByText((_, el) => el?.textContent === '↑ +12%')
    expect(changeEl).toHaveClass('text-emerald-400')
  })

  it('renders change with down trend', () => {
    render(<StatCard title="Sent" value={100} icon={Mail} change="-5%" trend="down" />)
    const changeEl = screen.getByText((_, el) => el?.textContent === '↓ -5%')
    expect(changeEl).toHaveClass('text-red-400')
  })

  it('does not render change when not provided', () => {
    const { container } = render(<StatCard title="Sent" value={100} icon={Mail} />)
    expect(container.querySelector('.text-xs')).toBeNull()
  })
})
