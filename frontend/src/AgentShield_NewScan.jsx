import { useState } from "react";

export default function NewScanPage() {
  const [mode, setMode] = useState("red_team");
  const [attacks, setAttacks] = useState({ injection: true, jailbreak: true, leakage: true, drift: false });

  const modes = [
    { key: "red_team", label: "Red team", desc: "Attack only, no defense", color: "#fb7185", bg: "rgba(244,63,94,0.12)", border: "rgba(251,113,133,0.3)" },
    { key: "blue_team", label: "Blue team", desc: "Defense only, measure FP rate", color: "#7dd3fc", bg: "rgba(14,165,233,0.12)", border: "rgba(56,189,248,0.3)" },
    { key: "adversarial", label: "Adversarial", desc: "Red + blue simultaneous", color: "#c4b5fd", bg: "rgba(139,92,246,0.12)", border: "rgba(139,92,246,0.3)" },
  ];

  const attackTypes = [
    { key: "injection", label: "Prompt injection", desc: "Direct, indirect, encoding bypass", owasp: "#1", gradient: "linear-gradient(90deg, #ef4444, #f97316, #ec4899)" },
    { key: "jailbreak", label: "Jailbreak", desc: "Role-play, crescendo, mutation", owasp: "—", gradient: "linear-gradient(90deg, #ec4899, #a855f7, #6366f1)" },
    { key: "leakage", label: "Data leakage", desc: "System prompt, PII, credentials", owasp: "#2", gradient: "linear-gradient(90deg, #8b5cf6, #6366f1, #3b82f6)" },
    { key: "drift", label: "Constraint drift", desc: "Context inflation, multi-turn erosion", owasp: "—", gradient: "linear-gradient(90deg, #6366f1, #3b82f6, #06b6d4)" },
  ];

  const defenseAgents = [
    { label: "Input guard", desc: "Detect injection patterns" },
    { label: "Output filter", desc: "Block leakage in responses" },
    { label: "Behavior monitor", desc: "Anomalous tool calling" },
    { label: "Constraint persistence", desc: "Verify safety instructions" },
  ];

  const Dot = ({ color, size = 7 }) => (
    <span style={{ height: size, width: size, borderRadius: "50%", background: color, boxShadow: `0 0 10px ${color}`, flexShrink: 0 }} />
  );

  const nav = ["Dashboard", "Scans", "Reports", "Judge", "Monitoring", "Settings"];

  return (
    <div style={{ minHeight: "100vh", background: "#08090e", color: "#fff", fontFamily: "'Inter', system-ui, -apple-system, sans-serif" }}>
      <div style={{ display: "grid", gridTemplateColumns: "240px 1fr", minHeight: "100vh" }}>

        {/* ═══ SIDEBAR ═══ */}
        <aside style={{
          borderRight: "1px solid rgba(255,255,255,0.06)",
          background: "linear-gradient(180deg, rgba(255,255,255,0.025) 0%, rgba(255,255,255,0.01) 100%)",
          padding: "28px 20px", display: "flex", flexDirection: "column",
        }}>
          <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "36px" }}>
            <div style={{ height: 40, width: 40, borderRadius: "14px", border: "1px solid rgba(217,70,239,0.25)", background: "linear-gradient(135deg, rgba(217,70,239,0.15), rgba(139,92,246,0.1))", display: "flex", alignItems: "center", justifyContent: "center", boxShadow: "0 0 24px rgba(217,70,239,0.15)" }}>
              <div style={{ height: 14, width: 14, borderRadius: 4, background: "linear-gradient(135deg, #d946ef, #8b5cf6)" }} />
            </div>
            <div>
              <div style={{ fontWeight: 700, fontSize: 16, letterSpacing: "-0.01em" }}>AgentShield</div>
              <div style={{ fontSize: 9, textTransform: "uppercase", letterSpacing: "0.18em", color: "#525252", marginTop: 2 }}>Adversarial evaluation</div>
            </div>
          </div>
          <nav style={{ display: "flex", flexDirection: "column", gap: "3px" }}>
            {nav.map((item, i) => {
              const active = i === 1;
              return (
                <div key={item} style={{ padding: "10px 14px", borderRadius: "12px", fontSize: "13px", display: "flex", alignItems: "center", gap: "10px", cursor: "pointer", border: active ? "1px solid rgba(217,70,239,0.18)" : "1px solid transparent", background: active ? "linear-gradient(135deg, rgba(217,70,239,0.1), rgba(139,92,246,0.06))" : "transparent", color: active ? "#e9d5ff" : "#525252" }}>
                  <Dot color={active ? "#c084fc" : "#3f3f3f"} size={7} />{item}
                </div>
              );
            })}
          </nav>
          <div style={{ marginTop: "auto", borderTop: "1px solid rgba(255,255,255,0.06)", paddingTop: "20px" }}>
            <div style={{ fontSize: 9, textTransform: "uppercase", letterSpacing: "0.2em", color: "#3f3f3f", marginBottom: 14 }}>System status</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10, fontSize: 12, color: "#737373" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}><Dot color="#34d399" size={7} />Orchestrator online</div>
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}><Dot color="#34d399" size={7} />8 agents ready</div>
            </div>
          </div>
        </aside>

        {/* ═══ MAIN ═══ */}
        <main style={{ background: "radial-gradient(ellipse at 10% 0%, rgba(217,70,239,0.08) 0%, transparent 50%), radial-gradient(ellipse at 90% 0%, rgba(59,130,246,0.06) 0%, transparent 40%), #0a0b10", padding: "28px 32px", overflowY: "auto" }}>

          {/* Header */}
          <div style={{ marginBottom: 28 }}>
            <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: "-0.02em" }}>Create new scan</h1>
            <p style={{ fontSize: 12, color: "#3f3f3f", marginTop: 4 }}>Configure target, mode, and attack agents</p>
          </div>

          {/* ── Target endpoint ── */}
          <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px", marginBottom: 20 }}>
            <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 12 }}>Target API endpoint</div>
            <div style={{ display: "flex", gap: 12 }}>
              <div style={{ flex: 1, height: 44, borderRadius: 10, border: "1px solid rgba(255,255,255,0.08)", background: "rgba(255,255,255,0.03)", display: "flex", alignItems: "center", padding: "0 16px" }}>
                <span style={{ fontSize: 13, color: "#525252" }}>https://api.docmind.dev/v1/chat</span>
              </div>
              <button style={{ height: 44, borderRadius: 10, border: "1px solid rgba(52,211,153,0.25)", background: "rgba(52,211,153,0.08)", padding: "0 20px", fontSize: 12, color: "#34d399", cursor: "pointer", whiteSpace: "nowrap" }}>Test connection</button>
            </div>
          </div>

          {/* ── Testing mode ── */}
          <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px", marginBottom: 20 }}>
            <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 16 }}>Testing mode</div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
              {modes.map((m) => {
                const selected = mode === m.key;
                return (
                  <div key={m.key} onClick={() => setMode(m.key)} style={{
                    borderRadius: 12, padding: "16px",
                    border: selected ? `1px solid ${m.border}` : "1px solid rgba(255,255,255,0.06)",
                    background: selected ? m.bg : "rgba(255,255,255,0.02)",
                    cursor: "pointer", transition: "all 0.15s",
                  }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                      <Dot color={selected ? m.color : "#3f3f3f"} size={8} />
                      <span style={{ fontSize: 13, fontWeight: 600, color: selected ? m.color : "#737373" }}>{m.label}</span>
                    </div>
                    <div style={{ fontSize: 11, color: "#525252", lineHeight: 1.4 }}>{m.desc}</div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* ── Attack agents ── */}
          <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px", marginBottom: 20 }}>
            <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 16 }}>Red team agents</div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 12 }}>
              {attackTypes.map((a) => {
                const checked = attacks[a.key];
                return (
                  <div key={a.key} onClick={() => setAttacks(p => ({ ...p, [a.key]: !p[a.key] }))} style={{
                    borderRadius: 12, padding: "16px",
                    border: checked ? "1px solid rgba(255,255,255,0.1)" : "1px solid rgba(255,255,255,0.04)",
                    background: checked ? "rgba(255,255,255,0.04)" : "rgba(255,255,255,0.015)",
                    cursor: "pointer", transition: "all 0.15s",
                  }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                      <span style={{ fontSize: 13, fontWeight: 600, color: checked ? "#fff" : "#525252" }}>{a.label}</span>
                      <div style={{
                        height: 20, width: 20, borderRadius: 5,
                        border: checked ? "none" : "1px solid rgba(255,255,255,0.12)",
                        background: checked ? a.gradient : "transparent",
                        display: "flex", alignItems: "center", justifyContent: "center",
                        fontSize: 11, color: "#fff", fontWeight: 700,
                      }}>{checked ? "✓" : ""}</div>
                    </div>
                    <div style={{ fontSize: 11, color: "#525252", lineHeight: 1.4 }}>{a.desc}</div>
                    {a.owasp !== "—" && (
                      <div style={{ marginTop: 8, fontSize: 10, color: "#3f3f3f" }}>OWASP LLM {a.owasp}</div>
                    )}
                    {/* Mini gradient bar */}
                    <div style={{ marginTop: 10, height: 3, borderRadius: 2, background: "rgba(255,255,255,0.04)", overflow: "hidden" }}>
                      <div style={{ height: "100%", borderRadius: 2, background: checked ? a.gradient : "rgba(255,255,255,0.04)", width: "100%", opacity: checked ? 0.7 : 0.2, transition: "opacity 0.3s" }} />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* ── Blue team agents (visible when adversarial or blue_team) ── */}
          {(mode === "adversarial" || mode === "blue_team") && (
            <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))", padding: "20px", marginBottom: 20 }}>
              <div style={{ fontSize: 10, textTransform: "uppercase", letterSpacing: "0.2em", color: "#525252", marginBottom: 16 }}>Blue team agents <span style={{ color: "#34d399" }}>• all active</span></div>
              <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 12 }}>
                {defenseAgents.map((d) => (
                  <div key={d.label} style={{ borderRadius: 12, padding: "14px 16px", border: "1px solid rgba(52,211,153,0.15)", background: "rgba(52,211,153,0.04)" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                      <Dot color="#34d399" size={6} />
                      <span style={{ fontSize: 13, fontWeight: 600, color: "#a3a3a3" }}>{d.label}</span>
                    </div>
                    <div style={{ fontSize: 11, color: "#525252" }}>{d.desc}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* ── Start scan ── */}
          <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
            <button style={{
              height: 48, borderRadius: 12, padding: "0 32px", fontSize: 14, fontWeight: 600, cursor: "pointer",
              border: "1px solid rgba(217,70,239,0.3)",
              background: "linear-gradient(135deg, rgba(217,70,239,0.2), rgba(139,92,246,0.12))",
              color: "#e9d5ff",
              boxShadow: "0 0 24px rgba(217,70,239,0.15), 0 4px 16px rgba(0,0,0,0.3)",
            }}>Start scan</button>
            <span style={{ fontSize: 12, color: "#3f3f3f" }}>Estimated: ~3 min · {Object.values(attacks).filter(Boolean).length} attack agents · {mode === "adversarial" ? "4 defense agents" : mode === "blue_team" ? "4 defense agents" : "no defense"}</span>
          </div>

        </main>
      </div>
    </div>
  );
}
