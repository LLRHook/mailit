import { describe, it, expect, vi } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import LogsPage from '../page'

const mockGet = vi.fn().mockResolvedValue({ data: { data: [] } })

vi.mock('@/lib/api', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
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
  beforeEach(() => {
    mockGet.mockResolvedValue({ data: { data: [] } })
  })

  it('renders the page header', () => {
    renderWithProviders(<LogsPage />)
    expect(screen.getByText('API Logs')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<LogsPage />)
    expect(screen.getByText('View API request logs')).toBeInTheDocument()
  })

  it('shows error state when API fails', async () => {
    mockGet.mockRejectedValue(new Error('Network error'))
    renderWithProviders(<LogsPage />)
    await waitFor(() => {
      expect(screen.getByText('Failed to load logs')).toBeInTheDocument()
    })
    expect(screen.getByText('Network error')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
  })

  it('calls refetch when Retry button is clicked', async () => {
    mockGet.mockRejectedValue(new Error('Network error'))
    renderWithProviders(<LogsPage />)
    await waitFor(() => {
      expect(screen.getByText('Failed to load logs')).toBeInTheDocument()
    })
    const callsBefore = mockGet.mock.calls.length
    fireEvent.click(screen.getByRole('button', { name: /retry/i }))
    await waitFor(() => {
      expect(mockGet.mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })
})
