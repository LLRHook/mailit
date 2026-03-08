import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/test-utils'
import MetricsPage from '../page'

vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { data: [], totals: { sent: 0, delivered: 0, bounced: 0, failed: 0, opened: 0, clicked: 0, complained: 0, delivery_rate: 0, open_rate: 0, bounce_rate: 0 } } }),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

// Mock recharts to avoid SVG rendering issues in jsdom
vi.mock('recharts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('recharts')>()
  return {
    ...actual,
    Area: () => null,
    AreaChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    Bar: () => null,
    BarChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    CartesianGrid: () => null,
    XAxis: () => null,
    YAxis: () => null,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  }
})

describe('MetricsPage', () => {
  it('renders the page header', () => {
    renderWithProviders(<MetricsPage />)
    expect(screen.getByText('Metrics')).toBeInTheDocument()
  })

  it('renders the page description', () => {
    renderWithProviders(<MetricsPage />)
    expect(screen.getByText('Monitor your email performance')).toBeInTheDocument()
  })

  it('renders all four stat cards', () => {
    renderWithProviders(<MetricsPage />)
    expect(screen.getByText('Emails Sent')).toBeInTheDocument()
    expect(screen.getByText('Delivery Rate')).toBeInTheDocument()
    expect(screen.getByText('Open Rate')).toBeInTheDocument()
    expect(screen.getByText('Bounce Rate')).toBeInTheDocument()
  })
})
