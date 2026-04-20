import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import JudgeContent from './JudgeContent'

const mockUseAuth = vi.fn()
const mockCalibrateJudge = vi.fn()
const mockGetJudgeCalibrationReport = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./api/client', () => ({
  calibrateJudge: (...args) => mockCalibrateJudge(...args),
  getJudgeCalibrationReport: (...args) => mockGetJudgeCalibrationReport(...args),
}))

describe('JudgeContent', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    mockCalibrateJudge.mockReset()
    mockGetJudgeCalibrationReport.mockReset()
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
    })
  })

  it('loads the latest calibration report', async () => {
    mockGetJudgeCalibrationReport.mockResolvedValue({
      sample_count: 3,
      overall: {
        f1: 0.8,
        precision: 0.75,
        recall: 0.9,
        accuracy: 0.67,
      },
      kendall_tau: 0.55,
      critical_severity_precision: 0.5,
      by_attack_type: {
        prompt_injection: {
          f1: 0.8,
          precision: 0.75,
          recall: 0.9,
          accuracy: 0.67,
        },
      },
    })

    render(<JudgeContent />)

    fireEvent.click(screen.getByRole('button', { name: /load latest report/i }))

    await waitFor(() => {
      expect(screen.getByText('Loaded latest calibration report.')).toBeInTheDocument()
      expect(screen.getByText('Overall Metrics')).toBeInTheDocument()
      expect(screen.getByText('prompt_injection')).toBeInTheDocument()
    })
  })

  it('runs calibration and renders the returned metrics', async () => {
    mockCalibrateJudge.mockResolvedValue({
      sample_count: 2,
      overall: {
        f1: 1,
        precision: 1,
        recall: 1,
        accuracy: 1,
      },
      kendall_tau: 0.9,
      critical_severity_precision: 1,
      by_attack_type: {},
    })

    render(<JudgeContent />)

    fireEvent.click(screen.getByRole('button', { name: /run calibration/i }))

    await waitFor(() => {
      expect(mockCalibrateJudge).toHaveBeenCalled()
      expect(screen.getByText('Calibration completed and persisted.')).toBeInTheDocument()
      expect(screen.getByText('No per-attack metrics.')).toBeInTheDocument()
    })
  })
})
