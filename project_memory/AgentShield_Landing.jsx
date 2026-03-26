import { useState } from "react";

export default function AgentShieldLanding() {
  const features = [
    {
      icon: "⚔",
      title: "Red team agents",
      desc: "4 parallel attack agents — prompt injection, jailbreak, data leakage, constraint drift — probing your LLM from every angle simultaneously.",
      gradient: "linear-gradient(135deg, #ef4444, #ec4899)",
      glow: "rgba(239,68,68,0.15)",
    },
    {
      icon: "🛡",
      title: "Blue team defense",
      desc: "4 defense agents intercept attacks in real time — input guard, output filter, behavior monitor, constraint persistence.",
      gradient: "linear-gradient(135deg, #34d399, #06b6d4)",
      glow: "rgba(52,211,153,0.15)",
    },
    {
      icon: "⚖",
      title: "LLM-as-Judge",
      desc: "Gemini 2.5 Pro evaluates every attack — success, severity, OWASP category, confidence — calibrated against human-labeled golden sets.",
      gradient: "linear-gradient(135deg, #a855f7, #6366f1)",
      glow: "rgba(139,92,246,0.15)",
    },
    {
      icon: "📊",
      title: "OWASP reports",
      desc: "Auto-generated security reports mapped to OWASP LLM Top 10. Compare red-only vs adversarial results side by side.",
      gradient: "linear-gradient(135deg, #f97316, #fbbf24)",
      glow: "rgba(249,115,22,0.15)",
    },
    {
      icon: "🔁",
      title: "CI/CD integration",
      desc: "Plug into Jenkins as a quality gate. Every model update gets security-scanned before deployment. Critical findings block the pipeline.",
      gradient: "linear-gradient(135deg, #3b82f6, #06b6d4)",
      glow: "rgba(59,130,246,0.15)",
    },
    {
      icon: "📡",
      title: "Real-time monitoring",
      desc: "Watch agents attack and defend live. Prometheus metrics, Grafana dashboards, alerting on anomalies and calibration drift.",
      gradient: "linear-gradient(135deg, #ec4899, #a855f7)",
      glow: "rgba(236,72,153,0.15)",
    },
  ];

  const incidents = [
    { who: "Meta alignment director", what: "200+ emails deleted by rogue OpenClaw agent that ignored safety constraints after context window compaction", when: "Feb 2026" },
    { who: "Cline GitHub triage bot", what: "Hijacked via prompt injection in an issue title — attacker gained code execution and compromised the release pipeline", when: "Mar 2026" },
    { who: "ServiceNow AI assistant", what: "Second-order prompt injection let a low-privilege agent trick a high-privilege agent into exporting case files", when: "Late 2025" },
  ];

  const stats = [
    { value: "45%", label: "of AI-generated code fails security tests", src: "Veracode 2025" },
    { value: "2.74×", label: "more security vulns in AI co-authored PRs", src: "CodeRabbit 2025" },
    { value: "#1", label: "Prompt injection — OWASP LLM Top 10 two years running", src: "OWASP 2025" },
  ];

  return (
    <div style={{ minHeight: "100vh", background: "#06070b", color: "#fff", fontFamily: "'Inter', system-ui, -apple-system, sans-serif" }}>

      {/* ═══ NAV ═══ */}
      <nav style={{
        position: "sticky", top: 0, zIndex: 50,
        padding: "16px 32px",
        display: "flex", justifyContent: "space-between", alignItems: "center",
        borderBottom: "1px solid rgba(255,255,255,0.06)",
        background: "rgba(6,7,11,0.85)",
        backdropFilter: "blur(12px)",
      }}>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <div style={{
            height: 36, width: 36, borderRadius: 12,
            border: "1px solid rgba(217,70,239,0.25)",
            background: "linear-gradient(135deg, rgba(217,70,239,0.15), rgba(139,92,246,0.1))",
            display: "flex", alignItems: "center", justifyContent: "center",
            boxShadow: "0 0 20px rgba(217,70,239,0.12)",
          }}>
            <div style={{ height: 12, width: 12, borderRadius: 3, background: "linear-gradient(135deg, #d946ef, #8b5cf6)" }} />
          </div>
          <span style={{ fontWeight: 700, fontSize: 16, letterSpacing: "-0.01em" }}>AgentShield</span>
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <button style={{
            borderRadius: 10, border: "1px solid rgba(255,255,255,0.1)",
            background: "rgba(255,255,255,0.04)", padding: "8px 20px",
            fontSize: 13, color: "#a3a3a3", cursor: "pointer",
          }}>Log in</button>
          <button style={{
            borderRadius: 10, border: "1px solid rgba(217,70,239,0.3)",
            background: "linear-gradient(135deg, rgba(217,70,239,0.2), rgba(139,92,246,0.12))",
            padding: "8px 20px", fontSize: 13, color: "#e9d5ff", cursor: "pointer",
            boxShadow: "0 0 20px rgba(217,70,239,0.15)",
          }}>Get started</button>
        </div>
      </nav>

      {/* ═══ HERO ═══ */}
      <section style={{
        padding: "100px 32px 80px",
        textAlign: "center",
        background: "radial-gradient(ellipse at 50% 0%, rgba(217,70,239,0.12) 0%, transparent 50%), radial-gradient(ellipse at 30% 20%, rgba(59,130,246,0.08) 0%, transparent 40%)",
        position: "relative",
      }}>
        {/* Subtle grid overlay */}
        <div style={{
          position: "absolute", inset: 0, opacity: 0.03,
          backgroundImage: "linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }} />

        <div style={{ position: "relative", maxWidth: 720, margin: "0 auto" }}>
          {/* Badge */}
          <div style={{
            display: "inline-flex", alignItems: "center", gap: 8,
            padding: "6px 16px", borderRadius: 20,
            border: "1px solid rgba(251,113,133,0.25)",
            background: "rgba(244,63,94,0.08)",
            fontSize: 12, color: "#fda4af", marginBottom: 28,
          }}>
            <span style={{ height: 6, width: 6, borderRadius: "50%", background: "#fb7185", boxShadow: "0 0 8px rgba(251,113,133,0.5)" }} />
            AI agents are under attack — is yours protected?
          </div>

          <h1 style={{
            fontSize: 52, fontWeight: 800, letterSpacing: "-0.03em",
            lineHeight: 1.1, margin: "0 0 20px",
            background: "linear-gradient(135deg, #fff 0%, #c084fc 50%, #818cf8 100%)",
            WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent",
          }}>
            Red-blue teaming<br />for LLM applications
          </h1>

          <p style={{ fontSize: 18, color: "#737373", lineHeight: 1.6, maxWidth: 560, margin: "0 auto 36px" }}>
            Parallel attack agents probe your system. Defense agents protect it in real time. LLM-as-Judge scores both sides. Ship with confidence.
          </p>

          <div style={{ display: "flex", justifyContent: "center", gap: 14 }}>
            <button style={{
              borderRadius: 12, padding: "14px 32px", fontSize: 15, fontWeight: 600, cursor: "pointer",
              border: "1px solid rgba(217,70,239,0.35)",
              background: "linear-gradient(135deg, rgba(217,70,239,0.25), rgba(139,92,246,0.15))",
              color: "#e9d5ff",
              boxShadow: "0 0 32px rgba(217,70,239,0.2), 0 8px 24px rgba(0,0,0,0.3)",
            }}>Start scanning — free</button>
            <button style={{
              borderRadius: 12, padding: "14px 32px", fontSize: 15, fontWeight: 500, cursor: "pointer",
              border: "1px solid rgba(255,255,255,0.1)",
              background: "rgba(255,255,255,0.04)",
              color: "#a3a3a3",
            }}>View demo</button>
          </div>
        </div>
      </section>

      {/* ═══ INCIDENTS ═══ */}
      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>This already happened</h2>
          <p style={{ fontSize: 14, color: "#525252" }}>Real incidents. Real damage. All preventable with automated security testing.</p>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {incidents.map((inc, idx) => (
            <div key={idx} style={{
              borderRadius: 14, padding: "20px 24px",
              border: "1px solid rgba(251,113,133,0.1)",
              background: "linear-gradient(135deg, rgba(244,63,94,0.04), rgba(255,255,255,0.01))",
              display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 20,
            }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "#fda4af", marginBottom: 4 }}>{inc.who}</div>
                <div style={{ fontSize: 13, color: "#737373", lineHeight: 1.5 }}>{inc.what}</div>
              </div>
              <span style={{ fontSize: 11, color: "#3f3f3f", whiteSpace: "nowrap", marginTop: 2 }}>{inc.when}</span>
            </div>
          ))}
        </div>
      </section>

      {/* ═══ STATS BAR ═══ */}
      <section style={{ padding: "40px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16 }}>
          {stats.map((s) => (
            <div key={s.value} style={{
              borderRadius: 14, padding: "24px",
              border: "1px solid rgba(255,255,255,0.06)",
              background: "linear-gradient(135deg, rgba(255,255,255,0.035), rgba(255,255,255,0.015))",
              textAlign: "center",
            }}>
              <div style={{ fontSize: 32, fontWeight: 700, letterSpacing: "-0.03em", background: "linear-gradient(135deg, #fb7185, #fbbf24)", WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>{s.value}</div>
              <div style={{ fontSize: 13, color: "#737373", marginTop: 8, lineHeight: 1.4 }}>{s.label}</div>
              <div style={{ fontSize: 10, color: "#3f3f3f", marginTop: 6 }}>{s.src}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ═══ FEATURES ═══ */}
      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>How AgentShield works</h2>
          <p style={{ fontSize: 14, color: "#525252" }}>Point it at any LLM API. Get a security report in minutes.</p>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16 }}>
          {features.map((f) => (
            <div key={f.title} style={{
              borderRadius: 16, padding: "24px",
              border: "1px solid rgba(255,255,255,0.06)",
              background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
              position: "relative", overflow: "hidden",
            }}>
              {/* Top glow accent */}
              <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: 2, background: f.gradient, opacity: 0.6 }} />

              <div style={{
                height: 40, width: 40, borderRadius: 12, marginBottom: 16,
                background: f.glow, display: "flex", alignItems: "center", justifyContent: "center",
                fontSize: 20,
              }}>{f.icon}</div>
              <div style={{ fontSize: 15, fontWeight: 600, marginBottom: 8 }}>{f.title}</div>
              <div style={{ fontSize: 13, color: "#737373", lineHeight: 1.5 }}>{f.desc}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ═══ THREE MODES ═══ */}
      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>Three testing modes</h2>
          <p style={{ fontSize: 14, color: "#525252" }}>Baseline → defend → validate. A complete security feedback loop.</p>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16 }}>
          {[
            { mode: "Red team", desc: "Attack only. Discover how many vulnerabilities your bare system has. This is your security baseline.", color: "#fb7185", bg: "rgba(244,63,94,0.06)", border: "rgba(251,113,133,0.15)", report: "Report A" },
            { mode: "Blue team", desc: "Defense only. Deploy protection agents and measure false positive rate on normal traffic. Ensure no user impact.", color: "#34d399", bg: "rgba(52,211,153,0.06)", border: "rgba(52,211,153,0.15)", report: "Report B" },
            { mode: "Adversarial", desc: "Red + blue simultaneous. Attack while defending. See which vulnerabilities escape and which get caught.", color: "#c4b5fd", bg: "rgba(139,92,246,0.06)", border: "rgba(139,92,246,0.15)", report: "Report C" },
          ].map((m) => (
            <div key={m.mode} style={{
              borderRadius: 16, padding: "28px 24px",
              border: `1px solid ${m.border}`,
              background: m.bg,
            }}>
              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
                <span style={{ height: 10, width: 10, borderRadius: "50%", background: m.color, boxShadow: `0 0 12px ${m.color}` }} />
                <span style={{ fontSize: 16, fontWeight: 700, color: m.color }}>{m.mode}</span>
              </div>
              <div style={{ fontSize: 13, color: "#737373", lineHeight: 1.6, marginBottom: 16 }}>{m.desc}</div>
              <div style={{
                display: "inline-flex", padding: "4px 12px", borderRadius: 6,
                background: "rgba(255,255,255,0.04)", border: "1px solid rgba(255,255,255,0.08)",
                fontSize: 11, color: "#525252",
              }}>Generates → {m.report}</div>
            </div>
          ))}
        </div>
      </section>

      {/* ═══ CTA ═══ */}
      <section style={{
        padding: "80px 32px", textAlign: "center",
        background: "radial-gradient(ellipse at 50% 100%, rgba(217,70,239,0.08) 0%, transparent 50%)",
      }}>
        <h2 style={{ fontSize: 32, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 12px" }}>
          Don't ship another vulnerable agent
        </h2>
        <p style={{ fontSize: 15, color: "#525252", marginBottom: 32 }}>
          Point AgentShield at your API. Get your OWASP report in 3 minutes.
        </p>
        <button style={{
          borderRadius: 14, padding: "16px 40px", fontSize: 16, fontWeight: 600, cursor: "pointer",
          border: "1px solid rgba(217,70,239,0.35)",
          background: "linear-gradient(135deg, rgba(217,70,239,0.25), rgba(139,92,246,0.15))",
          color: "#e9d5ff",
          boxShadow: "0 0 40px rgba(217,70,239,0.2), 0 8px 32px rgba(0,0,0,0.3)",
        }}>Get started — free</button>
      </section>

      {/* ═══ FOOTER ═══ */}
      <footer style={{
        padding: "24px 32px", borderTop: "1px solid rgba(255,255,255,0.06)",
        display: "flex", justifyContent: "space-between", alignItems: "center",
        fontSize: 11, color: "#3f3f3f",
      }}>
        <span>AgentShield · Multi-agent AI red-blue teaming platform</span>
      </footer>
    </div>
  );
}
