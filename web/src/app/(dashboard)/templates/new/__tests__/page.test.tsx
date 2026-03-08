import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import NewTemplatePage from '../page'

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

describe('NewTemplatePage', () => {
  it('renders the page title', () => {
    renderWithProviders(<NewTemplatePage />)
    expect(screen.getByText('New Template')).toBeInTheDocument()
  })

  it('renders the Template Details card', () => {
    renderWithProviders(<NewTemplatePage />)
    expect(screen.getByText('Template Details')).toBeInTheDocument()
  })

  it('renders Name, Subject, HTML Body, and Plain Text Body inputs', () => {
    renderWithProviders(<NewTemplatePage />)
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Subject')).toBeInTheDocument()
    expect(screen.getByLabelText('HTML Body')).toBeInTheDocument()
    expect(screen.getByLabelText('Plain Text Body')).toBeInTheDocument()
  })

  it('renders the Create Template button', () => {
    renderWithProviders(<NewTemplatePage />)
    expect(screen.getByText('Create Template')).toBeInTheDocument()
  })
})
