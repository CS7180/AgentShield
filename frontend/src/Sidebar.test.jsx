import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import Sidebar from './Sidebar'

const mockUseAuth = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

describe('Sidebar', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    vi.restoreAllMocks()
  })

  it('shows the current user email', () => {
    mockUseAuth.mockReturnValue({
      user: { email: 'user@example.com' },
      signOut: vi.fn().mockResolvedValue({ error: null }),
    })

    render(<Sidebar activeIndex={0} onNavigate={vi.fn()} />)

    expect(screen.getByText('user@example.com')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign out/i })).toBeInTheDocument()
  })

  it('calls signOut and redirects to the landing page', async () => {
    const signOut = vi.fn().mockResolvedValue({ error: null })
    const assignSpy = vi.fn()
    mockUseAuth.mockReturnValue({
      user: { email: 'user@example.com' },
      signOut,
    })

    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...originalLocation, assign: assignSpy },
    })

    render(<Sidebar activeIndex={0} onNavigate={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: /sign out/i }))

    await waitFor(() => {
      expect(signOut).toHaveBeenCalled()
      expect(assignSpy).toHaveBeenCalledWith('/')
    })
  })
})
