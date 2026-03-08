import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import BroadcastDetailPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        data: {
          id: '1',
          name: 'Welcome Campaign',
          from: 'hello@example.com',
          subject: 'Welcome!',
          audience_id: 'aud-1',
          audience_name: 'All Contacts',
          status: 'sent',
          recipients: 100,
          sent: 95,
          delivered: 90,
          opened: 50,
          html: '<p>Hello</p>',
          text: 'Hello',
          created_at: '2026-01-01T00:00:00Z',
          sent_at: '2026-01-01T01:00:00Z',
          completed_at: null,
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

describe('BroadcastDetailPage', () => {
  it('renders the broadcast name after loading', async () => {
    renderWithProviders(<BroadcastDetailPage />)
    expect(await screen.findByText('Welcome Campaign')).toBeInTheDocument()
  })

  it('renders stat cards for Recipients, Sent, and Delivered', async () => {
    renderWithProviders(<BroadcastDetailPage />)
    expect(await screen.findByText('Recipients')).toBeInTheDocument()
    expect(screen.getByText('Sent')).toBeInTheDocument()
    expect(screen.getByText('Delivered')).toBeInTheDocument()
  })

  it('renders Broadcast Details and Preview cards', async () => {
    renderWithProviders(<BroadcastDetailPage />)
    expect(await screen.findByText('Broadcast Details')).toBeInTheDocument()
    expect(screen.getByText('Preview')).toBeInTheDocument()
  })
})
