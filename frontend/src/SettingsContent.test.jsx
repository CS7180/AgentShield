import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import SettingsContent from './SettingsContent'

const mockUseAuth = vi.fn()
const mockListScans = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./api/client', () => ({
  API_BASE: 'http://localhost:8080',
  listScans: (...args) => mockListScans(...args),
}))

vi.mock('./lib/supabase', () => ({
  isSupabaseConfigured: true,
  supabaseConfigStatus: { missingKeys: [] },
}))

describe('SettingsContent', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    mockListScans.mockReset()
    vi.restoreAllMocks()
  })

  it('shows successful gateway and authenticated API checks', async () => {
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
      loading: false,
      authError: '',
    })
    vi.spyOn(global, 'fetch').mockResolvedValue({ ok: true, status: 200 })
    mockListScans.mockResolvedValue({ scans: [] })

    render(<SettingsContent />)

    await waitFor(() => {
      expect(screen.getByText('Authenticated API request succeeded.')).toBeInTheDocument()
      expect(screen.getByText('Gateway health endpoint responded successfully.')).toBeInTheDocument()
    })
  })

  it('reports rejected auth tokens clearly', async () => {
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
      loading: false,
      authError: '',
    })
    vi.spyOn(global, 'fetch').mockResolvedValue({ ok: true, status: 200 })
    mockListScans.mockRejectedValue({ message: 'invalid token', isAuthError: true })

    render(<SettingsContent />)

    await waitFor(() => {
      expect(screen.getByText(/access token is missing, expired, or rejected/i)).toBeInTheDocument()
    })
  })

  it('reports when no active session is present', async () => {
    mockUseAuth.mockReturnValue({
      session: null,
      loading: false,
      authError: '',
    })
    vi.spyOn(global, 'fetch').mockResolvedValue({ ok: true, status: 200 })

    render(<SettingsContent />)

    await waitFor(() => {
      expect(screen.getByText('No active Supabase session found.')).toBeInTheDocument()
    })
  })
})
