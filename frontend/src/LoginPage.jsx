import React from 'react';
import { useLocation, Link } from 'react-router-dom';
import useAuth from './auth/useAuth';
import { useState } from 'react';
import { supabaseConfigStatus } from './lib/supabase';

export default function LoginPage() {
  const location = useLocation();
  const { signInWithGoogle, isConfigured, loading, authError } = useAuth();
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');

  const from = location.state?.from?.pathname || '/dashboard';
  const reason = location.state?.reason;

  async function handleGoogleSignIn() {
    if (!isConfigured) {
      setErrorMessage('Supabase credentials are not configured yet.');
      return;
    }

    setSubmitting(true);
    setErrorMessage('');

    const redirectTo = `${window.location.origin}/auth/callback?next=${encodeURIComponent(from)}`;
    const { error } = await signInWithGoogle({ redirectTo });

    if (error) {
      setErrorMessage(error.message);
      setSubmitting(false);
    }
  }

  const statusMessage = !isConfigured
    ? 'Add VITE_SUPABASE_URL and VITE_SUPABASE_ANON_KEY in frontend/.env.local before testing login.'
    : reason === 'unauthorized'
      ? 'Sign in to continue to the protected application area.'
      : 'Sign in with your Google account to access the dashboard.';

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
              <div style={{ fontSize: 12, color: '#fcd34d', marginTop: 10, lineHeight: 1.6 }}>
                Missing:
                {' '}
                {supabaseConfigStatus.missingKeys.join(', ') || 'Unknown config values'}
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

          {errorMessage && (
            <div
              style={{
                marginBottom: 16,
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

          {!errorMessage && authError && (
            <div
              style={{
                marginBottom: 16,
                borderRadius: 12,
                background: 'rgba(244,63,94,0.08)',
                border: '1px solid rgba(251,113,133,0.16)',
                color: '#fda4af',
                padding: '12px 14px',
                fontSize: 13,
                lineHeight: 1.5,
              }}
            >
              {authError}
            </div>
          )}

          <button
            type="button"
            onClick={handleGoogleSignIn}
            disabled={submitting || loading || !isConfigured}
            style={{
              width: '100%',
              height: 48,
              borderRadius: 12,
              border: '1px solid rgba(255,255,255,0.12)',
              background: 'rgba(255,255,255,0.04)',
              color: '#fff',
              fontSize: 14,
              fontWeight: 700,
              cursor: submitting || loading || !isConfigured ? 'not-allowed' : 'pointer',
              opacity: submitting || loading || !isConfigured ? 0.65 : 1,
            }}
          >
            {submitting ? 'Redirecting to Google...' : 'Continue with Google'}
          </button>
        </div>
      </div>
    </div>
  );
}
