import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import RegisterPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

describe('RegisterPage', () => {
  it('renders the create account heading', () => {
    render(<RegisterPage />)
    expect(screen.getByText('Create your account')).toBeInTheDocument()
  })

  it('renders all registration fields', () => {
    render(<RegisterPage />)
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByLabelText('Team name')).toBeInTheDocument()
  })

  it('renders the create account button', () => {
    render(<RegisterPage />)
    expect(screen.getByRole('button', { name: 'Create account' })).toBeInTheDocument()
  })

  it('renders link to login page', () => {
    render(<RegisterPage />)
    expect(screen.getByText('Sign in')).toBeInTheDocument()
  })
})
