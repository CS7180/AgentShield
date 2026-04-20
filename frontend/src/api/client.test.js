import {
  API_BASE,
  calibrateJudge,
  compareScans,
  createScan,
  generateScanReport,
  getJudgeCalibrationReport,
  getScan,
  getScanReport,
  listAttackResults,
  listScanDeadLetters,
  listScans,
  startScan,
  stopScan,
} from './client'

describe('api client', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('uses the default API base when env is not provided', () => {
    expect(API_BASE).toBe('http://localhost:8080')
  })

  it('sends bearer auth and parses successful JSON responses', async () => {
    const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ scans: [] }),
    })

    const result = await listScans('token-123', { limit: 10, offset: 5 })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/scans?limit=10&offset=5',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({
          Authorization: 'Bearer token-123',
        }),
      }),
    )
    expect(result).toEqual({ scans: [] })
  })

  it('tags auth failures on error objects', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 401,
      json: vi.fn().mockResolvedValue({ error: 'bad token', code: 'UNAUTHORIZED' }),
    })

    await expect(getScan('scan-1', 'bad-token')).rejects.toMatchObject({
      message: 'bad token',
      status: 401,
      code: 'UNAUTHORIZED',
      isAuthError: true,
    })
  })

  it('tags gateway failures on error objects', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 500,
      json: vi.fn().mockResolvedValue({ error: 'boom', code: 'INTERNAL_ERROR' }),
    })

    await expect(getScanReport('scan-1', 'token')).rejects.toMatchObject({
      message: 'boom',
      status: 500,
      isGatewayError: true,
    })
  })

  it('issues POST and GET requests for the remaining API helpers', async () => {
    const fetchMock = vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ ok: true }),
    })

    await createScan({ target_endpoint: 'https://example.com', mode: 'red_team', attack_types: ['prompt_injection'] }, 'token')
    await startScan('scan-1', 'token')
    await stopScan('scan-1', 'token')
    await listAttackResults('scan-1', 'token', { limit: 5, offset: 1 })
    await listScanDeadLetters('scan-1', 'token', { limit: 2, offset: 0 })
    await generateScanReport('scan-1', 'token', true)
    await compareScans('scan-1', 'scan-2', 'token')
    await calibrateJudge([{ attack_type: 'prompt_injection' }], 'token')
    await getJudgeCalibrationReport('token')

    expect(fetchMock).toHaveBeenCalledTimes(9)
  })
})
