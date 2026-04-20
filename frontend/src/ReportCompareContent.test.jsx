import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import ReportCompareContent from './ReportCompareContent'

const mockUseAuth = vi.fn()
const mockListScans = vi.fn()
const mockCompareScans = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./api/client', () => ({
  listScans: (...args) => mockListScans(...args),
  compareScans: (...args) => mockCompareScans(...args),
}))

describe('ReportCompareContent', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    mockListScans.mockReset()
    mockCompareScans.mockReset()
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
    })
  })

  it('loads scans and compares two reports', async () => {
    mockListScans.mockResolvedValue({
      scans: [
        { id: 'scan-1', mode: 'red_team', status: 'completed', created_at: '2026-04-19T00:00:00Z' },
        { id: 'scan-2', mode: 'red_team', status: 'completed', created_at: '2026-04-19T01:00:00Z' },
      ],
    })
    mockCompareScans.mockResolvedValue({
      overall_score: { base: 70, other: 82, delta: 12, trend: 'up' },
      severity_counts: {
        base: { critical: 2, high: 1, medium: 0, low: 0 },
        other: { critical: 1, high: 1, medium: 1, low: 0 },
        delta: { critical: -1, high: 0, medium: 1, low: 0 },
      },
      generated_at: '2026-04-19T02:00:00Z',
    })

    render(<ReportCompareContent />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /compare/i })).toBeEnabled()
    })

    fireEvent.click(screen.getByRole('button', { name: /^compare$/i }))

    await waitFor(() => {
      expect(mockCompareScans).toHaveBeenCalledWith('scan-1', 'scan-2', 'token')
      expect(screen.getByText('Score Delta')).toBeInTheDocument()
      expect(screen.getByText('+12.00')).toBeInTheDocument()
      expect(screen.getByText('Trend')).toBeInTheDocument()
      expect(screen.getByText('up')).toBeInTheDocument()
    })
  })

  it('shows a warning when fewer than two scans are available', async () => {
    mockListScans.mockResolvedValue({
      scans: [{ id: 'scan-1', mode: 'red_team', status: 'completed', created_at: '2026-04-19T00:00:00Z' }],
    })

    render(<ReportCompareContent />)

    await waitFor(() => {
      expect(screen.getByText(/At least two scans are required/i)).toBeInTheDocument()
    })
  })
})
