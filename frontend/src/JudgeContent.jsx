import { useMemo, useState } from 'react';
import { calibrateJudge, getJudgeCalibrationReport } from './api/client';
import useAuth from './auth/useAuth';

const DEFAULT_SAMPLES = [
  {
    attack_type: 'prompt_injection',
    expected_success: true,
    predicted_success: true,
    expected_severity: 'high',
    predicted_severity: 'high',
    human_confidence: 0.92,
    judge_confidence: 0.88,
  },
  {
    attack_type: 'data_leakage',
    expected_success: false,
    predicted_success: false,
    expected_severity: 'low',
    predicted_severity: 'low',
    human_confidence: 0.2,
    judge_confidence: 0.25,
  },
];

function formatMaybe(value, digits = 2) {
  return typeof value === 'number' ? value.toFixed(digits) : 'N/A';
}

export default function JudgeContent() {
  const { session } = useAuth();
  const token = session?.access_token;

  const [samplesText, setSamplesText] = useState(JSON.stringify(DEFAULT_SAMPLES, null, 2));
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');

  const byAttackRows = useMemo(() => {
    if (!report?.by_attack_type) return [];
    return Object.entries(report.by_attack_type);
  }, [report]);

  async function handleLoadLatest() {
    if (!token) {
      setError('Missing access token. Please log in again.');
      return;
    }
    setLoading(true);
    setError('');
    setInfo('');
    try {
      const latest = await getJudgeCalibrationReport(token);
      setReport(latest);
      setInfo('Loaded latest calibration report.');
    } catch (err) {
      setError(err.message || 'Failed to load calibration report.');
    } finally {
      setLoading(false);
    }
  }

  async function handleCalibrate() {
    if (!token) {
      setError('Missing access token. Please log in again.');
      return;
    }

    let samples;
    try {
      samples = JSON.parse(samplesText);
      if (!Array.isArray(samples)) {
        throw new Error('Samples JSON must be an array.');
      }
    } catch (err) {
      setError(err.message || 'Invalid samples JSON.');
      return;
    }

    setLoading(true);
    setError('');
    setInfo('');
    try {
      const next = await calibrateJudge(samples, token);
      setReport(next);
      setInfo('Calibration completed and persisted.');
    } catch (err) {
      setError(err.message || 'Calibration failed.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Judge Calibration</h1>
        <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
          Run calibration with your samples, then inspect the latest report.
        </p>
      </div>

      {error && <div style={{ color: '#fb7185', fontSize: 12 }}>{error}</div>}
      {info && <div style={{ color: '#34d399', fontSize: 12 }}>{info}</div>}

      <div style={panelStyle}>
        <div style={{ fontSize: 10, textTransform: 'uppercase', letterSpacing: '0.16em', color: '#737373', marginBottom: 10 }}>
          Samples JSON
        </div>
        <textarea
          value={samplesText}
          onChange={(e) => setSamplesText(e.target.value)}
          spellCheck={false}
          style={{
            width: '100%',
            minHeight: 220,
            borderRadius: 10,
            border: '1px solid rgba(255,255,255,0.10)',
            background: 'rgba(255,255,255,0.03)',
            color: '#f5f5f5',
            padding: 12,
            fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
            fontSize: 12,
            lineHeight: 1.5,
          }}
        />

        <div style={{ display: 'flex', gap: 10, marginTop: 12 }}>
          <button
            type="button"
            onClick={handleCalibrate}
            disabled={loading}
            style={primaryButtonStyle(loading)}
          >
            {loading ? 'Running...' : 'Run Calibration'}
          </button>

          <button
            type="button"
            onClick={handleLoadLatest}
            disabled={loading}
            style={ghostButtonStyle(loading)}
          >
            Load Latest Report
          </button>
        </div>
      </div>

      {report && (
        <div style={panelStyle}>
          <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 10 }}>Overall Metrics</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 10 }}>
            <MetricCard label="Samples" value={String(report.sample_count)} />
            <MetricCard label="F1" value={formatMaybe(report.overall?.f1)} />
            <MetricCard label="Precision" value={formatMaybe(report.overall?.precision)} />
            <MetricCard label="Recall" value={formatMaybe(report.overall?.recall)} />
            <MetricCard label="Accuracy" value={formatMaybe(report.overall?.accuracy)} />
            <MetricCard label="Kendall Tau" value={formatMaybe(report.kendall_tau)} />
            <MetricCard
              label="Critical Precision"
              value={formatMaybe(report.critical_severity_precision)}
            />
          </div>

          <div style={{ marginTop: 16, fontSize: 14, fontWeight: 600 }}>By Attack Type</div>
          <div style={{ marginTop: 8, display: 'grid', gap: 8 }}>
            {byAttackRows.length === 0 && <div style={{ color: '#737373', fontSize: 12 }}>No per-attack metrics.</div>}
            {byAttackRows.map(([attackType, metrics]) => (
              <div
                key={attackType}
                style={{
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.07)',
                  background: 'rgba(255,255,255,0.02)',
                  padding: '10px 12px',
                  display: 'grid',
                  gridTemplateColumns: '1.2fr 1fr 1fr 1fr 1fr',
                  gap: 8,
                }}
              >
                <div style={{ color: '#f5f5f5', fontSize: 12 }}>{attackType}</div>
                <div style={{ color: '#a3a3a3', fontSize: 12 }}>F1 {formatMaybe(metrics.f1)}</div>
                <div style={{ color: '#a3a3a3', fontSize: 12 }}>P {formatMaybe(metrics.precision)}</div>
                <div style={{ color: '#a3a3a3', fontSize: 12 }}>R {formatMaybe(metrics.recall)}</div>
                <div style={{ color: '#a3a3a3', fontSize: 12 }}>Acc {formatMaybe(metrics.accuracy)}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function MetricCard({ label, value }) {
  return (
    <div
      style={{
        borderRadius: 10,
        border: '1px solid rgba(255,255,255,0.07)',
        background: 'rgba(255,255,255,0.02)',
        padding: '10px 12px',
      }}
    >
      <div style={{ fontSize: 11, color: '#a3a3a3' }}>{label}</div>
      <div style={{ marginTop: 6, fontSize: 20, color: '#f5f5f5', fontWeight: 700 }}>{value}</div>
    </div>
  );
}

const panelStyle = {
  borderRadius: 14,
  border: '1px solid rgba(255,255,255,0.06)',
  background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
  padding: 16,
};

const primaryButtonStyle = (disabled) => ({
  borderRadius: 10,
  border: '1px solid rgba(16,185,129,0.35)',
  background: disabled
    ? 'rgba(255,255,255,0.04)'
    : 'linear-gradient(135deg, rgba(16,185,129,0.18), rgba(59,130,246,0.14))',
  color: disabled ? '#737373' : '#a7f3d0',
  height: 38,
  padding: '0 14px',
  cursor: disabled ? 'not-allowed' : 'pointer',
});

const ghostButtonStyle = (disabled) => ({
  borderRadius: 10,
  border: '1px solid rgba(255,255,255,0.12)',
  background: 'rgba(255,255,255,0.03)',
  color: disabled ? '#737373' : '#d4d4d4',
  height: 38,
  padding: '0 14px',
  cursor: disabled ? 'not-allowed' : 'pointer',
});
