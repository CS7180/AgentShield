import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getJudgeCalibrationReport, getScanReport, listScans } from './api/client';
import useAuth from './auth/useAuth';

const ATTACK_TYPES = [
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

function formatScore(value) {
  if (typeof value !== 'number') return 'N/A';
  return value.toFixed(2);
}

function statusTone(status) {
  switch (status) {
    case 'completed':
      return '#34d399';
    case 'running':
      return '#fbbf24';
    case 'failed':
      return '#fb7185';
    case 'stopped':
      return '#a3a3a3';
    default:
      return '#737373';
  }
}

function parseScorecard(raw) {
  if (!raw) return null;
  if (typeof raw === 'object') return raw;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export default function DashboardContent() {
  const navigate = useNavigate();
  const { session } = useAuth();
  const token = session?.access_token;

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [scans, setScans] = useState([]);
  const [reportsByScanID, setReportsByScanID] = useState({});
  const [judgeTau, setJudgeTau] = useState(null);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError('');
      try {
        const scansResp = await listScans(token, { limit: 100, offset: 0 });
        const orderedScans = [...(scansResp.scans || [])].sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        );

        const reportEntries = await Promise.all(
          orderedScans.slice(0, 20).map(async (scan) => {
            try {
              const report = await getScanReport(scan.id, token);
              return [scan.id, report];
            } catch {
              return [scan.id, null];
            }
          }),
        );

        let nextJudgeTau = null;
        try {
          const judge = await getJudgeCalibrationReport(token);
          if (typeof judge?.kendall_tau === 'number') {
            nextJudgeTau = judge.kendall_tau;
          }
        } catch {
          nextJudgeTau = null;
        }

        if (cancelled) return;
        setScans(orderedScans);
        setReportsByScanID(Object.fromEntries(reportEntries));
        setJudgeTau(nextJudgeTau);
      } catch (err) {
        if (cancelled) return;
        setError(err.message || 'Failed to load dashboard data');
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [token]);

  const stats = useMemo(() => {
    const total = scans.length;

    let critical = 0;
    let scoreSum = 0;
    let scoreCount = 0;
    for (const scan of scans) {
      const report = reportsByScanID[scan.id];
      if (!report) continue;
      critical += Number(report.critical_count || 0);
      if (typeof report.overall_score === 'number') {
        scoreSum += report.overall_score;
        scoreCount += 1;
      }
    }

    const avgDefense = scoreCount > 0 ? scoreSum / scoreCount : null;
    return {
      total,
      critical,
      avgDefense,
      judgeTau,
    };
  }, [scans, reportsByScanID, judgeTau]);

  const recentScans = useMemo(() => scans.slice(0, 6), [scans]);

  const coverage = useMemo(() => {
    const attempted = Object.fromEntries(ATTACK_TYPES.map((t) => [t.key, 0]));
    const successful = Object.fromEntries(ATTACK_TYPES.map((t) => [t.key, 0]));

    for (const scan of scans) {
      const report = reportsByScanID[scan.id];
      const scorecard = parseScorecard(report?.owasp_scorecard);

      for (const attackType of scan.attack_types || []) {
        if (!(attackType in attempted)) continue;
        attempted[attackType] += 1;

        if (scorecard && scorecard[attackType] && Number(scorecard[attackType].successful || 0) > 0) {
          successful[attackType] += 1;
        }
      }
    }

    return ATTACK_TYPES.map((type) => {
      const a = attempted[type.key];
      const s = successful[type.key];
      const value = a > 0 ? Math.round((s / a) * 100) : 0;
      return { ...type, value, attempted: a, successful: s };
    });
  }, [scans, reportsByScanID]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>Dashboard</h1>
          <p style={{ fontSize: 12, color: '#737373', marginTop: 6 }}>
            Real-time scan summary from API data.
          </p>
        </div>
        <button
          type="button"
          onClick={() => navigate('/scans')}
          style={{
            borderRadius: 10,
            border: '1px solid rgba(16,185,129,0.3)',
            background: 'rgba(16,185,129,0.12)',
            padding: '8px 16px',
            color: '#6ee7b7',
            cursor: 'pointer',
          }}
        >
          + New Scan
        </button>
      </div>

      {error && <div style={{ color: '#fb7185', fontSize: 12 }}>{error}</div>}
      {loading && <div style={{ color: '#a3a3a3', fontSize: 12 }}>Loading dashboard...</div>}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 12 }}>
        <div style={cardStyle}>
          <div style={metricLabel}>TOTAL SCANS</div>
          <div style={metricValue}>{stats.total}</div>
        </div>
        <div style={cardStyle}>
          <div style={metricLabel}>CRITICAL FINDINGS</div>
          <div style={{ ...metricValue, color: '#fb7185' }}>{stats.critical}</div>
        </div>
        <div style={cardStyle}>
          <div style={metricLabel}>AVG DEFENSE SCORE</div>
          <div style={{ ...metricValue, color: '#34d399' }}>
            {stats.avgDefense == null ? 'N/A' : `${stats.avgDefense.toFixed(2)}%`}
          </div>
        </div>
        <div style={cardStyle}>
          <div style={metricLabel}>JUDGE TAU</div>
          <div style={metricValue}>{stats.judgeTau == null ? 'N/A' : stats.judgeTau.toFixed(2)}</div>
        </div>
      </div>

      <div style={panelStyle}>
        <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 12 }}>Recent Scans</div>
        {recentScans.length === 0 && <div style={{ color: '#737373', fontSize: 12 }}>No scans found.</div>}
        {recentScans.map((scan) => {
          const report = reportsByScanID[scan.id];
          const tone = statusTone(scan.status);
          return (
            <div
              key={scan.id}
              style={{
                padding: '12px 10px',
                borderBottom: '1px solid rgba(255,255,255,0.05)',
                display: 'grid',
                gridTemplateColumns: '1.8fr 1fr 1fr 0.8fr',
                gap: 10,
                alignItems: 'center',
              }}
            >
              <div>
                <div style={{ fontSize: 13, color: '#f5f5f5' }}>{scan.target_endpoint}</div>
                <div style={{ fontSize: 11, color: '#737373', marginTop: 4 }}>{scan.id}</div>
              </div>
              <div style={{ fontSize: 12, color: '#d4d4d4' }}>{scan.mode}</div>
              <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: tone, fontSize: 12 }}>
                <Dot color={tone} size={6} />
                {scan.status}
              </div>
              <div style={{ textAlign: 'right', fontSize: 12, color: '#e5e5e5' }}>
                {formatScore(report?.overall_score)}
              </div>
            </div>
          );
        })}
      </div>

      <div style={panelStyle}>
        <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 12 }}>Attack Success Coverage</div>
        <div style={{ display: 'grid', gap: 10 }}>
          {coverage.map((item) => (
            <div key={item.key}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 6 }}>
                <span style={{ color: '#d4d4d4' }}>{item.label}</span>
                <span style={{ color: '#a3a3a3' }}>
                  {item.value}% ({item.successful}/{item.attempted})
                </span>
              </div>
              <div style={{ height: 8, borderRadius: 6, background: 'rgba(255,255,255,0.06)', overflow: 'hidden' }}>
                <div
                  style={{
                    height: '100%',
                    width: `${item.value}%`,
                    background: 'linear-gradient(90deg, #22c55e, #06b6d4)',
                    transition: 'width 0.4s ease',
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

const cardStyle = {
  borderRadius: 14,
  border: '1px solid rgba(255,255,255,0.06)',
  background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
  padding: 16,
};

const metricLabel = {
  fontSize: 10,
  letterSpacing: '0.15em',
  color: '#737373',
};

const metricValue = {
  fontSize: 28,
  fontWeight: 700,
  letterSpacing: '-0.03em',
  marginTop: 8,
};

const panelStyle = {
  borderRadius: 14,
  border: '1px solid rgba(255,255,255,0.06)',
  background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
  padding: 16,
};
