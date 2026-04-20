import React from 'react';
import { useEffect, useState } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import useAuth from './useAuth';

function readNextPath(search) {
  const params = new URLSearchParams(search);
  const next = params.get('next');
  if (!next || !next.startsWith('/')) {
    return '/dashboard';
  }
  return next;
}

export default function AuthCallback() {
  const navigate = useNavigate();
  const location = useLocation();
  const { session, loading, isConfigured, authError } = useAuth();
  const [errorMessage, setErrorMessage] = useState('');
  const [hasCallbackState, setHasCallbackState] = useState(false);

  useEffect(() => {
    if (!isConfigured) {
      return;
    }

    const searchParams = new URLSearchParams(location.search);
    const hashParams = new URLSearchParams(location.hash.startsWith('#') ? location.hash.slice(1) : location.hash);
    const errorFromURL =
      searchParams.get('error_description') ||
      searchParams.get('error') ||
      hashParams.get('error_description') ||
      hashParams.get('error');

    const hasCode = searchParams.has('code');
    const hasTokensInHash = hashParams.has('access_token') || hashParams.has('refresh_token');

    setHasCallbackState(hasCode || hasTokensInHash || Boolean(errorFromURL));
    setErrorMessage(errorFromURL || '');
  }, [isConfigured, location.search, location.hash]);

  useEffect(() => {
    if (loading || !session) {
      return;
    }

    const nextPath = readNextPath(location.search);
    navigate(nextPath, { replace: true });
  }, [session, loading, location.search, navigate]);

  useEffect(() => {
    if (loading || session || errorMessage || authError || !hasCallbackState) {
      return;
    }

    const timer = window.setTimeout(() => {
      setErrorMessage('Supabase did not establish a session from the callback URL. Please try signing in again.');
    }, 4000);

    return () => window.clearTimeout(timer);
  }, [loading, session, errorMessage, authError, hasCallbackState]);

  if (!isConfigured) {
    return <Navigate to="/login" replace state={{ reason: 'missing-config' }} />;
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: 32,
        background:
          'radial-gradient(circle at top, rgba(139,92,246,0.14), transparent 35%), #06070b',
        color: '#fff',
      }}
    >
      <div
        style={{
          width: 'min(520px, 100%)',
          borderRadius: 18,
          padding: 28,
          border: '1px solid rgba(255,255,255,0.08)',
          background: 'rgba(255,255,255,0.03)',
        }}
      >
        <div style={{ fontSize: 22, fontWeight: 700, marginBottom: 10 }}>
          Completing sign-in
        </div>
        <div style={{ fontSize: 14, color: '#a3a3a3', lineHeight: 1.6 }}>
          {errorMessage || authError || 'Finalizing your Google OAuth session with Supabase.'}
        </div>
        {(errorMessage || authError) && (
          <button
            type="button"
            onClick={() => navigate('/login', { replace: true, state: { reason: 'oauth-error' } })}
            style={{
              marginTop: 16,
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.12)',
              background: 'rgba(255,255,255,0.04)',
              color: '#fff',
              height: 40,
              padding: '0 14px',
              cursor: 'pointer',
            }}
          >
            Back to login
          </button>
        )}
      </div>
    </div>
  );
}
