import { describe, it, expect, vi, beforeAll } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CommandPalette } from '../command-palette'

beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn()
})

const mockPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => '/',
}))

describe('CommandPalette', () => {
  beforeEach(() => {
    mockPush.mockClear()
  })

  it('opens on Ctrl+K', async () => {
    render(<CommandPalette />)
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search pages...')).toBeInTheDocument()
    })
  })

  it('shows all navigation pages when open', async () => {
    render(<CommandPalette />)
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })
    await waitFor(() => {
      expect(screen.getByText('Emails')).toBeInTheDocument()
    })
    expect(screen.getByText('Broadcasts')).toBeInTheDocument()
    expect(screen.getByText('Templates')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('navigates when an item is selected', async () => {
    render(<CommandPalette />)
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })
    await waitFor(() => {
      expect(screen.getByText('Emails')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Emails'))
    expect(mockPush).toHaveBeenCalledWith('/emails')
  })

  it('closes on Ctrl+K toggle', async () => {
    render(<CommandPalette />)
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search pages...')).toBeInTheDocument()
    })
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true })
    await waitFor(() => {
      expect(screen.queryByPlaceholderText('Search pages...')).not.toBeInTheDocument()
    })
  })
})
