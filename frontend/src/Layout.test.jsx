import React from 'react'
import { render, screen } from '@testing-library/react'
import Layout from './Layout'

vi.mock('./Sidebar', () => ({
  default: ({ activeIndex, onNavigate }) => (
    <button type="button" onClick={() => onNavigate('dashboard')}>
      Sidebar {activeIndex}
    </button>
  ),
}))

describe('Layout', () => {
  it('renders the sidebar shell and main content', () => {
    const onNavigate = vi.fn()

    render(
      <Layout activeIndex={2} onNavigate={onNavigate}>
        <div>Child content</div>
      </Layout>,
    )

    expect(screen.getByRole('button', { name: 'Sidebar 2' })).toBeInTheDocument()
    expect(screen.getByText('Child content')).toBeInTheDocument()
  })
})
