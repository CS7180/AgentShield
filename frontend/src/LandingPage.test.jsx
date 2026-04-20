import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import LandingPage from './LandingPage'

describe('LandingPage', () => {
  it('renders product messaging and main calls to action', () => {
    render(<LandingPage onLogIn={vi.fn()} onGetStarted={vi.fn()} onViewDemo={vi.fn()} />)

    expect(screen.getByRole('heading', { name: /Red-blue teaming/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Log in/i })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /Get started/i })).toHaveLength(2)
  })

  it('fires CTA callbacks', () => {
    const onLogIn = vi.fn()
    const onGetStarted = vi.fn()

    render(<LandingPage onLogIn={onLogIn} onGetStarted={onGetStarted} onViewDemo={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: /Log in/i }))
    fireEvent.click(screen.getAllByRole('button', { name: /Get started/i })[0])

    expect(onLogIn).toHaveBeenCalled()
    expect(onGetStarted).toHaveBeenCalled()
  })
})
