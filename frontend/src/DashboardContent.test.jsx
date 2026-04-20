import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import DashboardContent from './DashboardContent'

const mockUseAuth = vi.fn()
const mockListScans = vi.fn()
const mockGetScanReport = vi.fn()
const mockGetJudgeCalibrationReport = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./api/client', () => ({
  listScans: (...args) => mockListScans(...args),
  getScanReport: (...args) => mockGetScanReport(...args),
  getJudgeCalibrationReport: (...args) => mockGetJudgeCalibrationReport(...args),
}))

describe('DashboardContent', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    mockListScans.mockReset()
    mockGetScanReport.mockReset()
    mockGetJudgeCalibrationReport.mockReset()
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
    })
  })

  it('loads dashboard metrics from scan and report data', async () => {
    mockListScans.mockResolvedValue({
      scans: [
        {
          id: 'scan-1',
          created_at: '2026-04-19T00:00:00Z',
          attack_types: ['prompt_injection'],
          target_endpoint: 'https://example.com',
          mode: 'red_team',
          status: 'completed',
        },
      ],
    })
    mockGetScanReport.mockResolvedValue({
      overall_score: 87,
      critical_count: 1,
      owasp_scorecard: JSON.stringify({
        prompt_injection: { successful: 1 },
      }),
    })
    mockGetJudgeCalibrationReport.mockResolvedValue({ kendall_tau: 0.82 })

    render(
      <MemoryRouter>
        <DashboardContent />
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('TOTAL SCANS')).toBeInTheDocument()
      expect(screen.getByText('CRITICAL FINDINGS')).toBeInTheDocument()
      expect(screen.getByText('87.00%')).toBeInTheDocument()
      expect(screen.getByText('0.82')).toBeInTheDocument()
      expect(screen.getByText('https://example.com')).toBeInTheDocument()
    })
  })

  it('navigates to the scan page from the New Scan button', async () => {
    mockListScans.mockResolvedValue({ scans: [] })
    mockGetJudgeCalibrationReport.mockRejectedValue(new Error('not found'))

    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route path="/dashboard" element={<DashboardContent />} />
          <Route path="/scans" element={<div>Scans page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No scans found.')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: /\+ new scan/i }))

    await waitFor(() => {
      expect(screen.getByText('Scans page')).toBeInTheDocument()
    })
  })
})
