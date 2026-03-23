import React from 'react';

const DashboardContent = () => {
  const stats = [
    { label: "TOTAL SCANS", value: "24", dot: "#c084fc" },
    { label: "CRITICAL", value: "7", dot: "#fb7185" },
    { label: "AVG DEFENSE", value: "89.2%", dot: "#34d399" },
    { label: "JUDGE τ", value: "0.81", dot: "#c084fc" },
  ];

  const scans = [
    { target: "DocMind API", endpoint: "api.docmind.dev/v1/chat", mode: "adversarial", modeTone: "purple", status: "done", score: "72.00", vulns: [{ l: "1C", t: "c" }, { l: "2H", t: "h" }] },
    { target: "Chatbot v2.1", endpoint: "staging.chatbot.io/api", mode: "red team", modeTone: "red", status: "done", score: "45.00", vulns: [{ l: "3C", t: "c" }, { l: "4H", t: "h" }] },
    { target: "Support Agent", endpoint: "support.acme.com/agent", mode: "blue team", modeTone: "blue", status: "running", score: "—", vulns: [{ l: "—", t: "n" }] },
  ];

  const coverage = [
    { label: "#1 Injection", value: 85, gradient: "linear-gradient(90deg, #ef4444, #f97316, #ec4899)", glow: "rgba(239,68,68,0.25)" },
    { label: "#2 Data leak", value: 70, gradient: "linear-gradient(90deg, #ec4899, #a855f7, #6366f1)", glow: "rgba(168,85,247,0.25)" },
    { label: "#3 Jailbreak", value: 90, gradient: "linear-gradient(90deg, #8b5cf6, #6366f1, #3b82f6)", glow: "rgba(99,102,241,0.25)" },
    { label: "#4 Constraint", value: 60, gradient: "linear-gradient(90deg, #6366f1, #3b82f6, #06b6d4)", glow: "rgba(59,130,246,0.25)" },
  ];

  const modeColor = { purple: ["rgba(139,92,246,0.15)", "rgba(139,92,246,0.4)", "#c4b5fd"], red: ["rgba(244,63,94,0.15)", "rgba(244,63,94,0.4)", "#fda4af"], blue: ["rgba(14,165,233,0.15)", "rgba(14,165,233,0.4)", "#7dd3fc"] };
  const vulnColor = { c: ["rgba(244,63,94,0.12)", "rgba(251,113,133,0.35)", "#fb7185"], h: ["rgba(245,158,11,0.12)", "rgba(251,191,36,0.35)", "#fbbf24"], n: ["transparent", "rgba(255,255,255,0.08)", "#525252"] };

  const Pill = ({ bg, border, color, children }) => (
    <span style={{ display: "inline-flex", alignItems: "center", gap: "6px", padding: "4px 12px", borderRadius: "20px", fontSize: "11px", fontWeight: 500, background: bg, border: `1px solid ${border}`, color }}>{children}</span>
  );

  const Dot = ({ color, size = 7, shadow = true }) => (
    <span style={{ height: size, width: size, borderRadius: "50%", background: color, boxShadow: shadow ? `0 0 10px ${color}` : "none", flexShrink: 0 }} />
  );

  return (
    <>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <div>
          <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: "-0.02em" }}>Dashboard</h1>
          <p style={{ fontSize: 12, color: "#3f3f3f", marginTop: 4 }}>Continuous red-team assessment across targets</p>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <button style={{ borderRadius: 10, border: "1px solid rgba(255,255,255,0.08)", background: "rgba(255,255,255,0.03)", padding: "8px 14px", fontSize: 12, color: "#737373", cursor: "pointer" }}>Filters</button>
          <button style={{
            borderRadius: 10,
            border: "1px solid rgba(217,70,239,0.25)",
            background: "linear-gradient(135deg, rgba(217,70,239,0.15), rgba(139,92,246,0.08))",
            padding: "8px 18px", fontSize: 12, color: "#e9d5ff", cursor: "pointer",
            boxShadow: "0 0 20px rgba(217,70,239,0.12)",
          }}>+ New scan</button>
        </div>
      </div>

      {/* ── Stat cards ── */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 14, marginBottom: 24 }}>
        {stats.map((s) => (
          <div key={s.label} style={{
            borderRadius: 16, padding: "20px",
            border: "1px solid rgba(255,255,255,0.06)",
            background: "linear-gradient(135deg, rgba(255,255,255,0.035), rgba(255,255,255,0.015))",
            backdropFilter: "blur(8px)",
            position: "relative", overflow: "hidden",
          }}>
            <div style={{ position: "absolute", top: 16, right: 16 }}><Dot color={s.dot} size={8} /></div>
            <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252" }}>{s.label}</div>
            <div style={{ marginTop: 12, fontSize: 34, fontWeight: 700, letterSpacing: "-0.03em", color: s.dot === "#fb7185" ? "#fb7185" : s.dot === "#34d399" ? "#34d399" : "#fff" }}>{s.value}</div>
          </div>
        ))}
      </div>

      {/* ── Recent scans ── */}
      <div style={{
        borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)",
        background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
        overflow: "hidden", marginBottom: 24,
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "16px 20px", borderBottom: "1px solid rgba(255,255,255,0.05)" }}>
          <div>
            <div style={{ fontSize: 14, fontWeight: 600 }}>Recent scans</div>
            <div style={{ fontSize: 11, color: "#3f3f3f", marginTop: 2 }}>Continuous red-team assessment across targets</div>
          </div>
          <div style={{ fontSize: 11, color: "#3f3f3f" }}>Last updated 2 min ago</div>
        </div>

        {scans.map((scan, idx) => (
          <div key={scan.target} style={{
            padding: "18px 20px",
            borderBottom: idx < scans.length - 1 ? "1px solid rgba(255,255,255,0.04)" : "none",
            display: "flex", justifyContent: "space-between", alignItems: "center",
          }}>
            <div style={{ flex: "1 1 0" }}>
              <div style={{ fontSize: 15, fontWeight: 600 }}>{scan.target}</div>
              <div style={{ fontSize: 11, color: "#3f3f3f", marginTop: 3 }}>{scan.endpoint}</div>
              <div style={{ display: "flex", gap: 8, marginTop: 10, flexWrap: "wrap" }}>
                <Pill bg={modeColor[scan.modeTone][0]} border={modeColor[scan.modeTone][1]} color={modeColor[scan.modeTone][2]}>{scan.mode}</Pill>
                <Pill bg={scan.status === "done" ? "rgba(52,211,153,0.12)" : "rgba(251,191,36,0.12)"} border={scan.status === "done" ? "rgba(52,211,153,0.35)" : "rgba(251,191,36,0.35)"} color={scan.status === "done" ? "#34d399" : "#fbbf24"}>
                  <Dot color={scan.status === "done" ? "#34d399" : "#fbbf24"} size={6} />{scan.status}
                </Pill>
                {scan.vulns.map((v) => (
                  <Pill key={v.l} bg={vulnColor[v.t][0]} border={vulnColor[v.t][1]} color={vulnColor[v.t][2]}>{v.l}</Pill>
                ))}
              </div>
            </div>
            <div style={{ textAlign: "right", flexShrink: 0, marginLeft: 20 }}>
              <div style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.03em" }}>{scan.score}</div>
              <div style={{ fontSize: 10, color: "#3f3f3f", marginTop: 2 }}>judge score</div>
            </div>
          </div>
        ))}
      </div>

      {/* ── OWASP coverage ── */}
      <div style={{
        borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)",
        background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
        padding: "20px",
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 20 }}>
          <div>
            <div style={{ fontSize: 14, fontWeight: 600 }}>OWASP LLM Top 10</div>
            <div style={{ fontSize: 11, color: "#3f3f3f", marginTop: 2 }}>Detection coverage by attack class</div>
          </div>
          <div style={{ fontSize: 11, color: "#525252" }}>Coverage</div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          {coverage.map((item) => (
            <div key={item.label}>
              <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 8, fontSize: 13 }}>
                <span style={{ color: "#a3a3a3", fontWeight: 500 }}>{item.label}</span>
                <span style={{ color: "#737373", fontWeight: 500 }}>{item.value}%</span>
              </div>
              <div style={{ height: 10, borderRadius: 5, background: "rgba(255,255,255,0.04)", overflow: "hidden" }}>
                <div style={{
                  height: "100%", borderRadius: 5,
                  background: item.gradient,
                  width: `${item.value}%`,
                  boxShadow: `0 0 16px ${item.glow}, 0 2px 8px ${item.glow}`,
                  transition: "width 0.8s ease",
                }} />
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
};

export default DashboardContent;