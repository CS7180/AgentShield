import React from 'react'
import { render, screen } from '@testing-library/react'
import AuthContext from './AuthContext'
import useAuth from './useAuth'

function Consumer() {
  const auth = useAuth()
  return <div data-testid="token">{auth.session.access_token}</div>
}

describe('useAuth', () => {
  it('returns the context value when rendered inside an AuthContext.Provider', () => {
    const mockValue = {
      session: { access_token: 'test-token' },
      user: null,
      loading: false,
    }

    render(
      <AuthContext.Provider value={mockValue}>
        <Consumer />
      </AuthContext.Provider>,
    )

    expect(screen.getByTestId('token')).toHaveTextContent('test-token')
  })

  it('throws when used outside any AuthContext.Provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    function BrokenConsumer() {
      useAuth()
      return null
    }

    expect(() => render(<BrokenConsumer />)).toThrow(
      'useAuth must be used within AuthProvider',
    )

    spy.mockRestore()
  })
})
