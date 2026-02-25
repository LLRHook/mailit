import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { PageHeader } from '../page-header'

describe('PageHeader', () => {
  it('renders the title', () => {
    render(<PageHeader title="Test Page" />)
    expect(screen.getByText('Test Page')).toBeInTheDocument()
  })

  it('renders the description when provided', () => {
    render(<PageHeader title="Title" description="A helpful description" />)
    expect(screen.getByText('A helpful description')).toBeInTheDocument()
  })

  it('does not render description when not provided', () => {
    const { container } = render(<PageHeader title="Title" />)
    const desc = container.querySelector('.text-muted-foreground')
    expect(desc).toBeNull()
  })

  it('renders action button when actionLabel and onAction are provided', () => {
    const onAction = vi.fn()
    render(<PageHeader title="Title" actionLabel="Create" onAction={onAction} />)
    const button = screen.getByText('Create')
    expect(button).toBeInTheDocument()
    fireEvent.click(button)
    expect(onAction).toHaveBeenCalledOnce()
  })

  it('does not render action button when only actionLabel is provided', () => {
    render(<PageHeader title="Title" actionLabel="Create" />)
    expect(screen.queryByText('Create')).not.toBeInTheDocument()
  })

  it('renders children', () => {
    render(
      <PageHeader title="Title">
        <span data-testid="child">Custom Child</span>
      </PageHeader>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })
})
