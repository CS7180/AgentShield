import React from 'react';
import { useEffect, useMemo, useState } from 'react';
import { API_BASE, listScans } from './api/client';
import useAuth from './auth/useAuth';
import { isSupabaseConfigured, supabaseConfigStatus } from './lib/supabase';

function Dot({ color, size = 7 }) {
  return (
    <span
      style={{
        height: size,
        width: size,
        borderRadius: '50%',
        background: color,
        boxShadow: `0 0 10px ${color}`,
        flexShrink: 0,
      }}
    />
  );
}

function tone(ok) {
  return ok ? '#34d399' : '#fb7185';
}

export default function SettingsContent() {
  const { session, loading, authError } = useAuth();
  const token = session?.access_token;

  const [gatewayHealth, setGatewayHealth] = useState(null);
  const [apiAuthHealth, setAPIAuthHealth] = useState(null);
  const [apiAuthMessage, setAPIAuthMessage] = useState('');
  const [gatewayMessage, setGatewayMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;

    async function runChecks() {
      setError('');
      setGatewayMessage('');
      setAPIAuthMessage('');
      try {
        const res = await fetch(`${API_BASE}/health`);
        const ok = res.ok;
        if (!cancelled) {
          setGatewayHealth(ok);
          setGatewayMessage(ok ? 'Gateway health endpoint responded successfully.' : `Gateway health check returned HTTP ${res.status}.`);
        }
      } catch {
        if (!cancelled) {
          setGatewayHealth(false);
          setGatewayMessage('Failed to reach the gateway health endpoint.');
        }
      }

      if (loading) {
        if (!cancelled) {
          setAPIAuthHealth(null);
          setAPIAuthMessage('Waiting for auth session to finish loading.');
        }
        return;
      }

      if (!isSupabaseConfigured) {
        if (!cancelled) {
          setAPIAuthHealth(false);
          setAPIAuthMessage(`Missing Supabase env vars: ${supabaseConfigStatus.missingKeys.join(', ')}`);
        }
        return;
      }

      if (!token) {
        if (!cancelled) {
          setAPIAuthHealth(false);
          setAPIAuthMessage(authError || 'No active Supabase session found.');
        }
        return;
      }

      try {
        await listScans(token, { limit: 1, offset: 0 });
        if (!cancelled) {
          setAPIAuthHealth(true);
          setAPIAuthMessage('Authenticated API request succeeded.');
        }
      } catch (err) {
        if (!cancelled) {
          setAPIAuthHealth(false);
          if (err.isAuthError) {
            setAPIAuthMessage('Authenticated API request failed. The access token is missing, expired, or rejected by the backend.');
          } else {
            setAPIAuthMessage(err.message || 'Authenticated API check failed.');
          }
          setError(err.message || 'Authenticated API check failed');
        }
      }
    }

    runChecks();
    return () => {
      cancelled = true;
    };
  }, [token, loading, authError]);

  const wsBase = useMemo(() => API_BASE.replace(/^http/, 'ws'), []);

  const checklist = [
    { label: 'Supabase frontend config', ok: isSupabaseConfigured },
    { label: 'Current user session', ok: Boolean(token) && !loading },
    { label: 'Gateway health endpoint', ok: gatewayHealth === true },
    { label: 'Authenticated API access', ok: apiAuthHealth === true },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Settings & Diagnostics</h1>
        <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
          Runtime configuration visibility and quick connectivity checks.
        </p>
      </div>

      {error && <div style={{ color: '#fb7185', fontSize: 12 }}>{error}</div>}

      <div style={panelStyle}>
        <div style={sectionTitle}>Environment</div>
        <div style={{ display: 'grid', gap: 8, fontSize: 12, color: '#d4d4d4' }}>
          <div>
            API base: <code>{API_BASE}</code>
          </div>
          <div>
            WebSocket base: <code>{wsBase}</code>
          </div>
          <div>
            Supabase configured: <code>{String(isSupabaseConfigured)}</code>
          </div>
          <div>
            Missing Supabase keys: <code>{supabaseConfigStatus.missingKeys.join(', ') || 'None'}</code>
          </div>
          <div>
            Access token present: <code>{String(Boolean(token))}</code>
          </div>
          <div>
            Auth loading: <code>{String(loading)}</code>
          </div>
        </div>
      </div>

      <div style={panelStyle}>
        <div style={sectionTitle}>Health Checklist</div>
        <div style={{ display: 'grid', gap: 10 }}>
          {checklist.map((item) => (
            <div
              key={item.label}
              style={{
                borderRadius: 10,
                border: '1px solid rgba(255,255,255,0.07)',
                background: 'rgba(255,255,255,0.02)',
                padding: '10px 12px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 8,
              }}
            >
              <span style={{ color: '#d4d4d4', fontSize: 12 }}>{item.label}</span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: tone(item.ok), fontSize: 12 }}>
                <Dot color={tone(item.ok)} size={7} />
                {item.ok ? 'OK' : 'Not Ready'}
              </span>
            </div>
          ))}
        </div>
      </div>

      <div style={panelStyle}>
        <div style={sectionTitle}>Detailed Status</div>
        <div style={{ display: 'grid', gap: 10 }}>
          <div style={detailCardStyle}>
            <div style={detailLabelStyle}>Gateway</div>
            <div style={detailMessageStyle}>{gatewayMessage || 'Health check has not run yet.'}</div>
          </div>
          <div style={detailCardStyle}>
            <div style={detailLabelStyle}>Authenticated API</div>
            <div style={detailMessageStyle}>{apiAuthMessage || 'Authenticated API check has not run yet.'}</div>
          </div>
          {authError && (
            <div style={detailCardStyle}>
              <div style={detailLabelStyle}>Auth Provider</div>
              <div style={detailMessageStyle}>{authError}</div>
            </div>
          )}
        </div>
      </div>

      <div style={panelStyle}>
        <div style={sectionTitle}>Integration Notes</div>
        <div style={{ display: 'grid', gap: 6, fontSize: 12, color: '#a3a3a3', lineHeight: 1.7 }}>
          <div>Use `AGENTS_EXECUTION_MODE=target_http` to send attacks to target endpoint directly.</div>
          <div>Use `JUDGE_EVAL_MODE=openai_compat` with `JUDGE_LLM_*` vars to enable model-based judging.</div>
          <div>Orchestrator emits scan progress to Kafka topic `agent.status` for WebSocket fanout.</div>
        </div>
      </div>
    </div>
  );
}

const panelStyle = {
  borderRadius: 14,
  border: '1px solid rgba(255,255,255,0.06)',
  background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
  padding: 16,
};

const sectionTitle = {
  fontSize: 10,
  letterSpacing: '0.16em',
  textTransform: 'uppercase',
  color: '#737373',
  marginBottom: 12,
};

const detailCardStyle = {
  borderRadius: 10,
  border: '1px solid rgba(255,255,255,0.07)',
  background: 'rgba(255,255,255,0.02)',
  padding: '10px 12px',
};

const detailLabelStyle = {
  fontSize: 11,
  color: '#a3a3a3',
  marginBottom: 6,
};

const detailMessageStyle = {
  fontSize: 12,
  color: '#d4d4d4',
  lineHeight: 1.6,
};
