import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import NewBroadcastPage from '../page'

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

describe('NewBroadcastPage', () => {
  it('renders the page title', () => {
    renderWithProviders(<NewBroadcastPage />)
    expect(screen.getByText('New Broadcast')).toBeInTheDocument()
  })

  it('renders the broadcast details card', () => {
    renderWithProviders(<NewBroadcastPage />)
    expect(screen.getByText('Broadcast Details')).toBeInTheDocument()
  })

  it('renders Save Draft and Send buttons', () => {
    renderWithProviders(<NewBroadcastPage />)
    expect(screen.getByText('Save Draft')).toBeInTheDocument()
    expect(screen.getByText('Send Now')).toBeInTheDocument()
  })
})
