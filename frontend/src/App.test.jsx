import React from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from './App'

const mockUseAuth = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./LandingPage', () => ({
  default: () => <div>Landing page</div>,
}))

vi.mock('./LoginPage', () => ({
  default: () => <div>Login page</div>,
}))

vi.mock('./DashboardContent', () => ({
  default: () => <div>Dashboard page</div>,
}))

vi.mock('./NewScanContent', () => ({
  default: () => <div>Scans page</div>,
}))

vi.mock('./ScanMonitorContent', () => ({
  default: () => <div>Monitoring page</div>,
}))

vi.mock('./ReportCompareContent', () => ({
  default: () => <div>Reports page</div>,
}))

vi.mock('./JudgeContent', () => ({
  default: () => <div>Judge page</div>,
}))

vi.mock('./SettingsContent', () => ({
  default: () => <div>Settings page</div>,
}))

vi.mock('./Layout', () => ({
  default: ({ children }) => <div>{children}</div>,
}))

describe('App routes', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
  })

  it('renders the landing page at root', () => {
    mockUseAuth.mockReturnValue({ session: null, loading: false, isConfigured: true })

    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )

    expect(screen.getByText('Landing page')).toBeInTheDocument()
  })

  it('redirects authenticated users away from login', () => {
    mockUseAuth.mockReturnValue({ session: { user: { id: 'u1' } }, loading: false, isConfigured: true })

    render(
      <MemoryRouter initialEntries={['/login']}>
        <App />
      </MemoryRouter>,
    )

    expect(screen.getByText('Dashboard page')).toBeInTheDocument()
  })
})
