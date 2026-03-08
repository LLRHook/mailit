import { describe, it, expect, vi, beforeEach } from 'vitest'

const nextResponse = {
  next: vi.fn().mockReturnValue({ type: 'next' }),
  redirect: vi.fn().mockReturnValue({ type: 'redirect' }),
}

vi.mock('next/server', () => ({
  NextResponse: nextResponse,
}))

function createMockRequest(pathname: string, token?: string) {
  return {
    nextUrl: { pathname },
    url: 'http://localhost:3000' + pathname,
    cookies: {
      get: vi.fn((name: string) =>
        name === 'mailit_token' && token ? { value: token } : undefined
      ),
    },
  }
}

describe('middleware', () => {
  beforeEach(() => {
    vi.resetModules()
    nextResponse.next.mockClear()
    nextResponse.redirect.mockClear()
  })

  it('passes through in development mode', async () => {
    const originalEnv = process.env.NODE_ENV
    process.env.NODE_ENV = 'development'
    const { middleware } = await import('../middleware')
    const req = createMockRequest('/dashboard')
    middleware(req as never)
    expect(nextResponse.next).toHaveBeenCalled()
    process.env.NODE_ENV = originalEnv
  })

  it('allows public paths without token', async () => {
    const originalEnv = process.env.NODE_ENV
    process.env.NODE_ENV = 'production'
    const { middleware } = await import('../middleware')
    for (const path of ['/login', '/register', '/forgot-password']) {
      nextResponse.next.mockClear()
      middleware(createMockRequest(path) as never)
      expect(nextResponse.next).toHaveBeenCalled()
    }
    process.env.NODE_ENV = originalEnv
  })

  it('redirects to /login when no token on protected route', async () => {
    const originalEnv = process.env.NODE_ENV
    process.env.NODE_ENV = 'production'
    const { middleware } = await import('../middleware')
    middleware(createMockRequest('/dashboard') as never)
    expect(nextResponse.redirect).toHaveBeenCalled()
    process.env.NODE_ENV = originalEnv
  })

  it('passes through when token exists on protected route', async () => {
    const originalEnv = process.env.NODE_ENV
    process.env.NODE_ENV = 'production'
    const { middleware } = await import('../middleware')
    middleware(createMockRequest('/dashboard', 'valid-token') as never)
    expect(nextResponse.next).toHaveBeenCalled()
    process.env.NODE_ENV = originalEnv
  })
})
