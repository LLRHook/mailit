import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import LogsPage from '../page'

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

describe('LogsPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<LogsPage />)
    expect(screen.getByText('API Logs')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<LogsPage />)
    expect(screen.getByText('View API request logs')).toBeInTheDocument()
  })
})
