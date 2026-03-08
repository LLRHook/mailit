import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import NewWebhookPage from '../page'

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

describe('NewWebhookPage', () => {
  it('renders the page title', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('New Webhook')).toBeInTheDocument()
  })

  it('renders the Webhook Configuration card', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('Webhook Configuration')).toBeInTheDocument()
  })

  it('renders the Endpoint URL input', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByLabelText('Endpoint URL')).toBeInTheDocument()
  })

  it('renders all six event checkboxes', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('Email Sent')).toBeInTheDocument()
    expect(screen.getByText('Email Delivered')).toBeInTheDocument()
    expect(screen.getByText('Email Bounced')).toBeInTheDocument()
    expect(screen.getByText('Email Opened')).toBeInTheDocument()
    expect(screen.getByText('Email Clicked')).toBeInTheDocument()
    expect(screen.getByText('Email Complained')).toBeInTheDocument()
  })

  it('renders the Create Webhook button', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('Create Webhook')).toBeInTheDocument()
  })
})
