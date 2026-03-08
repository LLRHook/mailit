import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import EditTemplatePage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        data: {
          id: 't-1',
          name: 'Welcome Email',
          description: 'Onboarding template',
          subject: 'Welcome!',
          html: '<p>Hi</p>',
          text: 'Hi',
          published: false,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-02T00:00:00Z',
        },
      },
    }),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

describe('EditTemplatePage', () => {
  it('renders the page title after loading', async () => {
    renderWithProviders(<EditTemplatePage />)
    expect(await screen.findByText('Edit Template')).toBeInTheDocument()
  })

  it('renders Template Details card and form inputs', async () => {
    renderWithProviders(<EditTemplatePage />)
    expect(await screen.findByText('Template Details')).toBeInTheDocument()
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Subject')).toBeInTheDocument()
    expect(screen.getByLabelText('HTML Body')).toBeInTheDocument()
    expect(screen.getByLabelText('Plain Text Body')).toBeInTheDocument()
  })

  it('renders Save Changes and Publish buttons', async () => {
    renderWithProviders(<EditTemplatePage />)
    expect(await screen.findByText('Save Changes')).toBeInTheDocument()
    expect(screen.getByText('Publish')).toBeInTheDocument()
  })
})
