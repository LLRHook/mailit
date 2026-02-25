import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Mail } from 'lucide-react'
import { EmptyState } from '../empty-state'

describe('EmptyState', () => {
  it('renders the title', () => {
    render(
      <EmptyState icon={Mail} title="Nothing here" description="No items found." />
    )
    expect(screen.getByText('Nothing here')).toBeInTheDocument()
  })

  it('renders the description', () => {
    render(
      <EmptyState icon={Mail} title="Empty" description="Try adding something." />
    )
    expect(screen.getByText('Try adding something.')).toBeInTheDocument()
  })

  it('renders action button when actionLabel and onAction are provided', () => {
    const onAction = vi.fn()
    render(
      <EmptyState
        icon={Mail}
        title="No items"
        description="None found."
        actionLabel="Add Item"
        onAction={onAction}
      />
    )
    const button = screen.getByText('Add Item')
    expect(button).toBeInTheDocument()
    fireEvent.click(button)
    expect(onAction).toHaveBeenCalledOnce()
  })

  it('does not render action button when only actionLabel is provided', () => {
    render(
      <EmptyState icon={Mail} title="No items" description="None found." actionLabel="Add" />
    )
    expect(screen.queryByText('Add')).not.toBeInTheDocument()
  })
})
