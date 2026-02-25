import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import ApiKeysPage from '../page'

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

describe('ApiKeysPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<ApiKeysPage />)
    expect(screen.getByText('API Keys')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<ApiKeysPage />)
    expect(screen.getByText('Manage your API keys for programmatic access')).toBeInTheDocument()
  })

  it('renders the create API key button', () => {
    renderWithProviders(<ApiKeysPage />)
    expect(screen.getByText('Create API Key')).toBeInTheDocument()
  })
})
