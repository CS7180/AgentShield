import React from 'react';
import Sidebar from './Sidebar';

const Layout = ({ activeIndex, onNavigate, children }) => {
  return (
    <div style={{ minHeight: "100vh", background: "#08090e", color: "#fff", fontFamily: "'Inter', system-ui, -apple-system, sans-serif" }}>
      <div style={{ display: "grid", gridTemplateColumns: "240px 1fr", minHeight: "100vh" }}>
        <Sidebar activeIndex={activeIndex} onNavigate={onNavigate} />
        <main style={{ background: "radial-gradient(ellipse at 10% 0%, rgba(217,70,239,0.08) 0%, transparent 50%), radial-gradient(ellipse at 90% 0%, rgba(59,130,246,0.06) 0%, transparent 40%), #0a0b10", padding: "28px 32px", overflowY: "auto" }}>
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;