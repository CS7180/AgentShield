import { useEffect, useMemo, useRef, useState } from 'react';
import {
  API_BASE,
  getScan,
  listAttackResults,
  listScans,
  startScan,
  stopScan,
} from './api/client';
import useAuth from './auth/useAuth';

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

function statusColor(status) {
  switch (status) {
    case 'running':
      return '#fbbf24';
    case 'completed':
      return '#34d399';
    case 'failed':
      return '#fb7185';
    case 'stopped':
      return '#a3a3a3';
    case 'queued':
      return '#60a5fa';
    default:
      return '#737373';
  }
}

function attackLabel(key) {
  return {
    prompt_injection: 'Prompt Injection',
    jailbreak: 'Jailbreak',
    data_leakage: 'Data Leakage',
    constraint_drift: 'Constraint Drift',
  }[key] || key;
}

function toFeedTimestamp(ts) {
  if (!ts) return new Date().toLocaleTimeString();
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return new Date().toLocaleTimeString();
  return d.toLocaleTimeString();
}

export default function ScanMonitorContent() {
  const { session } = useAuth();
  const token = session?.access_token;

  const [scans, setScans] = useState([]);
  const [selectedScanID, setSelectedScanID] = useState('');
  const [scan, setScan] = useState(null);
  const [results, setResults] = useState([]);
  const [feed, setFeed] = useState([]);

  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState('');

  const wsRef = useRef(null);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;

    async function loadScans() {
      try {
        const resp = await listScans(token, { limit: 100, offset: 0 });
        const ordered = [...(resp.scans || [])].sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        );
        if (cancelled) return;
        setScans(ordered);

        if (!selectedScanID && ordered.length > 0) {
          const running = ordered.find((item) => item.status === 'running');
          setSelectedScanID((running || ordered[0]).id);
        }
      } catch (err) {
        if (!cancelled) setError(err.message || 'Failed to load scans');
      }
    }

    loadScans();
    const timer = setInterval(loadScans, 5000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [token, selectedScanID]);

  useEffect(() => {
    if (!token || !selectedScanID) return;
    let cancelled = false;

    async function loadDetail() {
      setLoading(true);
      setError('');
      try {
        const [scanResp, resultResp] = await Promise.all([
          getScan(selectedScanID, token),
          listAttackResults(selectedScanID, token, { limit: 200, offset: 0 }),
        ]);
        if (cancelled) return;

        setScan(scanResp);
        setResults(resultResp.results || []);

        const nextFeed = (resultResp.results || [])
          .slice()
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 20)
          .map((item) => ({
            id: item.id,
            time: toFeedTimestamp(item.created_at),
            type: item.attack_success ? 'red' : 'blue',
            severity: item.severity || (item.attack_success ? 'high' : 'low'),
            title: attackLabel(item.attack_type),
            message: item.attack_success
              ? `Attack succeeded (${item.owasp_category || 'N/A'})`
              : 'Attack blocked or ineffective',
          }));
        setFeed(nextFeed);
      } catch (err) {
        if (!cancelled) setError(err.message || 'Failed to load scan detail');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    loadDetail();
    const timer = setInterval(loadDetail, 3000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [selectedScanID, token]);

  useEffect(() => {
    if (!token || !selectedScanID) return;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    const wsBase = API_BASE.replace(/^http/, 'ws');
    const socket = new WebSocket(
      `${wsBase}/ws/scans/${selectedScanID}/status?token=${encodeURIComponent(token)}`,
    );

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        const message = JSON.stringify(payload.data || payload);
        setFeed((prev) => [
          {
            id: `ws-${Date.now()}`,
            time: new Date().toLocaleTimeString(),
            type: 'judge',
            severity: 'info',
            title: payload.topic || 'ws-event',
            message,
          },
          ...prev,
        ].slice(0, 30));
      } catch {
        setFeed((prev) => [
          {
            id: `ws-${Date.now()}`,
            time: new Date().toLocaleTimeString(),
            type: 'judge',
            severity: 'info',
            title: 'ws-event',
            message: String(event.data),
          },
          ...prev,
        ].slice(0, 30));
      }
    };

    wsRef.current = socket;
    return () => {
      socket.close();
      wsRef.current = null;
    };
  }, [selectedScanID, token]);

  const redAgents = useMemo(() => {
    if (!scan) return [];

    return (scan.attack_types || []).map((attackType) => {
      const subset = results.filter((r) => r.attack_type === attackType);
      const success = subset.filter((r) => r.attack_success).length;

      let status = 'pending';
      if (scan.status === 'running') {
        status = subset.length > 0 ? 'running' : 'pending';
      } else if (scan.status === 'completed' || scan.status === 'failed' || scan.status === 'stopped') {
        status = subset.length > 0 ? 'done' : 'pending';
      }

      return {
        key: attackType,
        label: attackLabel(attackType),
        tested: `${subset.length}/1`,
        success,
        status,
      };
    });
  }, [scan, results]);

  const defenseSummary = useMemo(() => {
    const blocked = results.filter((r) => r.defense_intercepted === true).length;
    const successful = results.filter((r) => r.attack_success).length;
    const confidenceList = results
      .map((r) => r.judge_confidence)
      .filter((v) => typeof v === 'number');
    const avgConfidence = confidenceList.length
      ? confidenceList.reduce((sum, v) => sum + v, 0) / confidenceList.length
      : null;

    return { blocked, successful, avgConfidence };
  }, [results]);

  const progress = useMemo(() => {
    if (!scan) return 0;
    if (scan.status === 'completed') return 100;
    if (scan.status === 'failed' || scan.status === 'stopped') return 100;

    const total = Math.max(1, (scan.attack_types || []).length);
    const pct = Math.round((results.length / total) * 100);
    return Math.max(scan.status === 'running' ? 10 : 0, Math.min(95, pct));
  }, [scan, results]);

  async function handleStart() {
    if (!scan || !token) return;
    setActionLoading(true);
    setError('');
    try {
      await startScan(scan.id, token);
      const refreshed = await getScan(scan.id, token);
      setScan(refreshed);
    } catch (err) {
      setError(err.message || 'Failed to start scan');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleStop() {
    if (!scan || !token) return;
    setActionLoading(true);
    setError('');
    try {
      await stopScan(scan.id, token);
      const refreshed = await getScan(scan.id, token);
      setScan(refreshed);
    } catch (err) {
      setError(err.message || 'Failed to stop scan');
    } finally {
      setActionLoading(false);
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Scan Monitoring</h1>
        <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
          Live status, attack results, and event stream.
        </p>
      </div>

      <div style={panelStyle}>
        <div style={{ display: 'grid', gap: 10, gridTemplateColumns: '2fr 1fr auto auto' }}>
          <select
            value={selectedScanID}
            onChange={(e) => setSelectedScanID(e.target.value)}
            style={inputStyle}
          >
            {scans.length === 0 && <option value="">No scans found</option>}
            {scans.map((item) => (
              <option key={item.id} value={item.id}>
                {item.id.slice(0, 8)} · {item.mode} · {item.status}
              </option>
            ))}
          </select>

          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: statusColor(scan?.status), fontSize: 12 }}>
            <Dot color={statusColor(scan?.status)} size={7} />
            {scan?.status || 'N/A'}
          </div>

          <button
            type="button"
            onClick={handleStart}
            disabled={!scan || actionLoading || scan.status === 'running'}
            style={buttonStyle('start', !scan || actionLoading || scan.status === 'running')}
          >
            Start
          </button>
          <button
            type="button"
            onClick={handleStop}
            disabled={!scan || actionLoading || scan.status !== 'running'}
            style={buttonStyle('stop', !scan || actionLoading || scan.status !== 'running')}
          >
            Stop
          </button>
        </div>

        {scan && (
          <div style={{ marginTop: 10, fontSize: 12, color: '#a3a3a3' }}>
            Target: {scan.target_endpoint} | Mode: {scan.mode}
          </div>
        )}
      </div>

      {error && <div style={{ color: '#fb7185', fontSize: 12 }}>{error}</div>}
      {loading && <div style={{ color: '#a3a3a3', fontSize: 12 }}>Loading scan data...</div>}

      <div style={panelStyle}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6, fontSize: 11, color: '#a3a3a3' }}>
          <span>Progress</span>
          <span>{progress}%</span>
        </div>
        <div style={{ height: 8, borderRadius: 6, background: 'rgba(255,255,255,0.06)', overflow: 'hidden' }}>
          <div
            style={{
              height: '100%',
              width: `${progress}%`,
              background: 'linear-gradient(90deg, #d946ef, #3b82f6)',
              transition: 'width 0.5s ease',
            }}
          />
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: 14 }}>
        <div style={panelStyle}>
          <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 10 }}>Red Team Agents</div>
          {redAgents.length === 0 && <div style={{ color: '#737373', fontSize: 12 }}>No attack types configured.</div>}
          <div style={{ display: 'grid', gap: 8 }}>
            {redAgents.map((agent) => (
              <div
                key={agent.key}
                style={{
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.07)',
                  background: 'rgba(255,255,255,0.02)',
                  padding: '10px 12px',
                  display: 'grid',
                  gridTemplateColumns: '1.5fr 1fr 1fr',
                  gap: 10,
                }}
              >
                <div style={{ color: '#f5f5f5', fontSize: 12 }}>{agent.label}</div>
                <div style={{ color: '#a3a3a3', fontSize: 12 }}>tested {agent.tested}</div>
                <div style={{ color: agent.success > 0 ? '#fb7185' : '#34d399', fontSize: 12 }}>
                  success {agent.success}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div style={panelStyle}>
          <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 10 }}>Defense/Judge Summary</div>
          <div style={{ display: 'grid', gap: 10 }}>
            <MetricRow label="Blocked by defense" value={String(defenseSummary.blocked)} />
            <MetricRow label="Successful attacks" value={String(defenseSummary.successful)} tone="#fb7185" />
            <MetricRow
              label="Avg judge confidence"
              value={
                defenseSummary.avgConfidence == null
                  ? 'N/A'
                  : defenseSummary.avgConfidence.toFixed(2)
              }
              tone="#60a5fa"
            />
          </div>
        </div>
      </div>

      <div style={panelStyle}>
        <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 10 }}>Live Activity Feed</div>
        {feed.length === 0 && <div style={{ color: '#737373', fontSize: 12 }}>No events yet.</div>}
        <div style={{ display: 'grid', gap: 8 }}>
          {feed.slice(0, 20).map((item) => (
            <div
              key={item.id}
              style={{
                borderRadius: 10,
                border: '1px solid rgba(255,255,255,0.07)',
                background: 'rgba(255,255,255,0.02)',
                padding: '10px 12px',
                display: 'grid',
                gridTemplateColumns: '70px 120px 1fr',
                gap: 8,
                alignItems: 'start',
                fontSize: 12,
              }}
            >
              <div style={{ color: '#737373', fontFamily: 'monospace' }}>{item.time}</div>
              <div style={{ color: '#d4d4d4' }}>{item.title}</div>
              <div style={{ color: '#a3a3a3', wordBreak: 'break-word' }}>{item.message}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function MetricRow({ label, value, tone = '#34d399' }) {
  return (
    <div
      style={{
        borderRadius: 10,
        border: '1px solid rgba(255,255,255,0.07)',
        background: 'rgba(255,255,255,0.02)',
        padding: '10px 12px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <span style={{ fontSize: 12, color: '#a3a3a3' }}>{label}</span>
      <span style={{ fontSize: 13, fontWeight: 600, color: tone }}>{value}</span>
    </div>
  );
}

const panelStyle = {
  borderRadius: 14,
  border: '1px solid rgba(255,255,255,0.06)',
  background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
  padding: 16,
};

const inputStyle = {
  borderRadius: 10,
  border: '1px solid rgba(255,255,255,0.10)',
  background: 'rgba(255,255,255,0.03)',
  color: '#f5f5f5',
  height: 38,
  padding: '0 12px',
};

function buttonStyle(kind, disabled) {
  const base = {
    height: 38,
    borderRadius: 10,
    padding: '0 14px',
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.6 : 1,
  };

  if (kind === 'start') {
    return {
      ...base,
      border: '1px solid rgba(16,185,129,0.35)',
      background: 'rgba(16,185,129,0.14)',
      color: '#6ee7b7',
    };
  }

  return {
    ...base,
    border: '1px solid rgba(251,113,133,0.35)',
    background: 'rgba(244,63,94,0.10)',
    color: '#fda4af',
  };
}
