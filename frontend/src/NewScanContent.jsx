import React from 'react';
import { useMemo, useState } from 'react';
import { createScan, startScan } from './api/client';
import useAuth from './auth/useAuth';

const MODES = [
  {
    key: 'red_team',
    label: 'Red Team',
    desc: 'Attack only, no defense.',
  },
  {
    key: 'blue_team',
    label: 'Blue Team',
    desc: 'Defense posture verification.',
  },
  {
    key: 'adversarial',
    label: 'Adversarial',
    desc: 'Red + blue together.',
  },
];

const ATTACK_OPTIONS = [
  { key: 'prompt_injection', label: 'Prompt Injection' },
  { key: 'jailbreak', label: 'Jailbreak' },
  { key: 'data_leakage', label: 'Data Leakage' },
  { key: 'constraint_drift', label: 'Constraint Drift' },
];

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

function isValidHTTPS(url) {
  try {
    const parsed = new URL(url);
    return parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

export default function NewScanContent() {
  const { session } = useAuth();
  const token = session?.access_token;

  const [targetEndpoint, setTargetEndpoint] = useState('');
  const [mode, setMode] = useState('red_team');
  const [attackSelections, setAttackSelections] = useState({
    prompt_injection: true,
    jailbreak: true,
    data_leakage: true,
    constraint_drift: false,
  });
  const [autoStart, setAutoStart] = useState(true);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const [createdScan, setCreatedScan] = useState(null);

  const selectedAttackTypes = useMemo(
    () => ATTACK_OPTIONS.filter((a) => attackSelections[a.key]).map((a) => a.key),
    [attackSelections],
  );

  async function handleTestEndpoint() {
    if (!isValidHTTPS(targetEndpoint)) {
      setError('Target endpoint must be a valid HTTPS URL.');
      return;
    }
    setError('');
    setInfo('Endpoint format looks valid.');
  }

  async function handleSubmit() {
    if (!token) {
      setError('Missing access token. Please log in again.');
      return;
    }
    if (!isValidHTTPS(targetEndpoint)) {
      setError('Target endpoint must be a valid HTTPS URL.');
      return;
    }
    if (selectedAttackTypes.length === 0) {
      setError('Select at least one attack type.');
      return;
    }

    setSubmitting(true);
    setError('');
    setInfo('');

    try {
      const created = await createScan(
        {
          target_endpoint: targetEndpoint,
          mode,
          attack_types: selectedAttackTypes,
        },
        token,
      );

      let status = created.status;
      let message = 'Scan created.';
      if (autoStart) {
        const started = await startScan(created.id, token);
        status = started.status;
        message = started.message || 'Scan created and start requested.';
      }

      setCreatedScan({ ...created, status });
      setInfo(message);
    } catch (err) {
      setError(err.message || 'Failed to create/start scan.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Create Scan</h1>
        <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
          Configure target and attack profile, then create or start a real scan.
        </p>
      </div>

      {error && <div style={{ color: '#fb7185', fontSize: 12 }}>{error}</div>}
      {info && <div style={{ color: '#34d399', fontSize: 12 }}>{info}</div>}

      <div style={panelStyle}>
        <div style={sectionTitle}>Target Endpoint</div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
          <input
            value={targetEndpoint}
            onChange={(e) => setTargetEndpoint(e.target.value.trim())}
            placeholder="https://example.com/v1/chat"
            style={inputStyle}
          />
          <button type="button" onClick={handleTestEndpoint} style={ghostButtonStyle}>
            Validate URL
          </button>
        </div>
      </div>

      <div style={panelStyle}>
        <div style={sectionTitle}>Mode</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 10 }}>
          {MODES.map((item) => {
            const selected = mode === item.key;
            return (
              <button
                key={item.key}
                type="button"
                onClick={() => setMode(item.key)}
                style={{
                  textAlign: 'left',
                  borderRadius: 12,
                  border: selected
                    ? '1px solid rgba(59,130,246,0.45)'
                    : '1px solid rgba(255,255,255,0.08)',
                  background: selected ? 'rgba(59,130,246,0.12)' : 'rgba(255,255,255,0.02)',
                  padding: 12,
                  color: '#f5f5f5',
                  cursor: 'pointer',
                }}
              >
                <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                  <Dot color={selected ? '#60a5fa' : '#737373'} size={7} />
                  <span style={{ fontSize: 13, fontWeight: 600 }}>{item.label}</span>
                </div>
                <div style={{ fontSize: 11, color: '#9ca3af' }}>{item.desc}</div>
              </button>
            );
          })}
        </div>
      </div>

      <div style={panelStyle}>
        <div style={sectionTitle}>Attack Types</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 10 }}>
          {ATTACK_OPTIONS.map((option) => {
            const selected = attackSelections[option.key];
            return (
              <button
                key={option.key}
                type="button"
                onClick={() =>
                  setAttackSelections((prev) => ({
                    ...prev,
                    [option.key]: !prev[option.key],
                  }))
                }
                style={{
                  textAlign: 'left',
                  borderRadius: 12,
                  border: selected
                    ? '1px solid rgba(251,113,133,0.45)'
                    : '1px solid rgba(255,255,255,0.08)',
                  background: selected ? 'rgba(244,63,94,0.10)' : 'rgba(255,255,255,0.02)',
                  padding: 12,
                  color: '#f5f5f5',
                  cursor: 'pointer',
                }}
              >
                <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                  <Dot color={selected ? '#fb7185' : '#737373'} size={7} />
                  <span style={{ fontSize: 13, fontWeight: 600 }}>{option.label}</span>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      <div style={panelStyle}>
        <label style={{ display: 'inline-flex', alignItems: 'center', gap: 8, fontSize: 13, color: '#d4d4d4' }}>
          <input
            type="checkbox"
            checked={autoStart}
            onChange={(e) => setAutoStart(e.target.checked)}
          />
          Auto-start scan after creation
        </label>

        <div style={{ marginTop: 14 }}>
          <button
            type="button"
            disabled={submitting}
            onClick={handleSubmit}
            style={primaryButtonStyle(submitting)}
          >
            {submitting ? 'Submitting...' : autoStart ? 'Create + Start Scan' : 'Create Scan'}
          </button>
        </div>
      </div>

      {createdScan && (
        <div style={panelStyle}>
          <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 8 }}>Latest Scan</div>
          <div style={{ fontSize: 12, color: '#d4d4d4', lineHeight: 1.8 }}>
            <div>ID: {createdScan.id}</div>
            <div>Status: {createdScan.status}</div>
            <div>Mode: {createdScan.mode}</div>
            <div>Target: {createdScan.target_endpoint}</div>
          </div>
        </div>
      )}
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

const inputStyle = {
  flex: 1,
  minWidth: 280,
  borderRadius: 10,
  border: '1px solid rgba(255,255,255,0.10)',
  background: 'rgba(255,255,255,0.03)',
  color: '#f5f5f5',
  height: 40,
  padding: '0 12px',
};

const ghostButtonStyle = {
  borderRadius: 10,
  border: '1px solid rgba(255,255,255,0.12)',
  background: 'rgba(255,255,255,0.03)',
  color: '#d4d4d4',
  height: 40,
  padding: '0 14px',
  cursor: 'pointer',
};

const primaryButtonStyle = (disabled) => ({
  borderRadius: 10,
  border: '1px solid rgba(16,185,129,0.35)',
  background: disabled
    ? 'rgba(255,255,255,0.04)'
    : 'linear-gradient(135deg, rgba(16,185,129,0.18), rgba(59,130,246,0.14))',
  color: disabled ? '#737373' : '#a7f3d0',
  height: 42,
  padding: '0 18px',
  cursor: disabled ? 'not-allowed' : 'pointer',
});
