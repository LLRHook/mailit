import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import DashboardHome from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { totals: { sent: 42, delivery_rate: 98.5, open_rate: 35.2, bounce_rate: 1.5 }, emails_sent_today: 10, emails_sent_month: 300, contacts: 150, domains: 2, api_keys: 3, webhooks: 1 } }),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

describe('DashboardHome', () => {
  it('renders the page header', () => {
    renderWithProviders(<DashboardHome />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<DashboardHome />)
    expect(screen.getByText('Overview of your email sending activity')).toBeInTheDocument()
  })

  it('renders stat cards', () => {
    renderWithProviders(<DashboardHome />)
    expect(screen.getByText('Emails Sent (7d)')).toBeInTheDocument()
    expect(screen.getByText('Delivery Rate')).toBeInTheDocument()
    expect(screen.getByText('Open Rate')).toBeInTheDocument()
    expect(screen.getByText('Bounce Rate')).toBeInTheDocument()
  })

  it('renders usage sections', () => {
    renderWithProviders(<DashboardHome />)
    expect(screen.getByText('Today')).toBeInTheDocument()
    expect(screen.getByText('This Month')).toBeInTheDocument()
    expect(screen.getByText('Contacts')).toBeInTheDocument()
  })

  it('renders quick links', () => {
    renderWithProviders(<DashboardHome />)
    expect(screen.getByText('Quick Links')).toBeInTheDocument()
    expect(screen.getByText('Domains')).toBeInTheDocument()
    expect(screen.getByText('API Keys')).toBeInTheDocument()
    expect(screen.getByText('Audience')).toBeInTheDocument()
    expect(screen.getByText('Webhooks')).toBeInTheDocument()
  })
})
