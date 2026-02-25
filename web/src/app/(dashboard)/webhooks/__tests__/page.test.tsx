import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import WebhooksPage from '../page'

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

describe('WebhooksPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<WebhooksPage />)
    expect(screen.getByText('Webhooks')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<WebhooksPage />)
    expect(screen.getByText('Receive real-time notifications for email events')).toBeInTheDocument()
  })

  it('renders the add webhook button', () => {
    renderWithProviders(<WebhooksPage />)
    expect(screen.getByText('Add Webhook')).toBeInTheDocument()
  })
})
