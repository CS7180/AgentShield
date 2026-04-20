import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { AuthProvider } from './AuthProvider'
import useAuth from './useAuth'

const mockGetSession = vi.fn()
const mockOnAuthStateChange = vi.fn()
const mockSignInWithOAuth = vi.fn()
const mockSignOut = vi.fn()

vi.mock('../lib/supabase', () => ({
  isSupabaseConfigured: true,
  supabase: {
    auth: {
      getSession: (...args) => mockGetSession(...args),
      onAuthStateChange: (...args) => mockOnAuthStateChange(...args),
      signInWithOAuth: (...args) => mockSignInWithOAuth(...args),
      signOut: (...args) => mockSignOut(...args),
    },
  },
}))

function Consumer() {
  const auth = useAuth()

  return (
    <div>
      <div>{auth.loading ? 'loading' : 'ready'}</div>
      <div>{auth.user?.email ?? 'no-user'}</div>
      <div>{auth.authError || 'no-error'}</div>
      <button type="button" onClick={() => auth.signInWithGoogle({ redirectTo: 'http://localhost/callback' })}>
        sign in
      </button>
      <button type="button" onClick={() => auth.signOut()}>
        sign out
      </button>
    </div>
  )
}

describe('AuthProvider', () => {
  let consoleErrorSpy

  beforeEach(() => {
    mockGetSession.mockReset()
    mockOnAuthStateChange.mockReset()
    mockSignInWithOAuth.mockReset()
    mockSignOut.mockReset()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockOnAuthStateChange.mockReturnValue({
      data: {
        subscription: {
          unsubscribe: vi.fn(),
        },
      },
    })
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  it('restores the current session and exposes auth actions', async () => {
    mockGetSession.mockResolvedValue({
      data: {
        session: {
          user: { email: 'user@example.com' },
        },
      },
      error: null,
    })

    render(
      <AuthProvider>
        <Consumer />
      </AuthProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('ready')).toBeInTheDocument()
      expect(screen.getByText('user@example.com')).toBeInTheDocument()
      expect(screen.getByText('no-error')).toBeInTheDocument()
    })

    screen.getByRole('button', { name: 'sign in' }).click()
    screen.getByRole('button', { name: 'sign out' }).click()

    expect(mockSignInWithOAuth).toHaveBeenCalledWith({
      provider: 'google',
      options: {
        redirectTo: 'http://localhost/callback',
      },
    })
    expect(mockSignOut).toHaveBeenCalled()
  })

  it('surfaces bootstrap errors from Supabase', async () => {
    mockGetSession.mockResolvedValue({
      data: { session: null },
      error: { message: 'bootstrap failed' },
    })

    render(
      <AuthProvider>
        <Consumer />
      </AuthProvider>,
    )

    await waitFor(() => {
      expect(screen.getByText('ready')).toBeInTheDocument()
      expect(screen.getByText('no-user')).toBeInTheDocument()
      expect(screen.getByText('bootstrap failed')).toBeInTheDocument()
    })
  })
})
