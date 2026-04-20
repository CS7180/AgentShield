import React from 'react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import ProtectedRoute from './ProtectedRoute'
import AuthContext from './AuthContext'

function renderRoute(authValue, initialEntry = '/dashboard') {
  return render(
    <AuthContext.Provider value={authValue}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <Routes>
          <Route element={<ProtectedRoute />}>
            <Route path="/dashboard" element={<div>Private page</div>} />
          </Route>
          <Route path="/login" element={<div>Login page</div>} />
        </Routes>
      </MemoryRouter>
    </AuthContext.Provider>,
  )
}

describe('ProtectedRoute', () => {
  it('shows a restoring message while loading', () => {
    renderRoute({ loading: true, session: null, isConfigured: true })

    expect(screen.getByText('Restoring session')).toBeInTheDocument()
  })

  it('redirects unauthenticated users to login', () => {
    renderRoute({ loading: false, session: null, isConfigured: true })

    expect(screen.getByText('Login page')).toBeInTheDocument()
  })

  it('redirects to login when config is missing', () => {
    renderRoute({ loading: false, session: null, isConfigured: false })

    expect(screen.getByText('Login page')).toBeInTheDocument()
  })

  it('renders protected content when a session exists', () => {
    renderRoute({ loading: false, session: { user: { id: 'u1' } }, isConfigured: true })

    expect(screen.getByText('Private page')).toBeInTheDocument()
  })
})
