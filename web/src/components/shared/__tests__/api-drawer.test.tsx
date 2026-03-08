import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ApiDrawer } from '../api-drawer'

const examples = [
  {
    title: 'Send an email',
    method: 'POST',
    endpoint: '/v1/emails',
    body: '{"to":"user@example.com","subject":"Hello","html":"<p>Hi</p>"}',
  },
]

describe('ApiDrawer', () => {
  it('renders the trigger button', () => {
    render(<ApiDrawer examples={examples} />)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('shows API Reference title when opened', () => {
    render(<ApiDrawer examples={examples} />)
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText('API Reference')).toBeInTheDocument()
  })

  it('shows method and endpoint', () => {
    render(<ApiDrawer examples={examples} />)
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/v1/emails')).toBeInTheDocument()
  })

  it('renders language tabs', () => {
    render(<ApiDrawer examples={examples} />)
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByRole('tab', { name: 'cURL' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Node.js' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Python' })).toBeInTheDocument()
  })

  it('shows curl code by default', () => {
    render(<ApiDrawer examples={examples} />)
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText(/curl -X POST/)).toBeInTheDocument()
  })
})
