import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import TemplatesPage from '../page'

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

describe('TemplatesPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<TemplatesPage />)
    expect(screen.getByText('Templates')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<TemplatesPage />)
    expect(screen.getByText('Manage reusable email templates')).toBeInTheDocument()
  })

  it('renders the new template button', () => {
    renderWithProviders(<TemplatesPage />)
    expect(screen.getByText('New Template')).toBeInTheDocument()
  })
})
