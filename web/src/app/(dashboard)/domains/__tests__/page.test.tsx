import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import DomainsPage from '../page'

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

describe('DomainsPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<DomainsPage />)
    expect(screen.getByText('Domains')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<DomainsPage />)
    expect(screen.getByText('Manage your sending domains')).toBeInTheDocument()
  })

  it('renders the add domain button', () => {
    renderWithProviders(<DomainsPage />)
    expect(screen.getByText('Add Domain')).toBeInTheDocument()
  })
})
