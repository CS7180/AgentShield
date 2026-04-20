const createClientMock = vi.fn(() => ({ auth: {} }))

vi.mock('@supabase/supabase-js', () => ({
  createClient: (...args) => createClientMock(...args),
}))

describe('supabase config', () => {
  beforeEach(() => {
    vi.resetModules()
    createClientMock.mockClear()
  })

  it('reports missing frontend env vars when config is incomplete', async () => {
    vi.stubEnv('VITE_SUPABASE_URL', '')
    vi.stubEnv('VITE_SUPABASE_ANON_KEY', '')

    const module = await import('./supabase.js')

    expect(module.isSupabaseConfigured).toBe(false)
    expect(module.supabase).toBeNull()
    expect(module.supabaseConfigStatus.missingKeys).toEqual([
      'VITE_SUPABASE_URL',
      'VITE_SUPABASE_ANON_KEY',
    ])
  })

  it('creates the Supabase client when env vars are present', async () => {
    vi.stubEnv('VITE_SUPABASE_URL', 'https://example.supabase.co')
    vi.stubEnv('VITE_SUPABASE_ANON_KEY', 'anon-key')

    const module = await import('./supabase.js')

    expect(module.isSupabaseConfigured).toBe(true)
    expect(createClientMock).toHaveBeenCalledWith(
      'https://example.supabase.co',
      'anon-key',
      {
        auth: {
          autoRefreshToken: true,
          persistSession: true,
          detectSessionInUrl: true,
          flowType: 'pkce',
        },
      },
    )
  })
})
