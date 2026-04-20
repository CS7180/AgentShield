import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import LoginPage from './LoginPage'

const mockUseAuth = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./lib/supabase', () => ({
  supabaseConfigStatus: {
    missingKeys: ['VITE_SUPABASE_URL', 'VITE_SUPABASE_ANON_KEY'],
  },
}))

describe('LoginPage', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
  })

  it('shows setup guidance when Supabase is not configured', () => {
    mockUseAuth.mockReturnValue({
      signInWithGoogle: vi.fn(),
      isConfigured: false,
      loading: false,
      authError: '',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    expect(screen.getByText('Setup required')).toBeInTheDocument()
    expect(screen.getByText(/VITE_SUPABASE_URL, VITE_SUPABASE_ANON_KEY/)).toBeInTheDocument()
  })

  it('calls Google sign-in with the auth callback URL', async () => {
    const signInWithGoogle = vi.fn().mockResolvedValue({ error: null })
    mockUseAuth.mockReturnValue({
      signInWithGoogle,
      isConfigured: true,
      loading: false,
      authError: '',
    })

    render(
      <MemoryRouter initialEntries={[{ pathname: '/login', state: { from: { pathname: '/settings' } } }]}>
        <LoginPage />
      </MemoryRouter>,
    )

    fireEvent.click(screen.getByRole('button', { name: /continue with google/i }))

    await waitFor(() => {
      expect(signInWithGoogle).toHaveBeenCalledWith({
        redirectTo: 'http://localhost:3000/auth/callback?next=%2Fsettings',
      })
    })
  })

  it('shows provider bootstrap errors from auth context', () => {
    mockUseAuth.mockReturnValue({
      signInWithGoogle: vi.fn(),
      isConfigured: true,
      loading: false,
      authError: 'Provider bootstrap failed',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    expect(screen.getByText('Provider bootstrap failed')).toBeInTheDocument()
  })
})
