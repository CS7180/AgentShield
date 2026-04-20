import React from 'react';
export default function LandingPage({ onLogIn, onGetStarted, onViewDemo }) {
  const features = [
    {
      icon: "R",
      title: "Red team agents",
      desc: "4 parallel attack agents probing prompt injection, jailbreak, data leakage, and constraint drift at the same time.",
      gradient: "linear-gradient(135deg, #ef4444, #ec4899)",
      glow: "rgba(239,68,68,0.15)",
    },
    {
      icon: "B",
      title: "Blue team defense",
      desc: "4 defense agents intercept attacks in real time through input guard, output filter, behavior monitor, and constraint persistence.",
      gradient: "linear-gradient(135deg, #34d399, #06b6d4)",
      glow: "rgba(52,211,153,0.15)",
    },
    {
      icon: "J",
      title: "LLM-as-Judge",
      desc: "Gemini 2.5 Pro evaluates every attack for success, severity, OWASP category, and confidence against calibrated golden sets.",
      gradient: "linear-gradient(135deg, #a855f7, #6366f1)",
      glow: "rgba(139,92,246,0.15)",
    },
    {
      icon: "O",
      title: "OWASP reports",
      desc: "Auto-generated security reports mapped to the OWASP LLM Top 10, including red-only versus adversarial comparisons.",
      gradient: "linear-gradient(135deg, #f97316, #fbbf24)",
      glow: "rgba(249,115,22,0.15)",
    },
    {
      icon: "C",
      title: "CI/CD integration",
      desc: "Plug into Jenkins as a quality gate so every model update is security-scanned before deployment.",
      gradient: "linear-gradient(135deg, #3b82f6, #06b6d4)",
      glow: "rgba(59,130,246,0.15)",
    },
    {
      icon: "M",
      title: "Real-time monitoring",
      desc: "Watch agents attack and defend live with Prometheus metrics, Grafana dashboards, and alerts on anomalies.",
      gradient: "linear-gradient(135deg, #ec4899, #a855f7)",
      glow: "rgba(236,72,153,0.15)",
    },
  ];

  const incidents = [
    {
      who: "Meta alignment director",
      what: "200+ emails deleted by a rogue OpenClaw agent that ignored safety constraints after context window compaction.",
      when: "Feb 2026",
    },
    {
      who: "Cline GitHub triage bot",
      what: "Hijacked via prompt injection in an issue title, leading to code execution and release pipeline compromise.",
      when: "Mar 2026",
    },
    {
      who: "ServiceNow AI assistant",
      what: "Second-order prompt injection let a low-privilege agent trick a high-privilege agent into exporting case files.",
      when: "Late 2025",
    },
  ];

  const stats = [
    { value: "45%", label: "of AI-generated code fails security tests", src: "Veracode 2025" },
    { value: "2.74x", label: "more security vulns in AI co-authored PRs", src: "CodeRabbit 2025" },
    { value: "#1", label: "Prompt injection in OWASP LLM Top 10 for two years", src: "OWASP 2025" },
  ];

  const modes = [
    {
      mode: "Red team",
      desc: "Attack only. Discover how many vulnerabilities your bare system has and establish the baseline.",
      color: "#fb7185",
      bg: "rgba(244,63,94,0.06)",
      border: "rgba(251,113,133,0.15)",
      report: "Report A",
    },
    {
      mode: "Blue team",
      desc: "Defense only. Measure false positive rate on normal traffic and verify no user impact.",
      color: "#34d399",
      bg: "rgba(52,211,153,0.06)",
      border: "rgba(52,211,153,0.15)",
      report: "Report B",
    },
    {
      mode: "Adversarial",
      desc: "Red and blue together. Attack while defending to see which vulnerabilities escape and which get caught.",
      color: "#c4b5fd",
      bg: "rgba(139,92,246,0.06)",
      border: "rgba(139,92,246,0.15)",
      report: "Report C",
    },
  ];

  return (
    <div
      style={{
        minHeight: "100vh",
        background:
          "radial-gradient(circle at top, rgba(139,92,246,0.16), transparent 35%), radial-gradient(circle at 20% 20%, rgba(244,63,94,0.12), transparent 25%), #06070b",
        color: "#fff",
        fontFamily: "'Inter', system-ui, -apple-system, BlinkMacSystemFont, sans-serif",
      }}
    >
      <nav
        style={{
          position: "sticky",
          top: 0,
          zIndex: 50,
          padding: "16px 32px",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 16,
          borderBottom: "1px solid rgba(255,255,255,0.06)",
          background: "rgba(6,7,11,0.85)",
          backdropFilter: "blur(12px)",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <div
            style={{
              height: 36,
              width: 36,
              borderRadius: 12,
              border: "1px solid rgba(217,70,239,0.25)",
              background: "linear-gradient(135deg, rgba(217,70,239,0.15), rgba(139,92,246,0.1))",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              boxShadow: "0 0 20px rgba(217,70,239,0.12)",
            }}
          >
            <div
              style={{
                height: 12,
                width: 12,
                borderRadius: 3,
                background: "linear-gradient(135deg, #d946ef, #8b5cf6)",
              }}
            />
          </div>
          <span style={{ fontWeight: 700, fontSize: 16, letterSpacing: "-0.01em" }}>AgentShield</span>
        </div>
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", justifyContent: "flex-end" }}>
          <button
            type="button"
            onClick={onLogIn}
            style={{
              borderRadius: 10,
              border: "1px solid rgba(255,255,255,0.1)",
              background: "rgba(255,255,255,0.04)",
              padding: "8px 20px",
              fontSize: 13,
              color: "#a3a3a3",
              cursor: "pointer",
            }}
          >
            Log in
          </button>
          <button
            type="button"
            onClick={onGetStarted}
            style={{
              borderRadius: 10,
              border: "1px solid rgba(217,70,239,0.3)",
              background: "linear-gradient(135deg, rgba(217,70,239,0.2), rgba(139,92,246,0.12))",
              padding: "8px 20px",
              fontSize: 13,
              color: "#e9d5ff",
              cursor: "pointer",
              boxShadow: "0 0 20px rgba(217,70,239,0.15)",
            }}
          >
            Get started
          </button>
        </div>
      </nav>

      <section
        style={{
          padding: "100px 32px 80px",
          textAlign: "center",
          background:
            "radial-gradient(ellipse at 50% 0%, rgba(217,70,239,0.12) 0%, transparent 50%), radial-gradient(ellipse at 30% 20%, rgba(59,130,246,0.08) 0%, transparent 40%)",
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            inset: 0,
            opacity: 0.03,
            backgroundImage:
              "linear-gradient(rgba(255,255,255,0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.5) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />

        <div style={{ position: "relative", maxWidth: 720, margin: "0 auto" }}>
          <div
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 8,
              padding: "6px 16px",
              borderRadius: 20,
              border: "1px solid rgba(251,113,133,0.25)",
              background: "rgba(244,63,94,0.08)",
              fontSize: 12,
              color: "#fda4af",
              marginBottom: 28,
            }}
          >
            <span
              style={{
                height: 6,
                width: 6,
                borderRadius: "50%",
                background: "#fb7185",
                boxShadow: "0 0 8px rgba(251,113,133,0.5)",
              }}
            />
            AI agents are under attack. Is yours protected?
          </div>

          <h1
            style={{
              fontSize: "clamp(40px, 7vw, 52px)",
              fontWeight: 800,
              letterSpacing: "-0.03em",
              lineHeight: 1.1,
              margin: "0 0 20px",
              background: "linear-gradient(135deg, #fff 0%, #c084fc 50%, #818cf8 100%)",
              WebkitBackgroundClip: "text",
              WebkitTextFillColor: "transparent",
            }}
          >
            Red-blue teaming
            <br />
            for LLM applications
          </h1>

          <p
            style={{
              fontSize: 18,
              color: "#b0b3be",
              lineHeight: 1.6,
              maxWidth: 560,
              margin: "0 auto 36px",
            }}
          >
            Parallel attack agents probe your system. Defense agents protect it in real time. LLM-as-Judge scores both sides.
          </p>

          <div style={{ display: "flex", justifyContent: "center", gap: 14, flexWrap: "wrap" }}>
            <button
              type="button"
              onClick={onGetStarted}
              style={{
                borderRadius: 12,
                padding: "14px 32px",
                fontSize: 15,
                fontWeight: 600,
                cursor: "pointer",
                border: "1px solid rgba(217,70,239,0.35)",
                background: "linear-gradient(135deg, rgba(217,70,239,0.25), rgba(139,92,246,0.15))",
                color: "#e9d5ff",
                boxShadow: "0 0 32px rgba(217,70,239,0.2), 0 8px 24px rgba(0,0,0,0.3)",
              }}
            >
              Start scanning
            </button>
            <button
              type="button"
              onClick={onViewDemo}
              style={{
                borderRadius: 12,
                padding: "14px 32px",
                fontSize: 15,
                fontWeight: 500,
                cursor: "pointer",
                border: "1px solid rgba(255,255,255,0.1)",
                background: "rgba(255,255,255,0.04)",
                color: "#a3a3a3",
              }}
            >
              View demo
            </button>
          </div>
        </div>
      </section>

      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>
            This already happened
          </h2>
          <p style={{ fontSize: 14, color: "#707380" }}>
            Real incidents. Real damage. Automated security testing is not optional anymore.
          </p>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {incidents.map((incident) => (
            <div
              key={incident.who}
              style={{
                borderRadius: 14,
                padding: "20px 24px",
                border: "1px solid rgba(251,113,133,0.1)",
                background: "linear-gradient(135deg, rgba(244,63,94,0.04), rgba(255,255,255,0.01))",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "flex-start",
                gap: 20,
                flexWrap: "wrap",
              }}
            >
              <div style={{ flex: 1, minWidth: 260 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: "#fda4af", marginBottom: 4 }}>{incident.who}</div>
                <div style={{ fontSize: 13, color: "#b0b3be", lineHeight: 1.5 }}>{incident.what}</div>
              </div>
              <span style={{ fontSize: 11, color: "#6a6e7d", whiteSpace: "nowrap", marginTop: 2 }}>{incident.when}</span>
            </div>
          ))}
        </div>
      </section>

      <section style={{ padding: "40px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
            gap: 16,
          }}
        >
          {stats.map((stat) => (
            <div
              key={stat.value}
              style={{
                borderRadius: 14,
                padding: 24,
                border: "1px solid rgba(255,255,255,0.06)",
                background: "linear-gradient(135deg, rgba(255,255,255,0.035), rgba(255,255,255,0.015))",
                textAlign: "center",
              }}
            >
              <div
                style={{
                  fontSize: 32,
                  fontWeight: 700,
                  letterSpacing: "-0.03em",
                  background: "linear-gradient(135deg, #fb7185, #fbbf24)",
                  WebkitBackgroundClip: "text",
                  WebkitTextFillColor: "transparent",
                }}
              >
                {stat.value}
              </div>
              <div style={{ fontSize: 13, color: "#b0b3be", marginTop: 8, lineHeight: 1.4 }}>{stat.label}</div>
              <div style={{ fontSize: 10, color: "#6a6e7d", marginTop: 6 }}>{stat.src}</div>
            </div>
          ))}
        </div>
      </section>

      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>
            How AgentShield works
          </h2>
          <p style={{ fontSize: 14, color: "#707380" }}>Point it at any LLM API. Get a security report in minutes.</p>
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
            gap: 16,
          }}
        >
          {features.map((feature) => (
            <div
              key={feature.title}
              style={{
                borderRadius: 16,
                padding: 24,
                border: "1px solid rgba(255,255,255,0.06)",
                background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
                position: "relative",
                overflow: "hidden",
              }}
            >
              <div
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  right: 0,
                  height: 2,
                  background: feature.gradient,
                  opacity: 0.6,
                }}
              />
              <div
                style={{
                  height: 40,
                  width: 40,
                  borderRadius: 12,
                  marginBottom: 16,
                  background: feature.glow,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: 18,
                  fontWeight: 700,
                }}
              >
                {feature.icon}
              </div>
              <div style={{ fontSize: 15, fontWeight: 600, marginBottom: 8 }}>{feature.title}</div>
              <div style={{ fontSize: 13, color: "#b0b3be", lineHeight: 1.5 }}>{feature.desc}</div>
            </div>
          ))}
        </div>
      </section>

      <section style={{ padding: "60px 32px", maxWidth: 960, margin: "0 auto" }}>
        <div style={{ textAlign: "center", marginBottom: 40 }}>
          <h2 style={{ fontSize: 28, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 8px" }}>
            Three testing modes
          </h2>
          <p style={{ fontSize: 14, color: "#707380" }}>Baseline, defend, validate. A complete security feedback loop.</p>
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
            gap: 16,
          }}
        >
          {modes.map((mode) => (
            <div
              key={mode.mode}
              style={{
                borderRadius: 16,
                padding: "28px 24px",
                border: `1px solid ${mode.border}`,
                background: mode.bg,
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
                <span
                  style={{
                    height: 10,
                    width: 10,
                    borderRadius: "50%",
                    background: mode.color,
                    boxShadow: `0 0 12px ${mode.color}`,
                  }}
                />
                <span style={{ fontSize: 16, fontWeight: 700, color: mode.color }}>{mode.mode}</span>
              </div>
              <div style={{ fontSize: 13, color: "#b0b3be", lineHeight: 1.6, marginBottom: 16 }}>{mode.desc}</div>
              <div
                style={{
                  display: "inline-flex",
                  padding: "4px 12px",
                  borderRadius: 6,
                  background: "rgba(255,255,255,0.04)",
                  border: "1px solid rgba(255,255,255,0.08)",
                  fontSize: 11,
                  color: "#7f8497",
                }}
              >
                Generates {mode.report}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section
        style={{
          padding: "80px 32px",
          textAlign: "center",
          background: "radial-gradient(ellipse at 50% 100%, rgba(217,70,239,0.08) 0%, transparent 50%)",
        }}
      >
        <h2 style={{ fontSize: 32, fontWeight: 700, letterSpacing: "-0.02em", margin: "0 0 12px" }}>
          Do not ship another vulnerable agent
        </h2>
        <p style={{ fontSize: 15, color: "#8e94a7", marginBottom: 32 }}>
          Point AgentShield at your API. Get your OWASP report in minutes.
        </p>
        <button
          type="button"
          onClick={onGetStarted}
          style={{
            borderRadius: 14,
            padding: "16px 40px",
            fontSize: 16,
            fontWeight: 600,
            cursor: "pointer",
            border: "1px solid rgba(217,70,239,0.35)",
            background: "linear-gradient(135deg, rgba(217,70,239,0.25), rgba(139,92,246,0.15))",
            color: "#e9d5ff",
            boxShadow: "0 0 40px rgba(217,70,239,0.2), 0 8px 32px rgba(0,0,0,0.3)",
          }}
        >
          Get started
        </button>
      </section>

      <footer
        style={{
          padding: "24px 32px",
          borderTop: "1px solid rgba(255,255,255,0.06)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 16,
          flexWrap: "wrap",
          fontSize: 11,
          color: "#6a6e7d",
        }}
      >
        <span>AgentShield | Multi-agent AI red-blue teaming platform</span>
        <span>Frontend prototype entry page</span>
      </footer>
    </div>
  );
}
