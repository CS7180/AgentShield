import React from 'react';
import { useEffect, useMemo, useState } from 'react';
import { compareScans, listScans } from './api/client';
import useAuth from './auth/useAuth';

function formatDate(ts) {
  if (!ts) return 'N/A';
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return 'N/A';
  return d.toLocaleString();
}

function formatScore(value) {
  if (typeof value !== 'number') return 'N/A';
  return value.toFixed(2);
}

function formatDelta(value) {
  if (typeof value !== 'number') return 'N/A';
  const sign = value > 0 ? '+' : '';
  return `${sign}${value.toFixed(2)}`;
}

function getDeltaTone(value, invert = false) {
  if (typeof value !== 'number' || value === 0) return '#a3a3a3';
  const improved = invert ? value < 0 : value > 0;
  return improved ? '#34d399' : '#fb7185';
}

const rows = [
  { key: 'critical', label: 'Critical' },
  { key: 'high', label: 'High' },
  { key: 'medium', label: 'Medium' },
  { key: 'low', label: 'Low' },
];

const emptyCompare = {
  overall_score: { trend: 'unknown' },
  severity_counts: {
    base: { critical: 0, high: 0, medium: 0, low: 0 },
    other: { critical: 0, high: 0, medium: 0, low: 0 },
    delta: { critical: 0, high: 0, medium: 0, low: 0 },
  },
};

export default function ReportCompareContent() {
  const { session } = useAuth();
  const accessToken = session?.access_token;

  const [scans, setScans] = useState([]);
  const [loadingScans, setLoadingScans] = useState(true);
  const [scanError, setScanError] = useState('');

  const [baseScanID, setBaseScanID] = useState('');
  const [otherScanID, setOtherScanID] = useState('');

  const [comparing, setComparing] = useState(false);
  const [compareError, setCompareError] = useState('');
  const [compareResult, setCompareResult] = useState(null);

  useEffect(() => {
    if (!accessToken) return;

    let cancelled = false;
    async function load() {
      setLoadingScans(true);
      setScanError('');
      try {
        const resp = await listScans(accessToken);
        const nextScans = Array.isArray(resp.scans) ? resp.scans : [];
        if (cancelled) return;
        setScans(nextScans);
        if (nextScans.length >= 2) {
          setBaseScanID((curr) => curr || nextScans[0].id);
          setOtherScanID((curr) => curr || nextScans[1].id);
        } else if (nextScans.length === 1) {
          setBaseScanID(nextScans[0].id);
          setOtherScanID('');
        }
      } catch (err) {
        if (cancelled) return;
        setScanError(err.message || 'Failed to load scans');
      } finally {
        if (!cancelled) {
          setLoadingScans(false);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [accessToken]);

  const selectedBase = useMemo(
    () => scans.find((scan) => scan.id === baseScanID),
    [scans, baseScanID],
  );
  const selectedOther = useMemo(
    () => scans.find((scan) => scan.id === otherScanID),
    [scans, otherScanID],
  );

  const canCompare = Boolean(baseScanID && otherScanID && baseScanID !== otherScanID && accessToken);
  const compareData = compareResult || emptyCompare;
  const scoreDelta = compareData?.overall_score?.delta;

  async function runCompare() {
    if (!canCompare) return;
    setComparing(true);
    setCompareError('');
    try {
      const data = await compareScans(baseScanID, otherScanID, accessToken);
      setCompareResult(data);
    } catch (err) {
      setCompareResult(null);
      setCompareError(err.message || 'Comparison failed');
    } finally {
      setComparing(false);
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Reports Compare</h1>
        <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
          Select two scans and compare score trend plus severity deltas.
        </p>
      </div>

      <div
        style={{
          borderRadius: 16,
          border: '1px solid rgba(255,255,255,0.06)',
          background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
          padding: 18,
          display: 'grid',
          gap: 16,
          gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
        }}
      >
        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3', marginBottom: 8 }}>Base Scan</div>
          <select
            value={baseScanID}
            onChange={(e) => setBaseScanID(e.target.value)}
            style={{
              width: '100%',
              height: 38,
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.12)',
              background: '#111319',
              color: '#f5f5f5',
              padding: '0 10px',
            }}
            disabled={loadingScans || scans.length === 0}
          >
            <option value="">Select base scan</option>
            {scans.map((scan) => (
              <option key={scan.id} value={scan.id}>
                {scan.id.slice(0, 8)} · {scan.mode} · {scan.status}
              </option>
            ))}
          </select>
        </div>

        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3', marginBottom: 8 }}>Compare Against</div>
          <select
            value={otherScanID}
            onChange={(e) => setOtherScanID(e.target.value)}
            style={{
              width: '100%',
              height: 38,
              borderRadius: 10,
              border: '1px solid rgba(255,255,255,0.12)',
              background: '#111319',
              color: '#f5f5f5',
              padding: '0 10px',
            }}
            disabled={loadingScans || scans.length === 0}
          >
            <option value="">Select comparison scan</option>
            {scans.map((scan) => (
              <option key={scan.id} value={scan.id}>
                {scan.id.slice(0, 8)} · {scan.mode} · {scan.status}
              </option>
            ))}
          </select>
        </div>

        <div style={{ display: 'flex', alignItems: 'end' }}>
          <button
            type="button"
            onClick={runCompare}
            disabled={!canCompare || comparing}
            style={{
              width: '100%',
              height: 38,
              borderRadius: 10,
              border: '1px solid rgba(16,185,129,0.35)',
              background: canCompare
                ? 'linear-gradient(135deg, rgba(16,185,129,0.2), rgba(59,130,246,0.12))'
                : 'rgba(255,255,255,0.04)',
              color: canCompare ? '#a7f3d0' : '#737373',
              cursor: canCompare ? 'pointer' : 'not-allowed',
            }}
          >
            {comparing ? 'Comparing...' : 'Compare'}
          </button>
        </div>
      </div>

      {loadingScans && <div style={{ color: '#a3a3a3', fontSize: 12 }}>Loading scans...</div>}
      {!loadingScans && scans.length < 2 && (
        <div style={{ color: '#fbbf24', fontSize: 12 }}>At least two scans are required for comparison.</div>
      )}
      {scanError && <div style={{ color: '#fb7185', fontSize: 12 }}>{scanError}</div>}
      {compareError && <div style={{ color: '#fb7185', fontSize: 12 }}>{compareError}</div>}

      <div
        style={{
          borderRadius: 16,
          border: '1px solid rgba(255,255,255,0.06)',
          background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
          padding: 20,
          display: 'grid',
          gap: 16,
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
        }}
      >
        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3' }}>Base Score</div>
          <div style={{ fontSize: 30, fontWeight: 700, marginTop: 6 }}>
            {formatScore(compareData?.overall_score?.base)}
          </div>
        </div>
        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3' }}>Compared Score</div>
          <div style={{ fontSize: 30, fontWeight: 700, marginTop: 6 }}>
            {formatScore(compareData?.overall_score?.other)}
          </div>
        </div>
        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3' }}>Score Delta</div>
          <div style={{ fontSize: 30, fontWeight: 700, marginTop: 6, color: getDeltaTone(scoreDelta, false) }}>
            {formatDelta(scoreDelta)}
          </div>
        </div>
        <div>
          <div style={{ fontSize: 11, color: '#a3a3a3' }}>Trend</div>
          <div style={{ fontSize: 22, fontWeight: 700, marginTop: 12, textTransform: 'capitalize' }}>
            {compareData?.overall_score?.trend || 'unknown'}
          </div>
        </div>
      </div>

      <div
        style={{
          borderRadius: 16,
          border: '1px solid rgba(255,255,255,0.06)',
          background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
          padding: 20,
        }}
      >
        <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 14 }}>Severity Delta</div>
        <div style={{ display: 'grid', gap: 10 }}>
          {rows.map(({ key, label }) => {
            const base = compareData?.severity_counts?.base?.[key];
            const other = compareData?.severity_counts?.other?.[key];
            const delta = compareData?.severity_counts?.delta?.[key];
            return (
              <div
                key={key}
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr 1fr 1fr',
                  gap: 8,
                  fontSize: 12,
                  padding: '10px 12px',
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.06)',
                  background: 'rgba(255,255,255,0.02)',
                }}
              >
                <div style={{ color: '#e5e5e5', fontWeight: 600 }}>{label}</div>
                <div style={{ color: '#a3a3a3' }}>Base: {base ?? 'N/A'}</div>
                <div style={{ color: '#a3a3a3' }}>Other: {other ?? 'N/A'}</div>
                <div style={{ color: getDeltaTone(delta, true) }}>
                  Delta: {typeof delta === 'number' && delta > 0 ? `+${delta}` : delta ?? 'N/A'}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div style={{ color: '#737373', fontSize: 11, lineHeight: 1.6 }}>
        <div>Base scan: {selectedBase ? `${selectedBase.id} (${selectedBase.mode}, ${selectedBase.status})` : 'N/A'}</div>
        <div>Compared scan: {selectedOther ? `${selectedOther.id} (${selectedOther.mode}, ${selectedOther.status})` : 'N/A'}</div>
        <div>
          Generated at:{' '}
          {compareData?.generated_at ? formatDate(compareData.generated_at) : 'N/A'}
        </div>
      </div>
    </div>
  );
}
