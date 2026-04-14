import React from 'react';

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

const navItems = ['Dashboard', 'Scans', 'Reports', 'Judge', 'Monitoring', 'Settings'];

const statusItems = [
  { label: 'Open Settings for live health checks', color: '#60a5fa' },
  { label: 'Monitoring page shows live scan events', color: '#a78bfa' },
];

const Sidebar = ({ activeIndex, onNavigate }) => {
  return (
    <aside
      style={{
        borderRight: '1px solid rgba(255,255,255,0.06)',
        background: 'linear-gradient(180deg, rgba(255,255,255,0.025) 0%, rgba(255,255,255,0.01) 100%)',
        padding: '28px 20px',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 36 }}>
        <div
          style={{
            height: 40,
            width: 40,
            borderRadius: 14,
            border: '1px solid rgba(217,70,239,0.25)',
            background: 'linear-gradient(135deg, rgba(217,70,239,0.15), rgba(139,92,246,0.1))',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 0 24px rgba(217,70,239,0.15)',
          }}
        >
          <div style={{ height: 14, width: 14, borderRadius: 4, background: 'linear-gradient(135deg, #d946ef, #8b5cf6)' }} />
        </div>
        <div>
          <div style={{ fontWeight: 700, fontSize: 16, letterSpacing: '-0.01em' }}>AgentShield</div>
          <div style={{ fontSize: 9, textTransform: 'uppercase', letterSpacing: '0.18em', color: '#525252', marginTop: 2 }}>
            Adversarial evaluation
          </div>
        </div>
      </div>

      <nav style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        {navItems.map((item, i) => {
          const active = i === activeIndex;
          return (
            <button
              type="button"
              key={item}
              onClick={() => onNavigate && onNavigate(item.toLowerCase())}
              style={{
                padding: '10px 14px',
                borderRadius: 12,
                fontSize: 13,
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                cursor: 'pointer',
                border: active ? '1px solid rgba(217,70,239,0.18)' : '1px solid transparent',
                background: active ? 'linear-gradient(135deg, rgba(217,70,239,0.1), rgba(139,92,246,0.06))' : 'transparent',
                color: active ? '#e9d5ff' : '#737373',
                textAlign: 'left',
              }}
            >
              <Dot color={active ? '#c084fc' : '#3f3f3f'} size={7} />
              {item}
            </button>
          );
        })}
      </nav>

      <div style={{ marginTop: 'auto', borderTop: '1px solid rgba(255,255,255,0.06)', paddingTop: 20 }}>
        <div style={{ fontSize: 9, textTransform: 'uppercase', letterSpacing: '0.2em', color: '#3f3f3f', marginBottom: 14 }}>
          System status
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, fontSize: 12, color: '#a3a3a3' }}>
          {statusItems.map((item) => (
            <div key={item.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Dot color={item.color} size={7} />
              {item.label}
            </div>
          ))}
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;
