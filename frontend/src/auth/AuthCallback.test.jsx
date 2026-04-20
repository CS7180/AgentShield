import React from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen, waitFor } from '@testing-library/react'
import AuthCallback from './AuthCallback'

const mockUseAuth = vi.fn()

vi.mock('./useAuth', () => ({
  default: () => mockUseAuth(),
}))

describe('AuthCallback', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
  })

  it('shows callback errors from the URL', () => {
    mockUseAuth.mockReturnValue({
      session: null,
      loading: false,
      isConfigured: true,
      authError: '',
    })

    render(
      <MemoryRouter initialEntries={['/auth/callback?error_description=OAuth+failed']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
        </Routes>
      </MemoryRouter>,
    )

    expect(screen.getByText(/OAuth failed/)).toBeInTheDocument()
  })

  it('redirects to the requested route when session is available', async () => {
    mockUseAuth.mockReturnValue({
      session: { user: { id: 'u1' } },
      loading: false,
      isConfigured: true,
      authError: '',
    })

    render(
      <MemoryRouter initialEntries={['/auth/callback?next=%2Fsettings&code=abc']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/settings" element={<div>Settings page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('Settings page')).toBeInTheDocument()
    })
  })

  it('redirects to login when Supabase is not configured', async () => {
    mockUseAuth.mockReturnValue({
      session: null,
      loading: false,
      isConfigured: false,
      authError: '',
    })

    render(
      <MemoryRouter initialEntries={['/auth/callback']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/login" element={<div>Login page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('Login page')).toBeInTheDocument()
    })
  })
})
