import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import useAuth from './useAuth';

function FullScreenMessage({ title, detail }) {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: 32,
        background:
          'radial-gradient(circle at top, rgba(139,92,246,0.14), transparent 35%), #06070b',
        color: '#fff',
      }}
    >
      <div
        style={{
          width: 'min(520px, 100%)',
          borderRadius: 18,
          padding: 28,
          border: '1px solid rgba(255,255,255,0.08)',
          background: 'rgba(255,255,255,0.03)',
        }}
      >
        <div style={{ fontSize: 22, fontWeight: 700, marginBottom: 10 }}>{title}</div>
        <div style={{ fontSize: 14, color: '#a3a3a3', lineHeight: 1.6 }}>{detail}</div>
      </div>
    </div>
  );
}

export default function ProtectedRoute() {
  const location = useLocation();
  const { loading, session, isConfigured } = useAuth();

  if (loading) {
    return <FullScreenMessage title="Restoring session" detail="Checking for an existing Supabase session." />;
  }

  if (!isConfigured) {
    return <Navigate to="/login" replace state={{ from: location, reason: 'missing-config' }} />;
  }

  if (!session) {
    return <Navigate to="/login" replace state={{ from: location, reason: 'unauthorized' }} />;
  }

  return <Outlet />;
}
