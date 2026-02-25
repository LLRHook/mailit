import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import SettingsPage from '../page'

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

describe('SettingsPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<SettingsPage />)
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<SettingsPage />)
    expect(screen.getByText('Manage your account and project settings')).toBeInTheDocument()
  })

  it('renders tab triggers', () => {
    renderWithProviders(<SettingsPage />)
    expect(screen.getByText('Usage')).toBeInTheDocument()
    expect(screen.getByText('Team')).toBeInTheDocument()
    expect(screen.getByText('SMTP')).toBeInTheDocument()
    expect(screen.getByText('Integrations')).toBeInTheDocument()
  })
})
