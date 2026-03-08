import { describe, it, expect, vi } from 'vitest'
import { screen, fireEvent } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import EmailsPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { data: [] } }),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

describe('EmailsPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<EmailsPage />)
    expect(screen.getByText('Emails')).toBeInTheDocument()
  })

  it('renders the stat cards', () => {
    renderWithProviders(<EmailsPage />)
    expect(screen.getByText('Total Sent')).toBeInTheDocument()
    expect(screen.getByText('Delivered')).toBeInTheDocument()
    expect(screen.getByText('Bounced')).toBeInTheDocument()
    expect(screen.getByText('Open Rate')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<EmailsPage />)
    expect(screen.getByText('View and manage sent emails')).toBeInTheDocument()
  })

  it('renders auto-refresh toggle defaulting to Pause', () => {
    renderWithProviders(<EmailsPage />)
    expect(screen.getByRole('button', { name: /pause/i })).toBeInTheDocument()
  })

  it('toggles auto-refresh to Resume when clicked', () => {
    renderWithProviders(<EmailsPage />)
    fireEvent.click(screen.getByRole('button', { name: /pause/i }))
    expect(screen.getByRole('button', { name: /resume/i })).toBeInTheDocument()
  })

  it('renders manual Refresh button', () => {
    renderWithProviders(<EmailsPage />)
    expect(screen.getByRole('button', { name: /refresh/i })).toBeInTheDocument()
  })
})
