import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import AudiencePage from '../page'

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

describe('AudiencePage', () => {
  it('renders the page header', () => {
    renderWithProviders(<AudiencePage />)
    expect(screen.getByText('Audience')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<AudiencePage />)
    expect(screen.getByText('Manage contacts, properties, segments, and topics')).toBeInTheDocument()
  })

  it('renders all four tab triggers', () => {
    renderWithProviders(<AudiencePage />)
    expect(screen.getByText('Contacts')).toBeInTheDocument()
    expect(screen.getByText('Properties')).toBeInTheDocument()
    expect(screen.getByText('Segments')).toBeInTheDocument()
    expect(screen.getByText('Topics')).toBeInTheDocument()
  })
})
