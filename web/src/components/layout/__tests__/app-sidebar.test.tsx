import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { AppSidebar } from '../app-sidebar'

const mockPush = vi.fn()
const mockSetTheme = vi.fn()
let mockTheme = 'dark'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => '/emails',
}))

vi.mock('next-themes', () => ({
  useTheme: () => ({ theme: mockTheme, setTheme: mockSetTheme }),
}))

// Sidebar components need a context provider in production, but for unit tests
// we can mock them to render children directly
vi.mock('@/components/ui/sidebar', () => ({
  Sidebar: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => <div data-testid="sidebar" {...props}>{children}</div>,
  SidebarContent: ({ children }: React.PropsWithChildren) => <div>{children}</div>,
  SidebarGroup: ({ children }: React.PropsWithChildren) => <div>{children}</div>,
  SidebarGroupContent: ({ children }: React.PropsWithChildren) => <div>{children}</div>,
  SidebarHeader: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => <div {...props}>{children}</div>,
  SidebarMenu: ({ children }: React.PropsWithChildren) => <ul>{children}</ul>,
  SidebarMenuButton: ({ children }: React.PropsWithChildren<Record<string, unknown>>) => <div>{children}</div>,
  SidebarMenuItem: ({ children }: React.PropsWithChildren) => <li>{children}</li>,
  SidebarFooter: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => <div {...props}>{children}</div>,
}))

describe('AppSidebar', () => {
  beforeEach(() => {
    mockSetTheme.mockClear()
    mockTheme = 'dark'
  })

  it('renders nav items', () => {
    render(<AppSidebar />)
    expect(screen.getByText('Emails')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('renders theme toggle button', () => {
    render(<AppSidebar />)
    expect(screen.getByText('Toggle theme')).toBeInTheDocument()
  })

  it('toggles theme from dark to light', () => {
    render(<AppSidebar />)
    fireEvent.click(screen.getByText('Toggle theme'))
    expect(mockSetTheme).toHaveBeenCalledWith('light')
  })

  it('toggles theme from light to dark', () => {
    mockTheme = 'light'
    render(<AppSidebar />)
    fireEvent.click(screen.getByText('Toggle theme'))
    expect(mockSetTheme).toHaveBeenCalledWith('dark')
  })

  it('renders logout button', () => {
    render(<AppSidebar />)
    expect(screen.getByText('Log out')).toBeInTheDocument()
  })
})
