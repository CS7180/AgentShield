import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import useAuth from './auth/useAuth';

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { signInWithGoogle, signInWithPassword, isConfigured, loading } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [googleSubmitting, setGoogleSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');

  const from = location.state?.from?.pathname || '/dashboard';
  const reason = location.state?.reason;

  async function handleSubmit(event) {
    event.preventDefault();

    if (!isConfigured) {
      setErrorMessage('Supabase credentials are not configured yet.');
      return;
    }

    setSubmitting(true);
    setErrorMessage('');

    const { error } = await signInWithPassword({ email, password });

    if (error) {
      setErrorMessage(error.message);
      setSubmitting(false);
      return;
    }

    navigate(from, { replace: true });
  }

  async function handleGoogleSignIn() {
    if (!isConfigured) {
      setErrorMessage('Supabase credentials are not configured yet.');
      return;
    }

    setGoogleSubmitting(true);
    setErrorMessage('');

    const redirectTo = `${window.location.origin}${from}`;
    const { error } = await signInWithGoogle({ redirectTo });

    if (error) {
      setErrorMessage(error.message);
      setGoogleSubmitting(false);
    }
  }

  const statusMessage = !isConfigured
    ? 'Add VITE_SUPABASE_URL and VITE_SUPABASE_ANON_KEY in frontend/.env.local before testing login.'
    : reason === 'unauthorized'
      ? 'Sign in to continue to the protected application area.'
      : 'Use your Supabase email/password or Google account to access the dashboard.';

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: 32,
        background:
          'radial-gradient(circle at top, rgba(139,92,246,0.16), transparent 35%), radial-gradient(circle at 20% 20%, rgba(59,130,246,0.10), transparent 25%), #06070b',
        color: '#fff',
        fontFamily: "'Inter', system-ui, -apple-system, BlinkMacSystemFont, sans-serif",
      }}
    >
      <div style={{ width: 'min(440px, 100%)' }}>
        <Link to="/" style={{ display: 'inline-block', marginBottom: 18, fontSize: 13, color: '#9ca3af' }}>
          Back to landing
        </Link>

        <div
          style={{
            borderRadius: 20,
            padding: 28,
            border: '1px solid rgba(255,255,255,0.08)',
            background: 'linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.015))',
            boxShadow: '0 20px 60px rgba(0,0,0,0.32)',
          }}
        >
          <div style={{ marginBottom: 24 }}>
            <div style={{ fontSize: 30, fontWeight: 800, letterSpacing: '-0.03em', marginBottom: 8 }}>Sign in</div>
            <div style={{ fontSize: 14, color: '#a3a3a3', lineHeight: 1.6 }}>{statusMessage}</div>
          </div>

          {!isConfigured && (
            <div
              style={{
                marginBottom: 20,
                borderRadius: 14,
                border: '1px solid rgba(251,191,36,0.2)',
                background: 'rgba(245,158,11,0.08)',
                padding: 16,
              }}
            >
              <div style={{ fontSize: 12, fontWeight: 700, color: '#fbbf24', marginBottom: 8 }}>Setup required</div>
              <div style={{ fontSize: 13, color: '#d1d5db', lineHeight: 1.6 }}>
                Create `frontend/.env.local` from `frontend/.env.example`, then paste your Supabase project URL and anon key.
              </div>
            </div>
          )}

          {isConfigured && (
            <div
              style={{
                marginBottom: 20,
                borderRadius: 14,
                border: '1px solid rgba(96,165,250,0.18)',
                background: 'rgba(59,130,246,0.08)',
                padding: 16,
              }}
            >
              <div style={{ fontSize: 12, fontWeight: 700, color: '#93c5fd', marginBottom: 8 }}>Google OAuth setup</div>
              <div style={{ fontSize: 13, color: '#d1d5db', lineHeight: 1.6 }}>
                In Supabase Auth, enable the Google provider and add
                {' '}
                <code style={{ color: '#fff', background: 'rgba(255,255,255,0.06)', padding: '2px 6px', borderRadius: 6 }}>
                  {window.location.origin}
                </code>
                {' '}
                and your post-login routes to the allowed redirect URLs.
              </div>
            </div>
          )}

          <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 20 }}>
            <button
              type="button"
              onClick={handleGoogleSignIn}
              disabled={googleSubmitting || loading || !isConfigured}
              style={{
                height: 48,
                borderRadius: 12,
                border: '1px solid rgba(255,255,255,0.12)',
                background: 'rgba(255,255,255,0.04)',
                color: '#fff',
                fontSize: 14,
                fontWeight: 700,
                cursor: googleSubmitting || loading || !isConfigured ? 'not-allowed' : 'pointer',
                opacity: googleSubmitting || loading || !isConfigured ? 0.65 : 1,
              }}
            >
              {googleSubmitting ? 'Redirecting to Google...' : 'Continue with Google'}
            </button>

            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{ flex: 1, height: 1, background: 'rgba(255,255,255,0.08)' }} />
              <span style={{ fontSize: 11, color: '#6b7280', textTransform: 'uppercase', letterSpacing: '0.18em' }}>
                or
              </span>
              <div style={{ flex: 1, height: 1, background: 'rgba(255,255,255,0.08)' }} />
            </div>
          </div>

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <span style={{ fontSize: 12, color: '#9ca3af' }}>Email</span>
              <input
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="name@company.com"
                autoComplete="email"
                style={{
                  height: 46,
                  borderRadius: 12,
                  border: '1px solid rgba(255,255,255,0.08)',
                  background: 'rgba(255,255,255,0.03)',
                  color: '#fff',
                  padding: '0 14px',
                }}
              />
            </label>

            <label style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <span style={{ fontSize: 12, color: '#9ca3af' }}>Password</span>
              <input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Your Supabase password"
                autoComplete="current-password"
                style={{
                  height: 46,
                  borderRadius: 12,
                  border: '1px solid rgba(255,255,255,0.08)',
                  background: 'rgba(255,255,255,0.03)',
                  color: '#fff',
                  padding: '0 14px',
                }}
              />
            </label>

            {errorMessage && (
              <div
                style={{
                  borderRadius: 12,
                  background: 'rgba(244,63,94,0.08)',
                  border: '1px solid rgba(251,113,133,0.16)',
                  color: '#fda4af',
                  padding: '12px 14px',
                  fontSize: 13,
                  lineHeight: 1.5,
                }}
              >
                {errorMessage}
              </div>
            )}

            <button
              type="submit"
              disabled={submitting || loading || !isConfigured}
              style={{
                marginTop: 6,
                height: 48,
                borderRadius: 12,
                border: '1px solid rgba(217,70,239,0.35)',
                background: 'linear-gradient(135deg, rgba(217,70,239,0.25), rgba(139,92,246,0.15))',
                color: '#f3e8ff',
                fontSize: 14,
                fontWeight: 700,
                cursor: submitting || loading || !isConfigured ? 'not-allowed' : 'pointer',
                opacity: submitting || loading || !isConfigured ? 0.65 : 1,
              }}
            >
              {submitting ? 'Signing in...' : 'Sign in with email'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
