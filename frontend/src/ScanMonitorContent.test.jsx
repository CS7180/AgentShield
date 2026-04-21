import React, { act } from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import ScanMonitorContent from './ScanMonitorContent'

// ─── Mock: auth ──────────────────────────────────────────────────────────────
const mockUseAuth = vi.fn()
vi.mock('./auth/useAuth', () => ({
  default: () => mockUseAuth(),
}))

// ─── Mock: API client ────────────────────────────────────────────────────────
const mockListScans = vi.fn()
const mockGetScan = vi.fn()
const mockListAttackResults = vi.fn()
const mockListScanDeadLetters = vi.fn()
const mockGetScanReport = vi.fn()
const mockGenerateScanReport = vi.fn()
const mockStartScan = vi.fn()
const mockStopScan = vi.fn()

vi.mock('./api/client', () => ({
  API_BASE: 'http://localhost:8080',
  listScans: (...args) => mockListScans(...args),
  getScan: (...args) => mockGetScan(...args),
  listAttackResults: (...args) => mockListAttackResults(...args),
  listScanDeadLetters: (...args) => mockListScanDeadLetters(...args),
  getScanReport: (...args) => mockGetScanReport(...args),
  generateScanReport: (...args) => mockGenerateScanReport(...args),
  startScan: (...args) => mockStartScan(...args),
  stopScan: (...args) => mockStopScan(...args),
}))

// ─── Mock: WebSocket ─────────────────────────────────────────────────────────
let wsInstance = null

beforeAll(() => {
  vi.stubGlobal(
    'WebSocket',
    function MockWebSocket(url) {
      this.url = url
      this.onmessage = null
      this.close = vi.fn()
      wsInstance = this
    },
  )
})

afterAll(() => {
  vi.unstubAllGlobals()
})

// ─── Shared fixtures ─────────────────────────────────────────────────────────
const BASE_SCAN = {
  id: 'scan-1',
  mode: 'red_team',
  status: 'pending',
  attack_types: ['prompt_injection', 'jailbreak'],
  target_endpoint: 'https://example.com',
  created_at: '2026-04-19T00:00:00Z',
}

const RUNNING_SCAN = { ...BASE_SCAN, status: 'running' }
const COMPLETED_SCAN = { ...BASE_SCAN, status: 'completed' }
const FAILED_SCAN = { ...BASE_SCAN, status: 'failed' }

/**
 * Sets up the minimal happy-path mocks for a given scan state.
 * Individual tests can override specific mocks afterward.
 */
function setupMocks(scan = BASE_SCAN, { results = [], deadLetters = [] } = {}) {
  mockListScans.mockResolvedValue({ scans: [scan] })
  mockGetScan.mockResolvedValue(scan)
  mockListAttackResults.mockResolvedValue({ results })
  mockListScanDeadLetters.mockResolvedValue({ entries: deadLetters })
  // getScanReport is only called for completed/failed/stopped scans.
  // Default to rejecting so tests for non-terminal scans don't need to configure it.
  mockGetScanReport.mockRejectedValue(new Error('not found'))
}

// ─── Test suite ───────────────────────────────────────────────────────────────
describe('ScanMonitorContent', () => {
  beforeEach(() => {
    wsInstance = null
    mockUseAuth.mockReset()
    mockListScans.mockReset()
    mockGetScan.mockReset()
    mockListAttackResults.mockReset()
    mockListScanDeadLetters.mockReset()
    mockGetScanReport.mockReset()
    mockGenerateScanReport.mockReset()
    mockStartScan.mockReset()
    mockStopScan.mockReset()

    mockUseAuth.mockReturnValue({ session: { access_token: 'tok' } })
  })

  // ── Empty / error states ────────────────────────────────────────────────────

  it('shows "No scans found" option when the scan list is empty', async () => {
    mockListScans.mockResolvedValue({ scans: [] })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('No scans found')).toBeInTheDocument()
    })
  })

  it('shows an error message when the scan list fails to load', async () => {
    mockListScans.mockRejectedValue(new Error('network error'))

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('network error')).toBeInTheDocument()
    })
  })

  it('shows a loading indicator while the first detail fetch is in progress', async () => {
    // Block detail loading so we can observe the loading state
    mockListScans.mockResolvedValue({ scans: [BASE_SCAN] })
    mockGetScan.mockReturnValue(new Promise(() => {})) // never resolves
    mockListAttackResults.mockResolvedValue({ results: [] })
    mockListScanDeadLetters.mockResolvedValue({ entries: [] })

    render(<ScanMonitorContent />)

    // After listScans resolves, selectedScanID is set and loadDetail fires.
    // Because getScan never resolves, loading stays true.
    await waitFor(() => {
      expect(screen.getByText('Loading scan data...')).toBeInTheDocument()
    })
  })

  // ── Scan selection ──────────────────────────────────────────────────────────

  it('auto-selects the first scan and loads its detail', async () => {
    setupMocks(BASE_SCAN)

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(mockGetScan).toHaveBeenCalledWith('scan-1', 'tok')
    })
    // Target endpoint line appears after detail resolves
    await waitFor(() => {
      expect(screen.getByText(/https:\/\/example\.com/)).toBeInTheDocument()
    })
  })

  it('prioritises a running scan over the most-recent scan during auto-selection', async () => {
    // newerPendingScan is created_at 2026-04-20 → sorts first, but is not running
    // olderRunningScan is created_at 2026-04-19 → sorts second, but IS running → wins
    const newerPendingScan = { ...BASE_SCAN, id: 'scan-A', status: 'pending', created_at: '2026-04-20T00:00:00Z' }
    const olderRunningScan = { ...BASE_SCAN, id: 'scan-B', status: 'running', created_at: '2026-04-19T00:00:00Z' }

    mockListScans.mockResolvedValue({ scans: [newerPendingScan, olderRunningScan] })
    mockGetScan.mockResolvedValue(olderRunningScan)
    mockListAttackResults.mockResolvedValue({ results: [] })
    mockListScanDeadLetters.mockResolvedValue({ entries: [] })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(mockGetScan).toHaveBeenCalledWith('scan-B', 'tok')
    })
  })

  // ── Data display ────────────────────────────────────────────────────────────

  it('renders a row for each attack type in the Red Team Agents section', async () => {
    setupMocks(BASE_SCAN)

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('Prompt Injection')).toBeInTheDocument()
      expect(screen.getByText('Jailbreak')).toBeInTheDocument()
    })
  })

  it('computes defense summary metrics (blocked, successful, avg confidence) from results', async () => {
    const results = [
      {
        id: 'r1',
        attack_type: 'prompt_injection',
        attack_success: true,
        defense_intercepted: false,
        judge_confidence: 0.9,
        severity: 'high',
        owasp_category: 'LLM01',
        created_at: '2026-04-19T00:00:00Z',
      },
      {
        id: 'r2',
        attack_type: 'jailbreak',
        attack_success: false,
        defense_intercepted: true,
        judge_confidence: 0.7,
        severity: 'low',
        owasp_category: 'LLM02',
        created_at: '2026-04-19T00:00:01Z',
      },
    ]
    setupMocks(RUNNING_SCAN, { results })

    render(<ScanMonitorContent />)

    // avg confidence = (0.9 + 0.7) / 2 = 0.80
    await waitFor(() => {
      expect(screen.getByText('0.80')).toBeInTheDocument()
    })
  })

  it('shows 100% progress for a completed scan', async () => {
    setupMocks(COMPLETED_SCAN)
    // completed scan triggers auto-generate; stub it to avoid extra assertions
    mockGenerateScanReport.mockResolvedValue({})

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('100%')).toBeInTheDocument()
    })
  })

  it('shows 100% progress for a failed scan', async () => {
    setupMocks(FAILED_SCAN)
    mockGetScanReport.mockRejectedValue(new Error('no report'))

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('100%')).toBeInTheDocument()
    })
  })

  it('populates the live activity feed from attack result history', async () => {
    const results = [
      {
        id: 'r1',
        attack_type: 'data_leakage',
        attack_success: true,
        owasp_category: 'LLM02',
        defense_intercepted: false,
        judge_confidence: 0.85,
        severity: 'high',
        created_at: '2026-04-19T00:00:00Z',
      },
    ]
    setupMocks(RUNNING_SCAN, { results })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('Data Leakage')).toBeInTheDocument()
      expect(screen.getByText('Attack succeeded (LLM02)')).toBeInTheDocument()
    })
  })

  it('shows "Attack blocked or ineffective" feed entry for failed attacks', async () => {
    const results = [
      {
        id: 'r1',
        attack_type: 'jailbreak',
        attack_success: false,
        defense_intercepted: true,
        judge_confidence: 0.8,
        severity: 'low',
        owasp_category: null,
        created_at: '2026-04-19T00:00:00Z',
      },
    ]
    setupMocks(RUNNING_SCAN, { results })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('Attack blocked or ineffective')).toBeInTheDocument()
    })
  })

  it('renders dead-letter entries when present', async () => {
    const deadLetters = [
      {
        id: 'dl-1',
        attempt_count: 3,
        error_stage: 'judge',
        error_message: 'Gemini API timeout',
        failed_at: '2026-04-19T01:00:00Z',
      },
    ]
    setupMocks(BASE_SCAN, { deadLetters })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText(/Gemini API timeout/)).toBeInTheDocument()
      expect(screen.getByText(/attempts 3/)).toBeInTheDocument()
    })
  })

  // ── Report lifecycle ────────────────────────────────────────────────────────

  it('auto-generates a report when scan is completed and report is missing', async () => {
    setupMocks(COMPLETED_SCAN)
    mockGenerateScanReport.mockResolvedValue({
      report_json_path: '/reports/scan-1.json',
      report_pdf_path: '/reports/scan-1.pdf',
    })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(mockGenerateScanReport).toHaveBeenCalledWith('scan-1', 'tok', true)
    })
    await waitFor(() => {
      expect(screen.getByText('ready')).toBeInTheDocument()
    })
  })

  it('shows "partial" report status when only the JSON artifact is generated', async () => {
    setupMocks(COMPLETED_SCAN)
    mockGenerateScanReport.mockResolvedValue({
      report_json_path: '/reports/scan-1.json',
      report_pdf_path: null,
    })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('partial')).toBeInTheDocument()
    })
  })

  it('shows "ready" report status when existing report has both artifacts', async () => {
    setupMocks(COMPLETED_SCAN)
    mockGetScanReport.mockResolvedValue({
      report_json_path: '/reports/scan-1.json',
      report_pdf_path: '/reports/scan-1.pdf',
    })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('ready')).toBeInTheDocument()
    })
  })

  it('shows "failed" report status when auto-generation throws', async () => {
    setupMocks(COMPLETED_SCAN)
    mockGenerateScanReport.mockRejectedValue(new Error('storage unavailable'))

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByText('failed')).toBeInTheDocument()
    })
  })

  it('generates a report manually when the user clicks "Generate Report Now"', async () => {
    setupMocks(RUNNING_SCAN)
    mockGenerateScanReport.mockResolvedValue({
      report_json_path: '/reports/scan-1.json',
      report_pdf_path: null,
    })

    render(<ScanMonitorContent />)

    // Wait for initial load to complete
    await waitFor(() => {
      expect(screen.queryByText('Loading scan data...')).not.toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Generate Report Now' }))

    await waitFor(() => {
      expect(mockGenerateScanReport).toHaveBeenCalledWith('scan-1', 'tok', true)
    })
  })

  // ── Actions: Start / Stop ───────────────────────────────────────────────────

  it('calls startScan and shows the response message on success', async () => {
    setupMocks(BASE_SCAN)
    mockStartScan.mockResolvedValue({ status: 'running', message: 'Scan started successfully' })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.queryByText('Loading scan data...')).not.toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Start' }))

    await waitFor(() => {
      expect(mockStartScan).toHaveBeenCalledWith('scan-1', 'tok')
      expect(screen.getByText('Scan started successfully')).toBeInTheDocument()
    })
  })

  it('shows a queued-error message when the orchestrator is unavailable', async () => {
    // Initial getScan → pending; after handleStart's refresh → queued
    mockListScans.mockResolvedValue({ scans: [BASE_SCAN] })
    mockGetScan
      .mockResolvedValueOnce(BASE_SCAN)
      .mockResolvedValue({ ...BASE_SCAN, status: 'queued' })
    mockListAttackResults.mockResolvedValue({ results: [] })
    mockListScanDeadLetters.mockResolvedValue({ entries: [] })
    mockStartScan.mockResolvedValue({ status: 'queued', message: 'queued' })

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.queryByText('Loading scan data...')).not.toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Start' }))

    await waitFor(() => {
      expect(screen.getByText(/Queued — orchestrator unavailable/)).toBeInTheDocument()
    })
  })

  it('calls stopScan when the Stop button is clicked', async () => {
    setupMocks(RUNNING_SCAN)
    mockStopScan.mockResolvedValue({})

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Stop' })).not.toBeDisabled()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Stop' }))

    await waitFor(() => {
      expect(mockStopScan).toHaveBeenCalledWith('scan-1', 'tok')
    })
  })

  it('shows an error message when stopScan throws', async () => {
    setupMocks(RUNNING_SCAN)
    mockStopScan.mockRejectedValue(new Error('stop failed'))

    render(<ScanMonitorContent />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Stop' })).not.toBeDisabled()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Stop' }))

    await waitFor(() => {
      expect(screen.getByText('stop failed')).toBeInTheDocument()
    })
  })

  // ── WebSocket ───────────────────────────────────────────────────────────────

  it('appends a parsed JSON WebSocket event to the live activity feed', async () => {
    setupMocks(RUNNING_SCAN)

    render(<ScanMonitorContent />)

    await waitFor(() => expect(wsInstance).not.toBeNull())

    act(() => {
      wsInstance.onmessage({
        data: JSON.stringify({ topic: 'agent.status', data: { phase: 'attacking' } }),
      })
    })

    await waitFor(() => {
      expect(screen.getByText('agent.status')).toBeInTheDocument()
    })
  })

  it('falls back to raw string in the feed for non-JSON WebSocket messages', async () => {
    setupMocks(RUNNING_SCAN)

    render(<ScanMonitorContent />)

    await waitFor(() => expect(wsInstance).not.toBeNull())

    act(() => {
      wsInstance.onmessage({ data: 'plain-text-event' })
    })

    await waitFor(() => {
      expect(screen.getByText('plain-text-event')).toBeInTheDocument()
    })
  })

  it('closes the WebSocket when the selected scan changes', async () => {
    const scan2 = { ...BASE_SCAN, id: 'scan-2', status: 'pending', created_at: '2026-04-18T00:00:00Z' }

    mockListScans.mockResolvedValue({ scans: [BASE_SCAN, scan2] })
    mockGetScan.mockResolvedValue(BASE_SCAN)
    mockListAttackResults.mockResolvedValue({ results: [] })
    mockListScanDeadLetters.mockResolvedValue({ entries: [] })

    render(<ScanMonitorContent />)

    // Wait for first WS to be created
    await waitFor(() => expect(wsInstance).not.toBeNull())
    const firstWs = wsInstance

    // Change the selected scan via the dropdown
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'scan-2' } })

    await waitFor(() => {
      expect(firstWs.close).toHaveBeenCalled()
    })
  })
})
