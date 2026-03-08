import { describe, it, expect, beforeEach } from 'vitest'
import { useUIStore } from '../store'
import { act, renderHook } from '@testing-library/react'

describe('useUIStore', () => {
  beforeEach(() => {
    // Reset store state between tests
    useUIStore.setState({ sidebarCollapsed: false })
  })

  it('starts with sidebar expanded', () => {
    const { result } = renderHook(() => useUIStore())
    expect(result.current.sidebarCollapsed).toBe(false)
  })

  it('toggles sidebar to collapsed', () => {
    const { result } = renderHook(() => useUIStore())
    act(() => result.current.toggleSidebar())
    expect(result.current.sidebarCollapsed).toBe(true)
  })

  it('toggles sidebar back to expanded', () => {
    const { result } = renderHook(() => useUIStore())
    act(() => result.current.toggleSidebar())
    act(() => result.current.toggleSidebar())
    expect(result.current.sidebarCollapsed).toBe(false)
  })
})
