import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import EmailDetailPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        data: {
          id: 'e-1',
          from: 'noreply@example.com',
          to: 'user@example.com',
          subject: 'Test Email Subject',
          status: 'delivered',
          html: '<p>Hello</p>',
          text: 'Hello',
          created_at: '2026-01-01T00:00:00Z',
          sent_at: '2026-01-01T00:01:00Z',
          delivered_at: '2026-01-01T00:02:00Z',
          events: [],
        },
      },
    }),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

describe('EmailDetailPage', () => {
  it('renders the email subject after loading', async () => {
    renderWithProviders(<EmailDetailPage />)
    expect(await screen.findByRole('heading', { name: 'Test Email Subject' })).toBeInTheDocument()
  })

  it('renders the Email Details card', async () => {
    renderWithProviders(<EmailDetailPage />)
    expect(await screen.findByText('Email Details')).toBeInTheDocument()
  })

  it('renders Preview, Source, and Events tabs', async () => {
    renderWithProviders(<EmailDetailPage />)
    expect(await screen.findByText('Preview')).toBeInTheDocument()
    expect(screen.getByText('Source')).toBeInTheDocument()
    expect(screen.getByText('Events')).toBeInTheDocument()
  })
})
