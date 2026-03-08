import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import DomainDetailPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        id: 'd-1',
        name: 'example.com',
        status: 'verified',
        dns_records: [
          { type: 'TXT', name: '_mailit.example.com', value: 'v=mailit1', status: 'verified', ttl: '3600' },
        ],
        created_at: '2026-01-01T00:00:00Z',
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

describe('DomainDetailPage', () => {
  it('renders the domain name after loading', async () => {
    renderWithProviders(<DomainDetailPage />)
    expect(await screen.findByText('example.com')).toBeInTheDocument()
  })

  it('renders the DNS Records card', async () => {
    renderWithProviders(<DomainDetailPage />)
    expect(await screen.findByText('DNS Records')).toBeInTheDocument()
  })

  it('renders the Verify Domain button', async () => {
    renderWithProviders(<DomainDetailPage />)
    expect(await screen.findByText('Verify Domain')).toBeInTheDocument()
  })
})
