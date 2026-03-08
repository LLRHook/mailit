import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import NewWebhookPage from '../new/page'

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

  it('renders the endpoint URL input', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByLabelText('Endpoint URL')).toBeInTheDocument()
  })

  it('renders all event types with labels', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('Email Sent')).toBeInTheDocument()
    expect(screen.getByText('Email Delivered')).toBeInTheDocument()
    expect(screen.getByText('Email Bounced')).toBeInTheDocument()
    expect(screen.getByText('Email Opened')).toBeInTheDocument()
    expect(screen.getByText('Email Clicked')).toBeInTheDocument()
    expect(screen.getByText('Email Complained')).toBeInTheDocument()
  })

  it('renders event descriptions', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByText('Triggered when an email is accepted for delivery')).toBeInTheDocument()
    expect(screen.getByText("Triggered when the recipient's mail server confirms receipt")).toBeInTheDocument()
    expect(screen.getByText('Triggered when the recipient marks the email as spam')).toBeInTheDocument()
  })

  it('renders create button as disabled initially', () => {
    renderWithProviders(<NewWebhookPage />)
    expect(screen.getByRole('button', { name: 'Create Webhook' })).toBeDisabled()
  })
})
