import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import NewScanContent from './NewScanContent'

const mockUseAuth = vi.fn()
const mockCreateScan = vi.fn()
const mockStartScan = vi.fn()

vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

vi.mock('./api/client', () => ({
  createScan: (...args) => mockCreateScan(...args),
  startScan: (...args) => mockStartScan(...args),
}))

describe('NewScanContent', () => {
  beforeEach(() => {
    mockUseAuth.mockReset()
    mockCreateScan.mockReset()
    mockStartScan.mockReset()
    mockUseAuth.mockReturnValue({
      session: { access_token: 'token' },
    })
  })

  it('blocks invalid URLs before hitting the API', async () => {
    render(<MemoryRouter><NewScanContent /></MemoryRouter>)

    fireEvent.click(screen.getByRole('button', { name: /create \+ start scan/i }))

    await waitFor(() => {
      expect(screen.getByText('Target endpoint must be a valid HTTPS URL.')).toBeInTheDocument()
    })
    expect(mockCreateScan).not.toHaveBeenCalled()
  })

  it('creates and auto-starts a scan successfully', async () => {
    mockCreateScan.mockResolvedValue({
      id: 'scan-1',
      mode: 'red_team',
      target_endpoint: 'https://example.com',
      status: 'pending',
    })
    mockStartScan.mockResolvedValue({
      status: 'running',
      message: 'started',
    })

    render(<MemoryRouter><NewScanContent /></MemoryRouter>)

    fireEvent.change(screen.getByPlaceholderText('https://example.com/v1/chat'), {
      target: { value: 'https://example.com' },
    })
    fireEvent.click(screen.getByRole('button', { name: /create \+ start scan/i }))

    await waitFor(() => {
      expect(mockCreateScan).toHaveBeenCalled()
      expect(mockStartScan).toHaveBeenCalledWith('scan-1', 'token')
      expect(screen.getByText('started')).toBeInTheDocument()
    })
  })

  it('can create a scan without auto-starting it', async () => {
    mockCreateScan.mockResolvedValue({
      id: 'scan-2',
      mode: 'red_team',
      target_endpoint: 'https://example.com',
      status: 'pending',
    })

    render(<MemoryRouter><NewScanContent /></MemoryRouter>)

    fireEvent.change(screen.getByPlaceholderText('https://example.com/v1/chat'), {
      target: { value: 'https://example.com' },
    })
    fireEvent.click(screen.getByRole('checkbox'))
    fireEvent.click(screen.getByRole('button', { name: /create scan/i }))

    await waitFor(() => {
      expect(mockCreateScan).toHaveBeenCalled()
      expect(mockStartScan).not.toHaveBeenCalled()
      expect(screen.getByText('Scan created.')).toBeInTheDocument()
    })
  })
})
