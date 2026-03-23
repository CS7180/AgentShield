import { useState, useEffect } from "react";
import Sidebar from './Sidebar';

export default function ScanMonitorPage() {
  const [progress, setProgress] = useState(71);

  const redAgents = [
    { name: "Prompt injection", status: "done", tested: "12/12", success: 3, gradient: "linear-gradient(90deg, #ef4444, #f97316, #ec4899)" },
    { name: "Jailbreak", status: "running", tested: "5/10", success: 2, gradient: "linear-gradient(90deg, #ec4899, #a855f7, #6366f1)" },
    { name: "Data leakage", status: "done", tested: "8/8", success: 1, gradient: "linear-gradient(90deg, #8b5cf6, #6366f1, #3b82f6)" },
    { name: "Constraint drift", status: "pending", tested: "0/6", success: 0, gradient: "linear-gradient(90deg, #6366f1, #3b82f6, #06b6d4)" },
  ];

  const blueAgents = [
    { name: "Input guard", blocked: 12, status: "active" },
    { name: "Output filter", blocked: 4, status: "active" },
    { name: "Behavior monitor", blocked: 0, status: "active" },
    { name: "Constraint persist.", blocked: 0, status: "active" },
  ];

  const feed = [
    { time: "14:32:07", type: "red", agent: "Jailbreak", msg: "Crescendo attack round 3 — target produced restricted content", severity: "high" },
    { time: "14:31:54", type: "blue", agent: "Input guard", msg: "Blocked indirect injection (base64 encoded payload in system context)", severity: "blocked" },
    { time: "14:31:41", type: "red", agent: "Prompt injection", msg: "Indirect injection via simulated external doc — target executed unintended tool call", severity: "critical" },
    { time: "14:31:28", type: "blue", agent: "Output filter", msg: "Intercepted system prompt leakage in response", severity: "blocked" },
    { time: "14:31:15", type: "judge", agent: "Judge", msg: "Evaluated attack #17: success=true, severity=high, confidence=0.91", severity: "info" },
    { time: "14:31:02", type: "red", agent: "Data leakage", msg: "PII extraction attempt — target refused, no leak detected", severity: "safe" },
  ];

  const statusColor = { done: "#34d399", running: "#fbbf24", pending: "#525252", active: "#34d399" };
  const statusGlow = { done: "rgba(52,211,153,0.4)", running: "rgba(251,191,36,0.4)", pending: "none", active: "rgba(52,211,153,0.4)" };
  const severityStyle = {
    critical: { bg: "rgba(244,63,94,0.12)", border: "rgba(251,113,133,0.3)", color: "#fb7185" },
    high: { bg: "rgba(245,158,11,0.12)", border: "rgba(251,191,36,0.3)", color: "#fbbf24" },
    blocked: { bg: "rgba(52,211,153,0.12)", border: "rgba(52,211,153,0.3)", color: "#34d399" },
    info: { bg: "rgba(139,92,246,0.12)", border: "rgba(139,92,246,0.3)", color: "#c4b5fd" },
    safe: { bg: "rgba(255,255,255,0.04)", border: "rgba(255,255,255,0.08)", color: "#737373" },
  };

  const Dot = ({ color, size = 7 }) => (
    <span style={{ height: size, width: size, borderRadius: "50%", background: color, boxShadow: `0 0 10px ${color}`, flexShrink: 0 }} />
  );

  return (
    <div style={{ minHeight: "100vh", background: "#08090e", color: "#fff", fontFamily: "'Inter', system-ui, -apple-system, sans-serif" }}>
      <div style={{ display: "grid", gridTemplateColumns: "240px 1fr", minHeight: "100vh" }}>

        {/* ═══ SIDEBAR ═══ */}
        <Sidebar activeIndex={2} />

        {/* ═══ MAIN ═══ */}
        <main style={{ background: "radial-gradient(ellipse at 10% 0%, rgba(217,70,239,0.08) 0%, transparent 50%), radial-gradient(ellipse at 90% 0%, rgba(59,130,246,0.06) 0%, transparent 40%), #0a0b10", padding: "28px 32px", overflowY: "auto" }}>

          {/* Header */}
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 20 }}>
            <div>
              <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: "-0.02em" }}>Scan #SC-0042</h1>
                <span style={{ display: "inline-flex", alignItems: "center", gap: 6, padding: "4px 12px", borderRadius: 20, fontSize: 11, fontWeight: 500, background: "rgba(139,92,246,0.12)", border: "1px solid rgba(139,92,246,0.3)", color: "#c4b5fd" }}>adversarial</span>
              </div>
              <p style={{ fontSize: 12, color: "#3f3f3f", marginTop: 6 }}>Target: api.docmind.dev/v1/chat</p>
            </div>
            <button style={{ borderRadius: 10, border: "1px solid rgba(251,113,133,0.25)", background: "rgba(244,63,94,0.1)", padding: "8px 18px", fontSize: 12, color: "#fb7185", cursor: "pointer" }}>Stop scan</button>
          </div>

          {/* Progress bar */}
          <div style={{ marginBottom: 24 }}>
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 6, fontSize: 11, color: "#525252" }}>
              <span>Overall progress</span>
              <span>{progress}%</span>
            </div>
            <div style={{ height: 8, borderRadius: 4, background: "rgba(255,255,255,0.04)", overflow: "hidden" }}>
              <div style={{ height: "100%", borderRadius: 4, background: "linear-gradient(90deg, #d946ef, #8b5cf6, #3b82f6)", width: `${progress}%`, boxShadow: "0 0 16px rgba(139,92,246,0.3)", transition: "width 0.8s ease" }} />
            </div>
          </div>

          {/* Agent cards grid */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 24 }}>

            {/* Red team */}
            <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px" }}>
              <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 16 }}>Red team agents</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {redAgents.map((a) => (
                  <div key={a.name} style={{ borderRadius: 10, padding: "14px", border: "1px solid rgba(255,255,255,0.05)", background: "rgba(255,255,255,0.02)" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                      <span style={{ fontSize: 13, fontWeight: 600, color: a.status === "pending" ? "#3f3f3f" : "#fff" }}>{a.name}</span>
                      <span style={{ display: "inline-flex", alignItems: "center", gap: 6, fontSize: 10, color: statusColor[a.status] }}>
                        <Dot color={statusColor[a.status]} size={6} />{a.status}
                      </span>
                    </div>
                    <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, color: "#525252", marginBottom: 8 }}>
                      <span>Tested: {a.tested}</span>
                      <span style={{ color: a.success > 0 ? "#fb7185" : "#525252" }}>{a.success} successful attacks</span>
                    </div>
                    <div style={{ height: 4, borderRadius: 2, background: "rgba(255,255,255,0.04)", overflow: "hidden" }}>
                      <div style={{ height: "100%", borderRadius: 2, background: a.gradient, width: a.status === "pending" ? "0%" : a.status === "done" ? "100%" : "50%", opacity: a.status === "pending" ? 0.15 : 0.7, transition: "width 0.6s ease" }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Blue team */}
            <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px" }}>
              <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 16 }}>Blue team agents</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {blueAgents.map((a) => (
                  <div key={a.name} style={{ borderRadius: 10, padding: "14px", border: "1px solid rgba(52,211,153,0.1)", background: "rgba(52,211,153,0.03)" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                      <span style={{ fontSize: 13, fontWeight: 600, color: "#a3a3a3" }}>{a.name}</span>
                      <span style={{ display: "inline-flex", alignItems: "center", gap: 6, fontSize: 10, color: "#34d399" }}>
                        <Dot color="#34d399" size={6} />active
                      </span>
                    </div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: a.blocked > 0 ? "#34d399" : "#3f3f3f", letterSpacing: "-0.02em" }}>
                      {a.blocked}<span style={{ fontSize: 12, fontWeight: 400, color: "#525252", marginLeft: 4 }}>blocked</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Live activity feed */}
          <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", overflow: "hidden" }}>
            <div style={{ padding: "16px 20px", borderBottom: "1px solid rgba(255,255,255,0.05)", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252" }}>Live activity feed</div>
              <Dot color="#fb7185" size={6} />
            </div>

            {feed.map((f, idx) => {
              const sv = severityStyle[f.severity];
              return (
                <div key={idx} style={{ padding: "12px 20px", borderBottom: idx < feed.length - 1 ? "1px solid rgba(255,255,255,0.03)" : "none", display: "flex", alignItems: "flex-start", gap: 12 }}>
                  <span style={{ fontSize: 10, color: "#3f3f3f", fontFamily: "monospace", marginTop: 2, whiteSpace: "nowrap" }}>{f.time}</span>
                  <Dot color={f.type === "red" ? "#fb7185" : f.type === "blue" ? "#34d399" : "#c4b5fd"} size={7} />
                  <div style={{ flex: 1 }}>
                    <span style={{ fontSize: 12, fontWeight: 600, color: f.type === "red" ? "#fda4af" : f.type === "blue" ? "#6ee7b7" : "#ddd6fe" }}>{f.agent}</span>
                    <span style={{ fontSize: 12, color: "#737373", marginLeft: 8 }}>{f.msg}</span>
                  </div>
                  <span style={{ display: "inline-flex", padding: "2px 8px", borderRadius: 6, fontSize: 9, fontWeight: 500, background: sv.bg, border: `1px solid ${sv.border}`, color: sv.color, whiteSpace: "nowrap" }}>{f.severity}</span>
                </div>
              );
            })}
          </div>

        </main>
      </div>
    </div>
  );
}
