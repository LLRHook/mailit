import { describe, it, expect, vi } from 'vitest'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import BroadcastsPage from '../page'

const mockPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const mockGet = vi.fn()
vi.mock('@/lib/api', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn().mockResolvedValue({ data: {} }),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

const broadcasts = [
  {
    id: '1',
    name: 'Welcome Campaign',
    audience_id: 'a1',
    audience_name: 'All Users',
    status: 'sent',
    recipients: 100,
    sent: 95,
    created_at: '2026-01-15T10:00:00Z',
  },
  {
    id: '2',
    name: 'Product Update',
    audience_id: 'a2',
    audience_name: 'Beta Users',
    status: 'draft',
    recipients: 50,
    sent: 0,
    created_at: '2026-02-20T10:00:00Z',
  },
]

describe('BroadcastsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGet.mockResolvedValue({ data: { data: broadcasts } })
  })

  it('renders the search input for filtering broadcasts', async () => {
    renderWithProviders(<BroadcastsPage />)
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search broadcasts...')).toBeInTheDocument()
    })
  })

  it('filters broadcasts by name when searching', async () => {
    renderWithProviders(<BroadcastsPage />)
    await waitFor(() => {
      expect(screen.getByText('Welcome Campaign')).toBeInTheDocument()
    })
    const searchInput = screen.getByPlaceholderText('Search broadcasts...')
    fireEvent.change(searchInput, { target: { value: 'Product' } })
    expect(screen.getByText('Product Update')).toBeInTheDocument()
    expect(screen.queryByText('Welcome Campaign')).not.toBeInTheDocument()
  })

  it('renders the page header and new broadcast button', async () => {
    renderWithProviders(<BroadcastsPage />)
    expect(screen.getByText('Broadcasts')).toBeInTheDocument()
    expect(screen.getByText('New Broadcast')).toBeInTheDocument()
  })
})
