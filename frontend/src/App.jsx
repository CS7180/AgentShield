import { Navigate, Outlet, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import Layout from './Layout';
import DashboardContent from './DashboardContent';
import NewScanContent from './NewScanContent';
import ScanMonitorContent from './ScanMonitorContent';
import LandingPage from './LandingPage';
import LoginPage from './LoginPage';
import ProtectedRoute from './auth/ProtectedRoute';
import useAuth from './auth/useAuth';

const routeIndexByPath = {
  '/dashboard': 0,
  '/scans': 1,
  '/reports': 2,
  '/judge': 3,
  '/monitoring': 4,
  '/settings': 5,
};

function ComingSoonContent({ title, description }) {
  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontSize: 26, fontWeight: 700, margin: 0, letterSpacing: '-0.02em' }}>{title}</h1>
        <p style={{ fontSize: 12, color: '#525252', marginTop: 6 }}>{description}</p>
      </div>
      <div
        style={{
          borderRadius: 16,
          border: '1px solid rgba(255,255,255,0.06)',
          background: 'linear-gradient(135deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))',
          padding: 24,
        }}
      >
        <div style={{ fontSize: 12, color: '#a3a3a3', lineHeight: 1.6 }}>
          This route exists now, but the product flow for this page has not been implemented yet.
        </div>
      </div>
    </div>
  );
}

function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const activeIndex = routeIndexByPath[location.pathname] ?? 0;

  return (
    <Layout activeIndex={activeIndex} onNavigate={(page) => navigate(`/${page}`)}>
      <Outlet />
    </Layout>
  );
}

function MarketingEntry() {
  const navigate = useNavigate();
  const { session } = useAuth();

  return (
    <LandingPage
      onLogIn={() => navigate(session ? '/dashboard' : '/login')}
      onGetStarted={() => navigate(session ? '/scans' : '/login')}
      onViewDemo={() => navigate('/reports')}
    />
  );
}

function LoginRoute() {
  const { session, loading, isConfigured } = useAuth();

  if (!loading && isConfigured && session) {
    return <Navigate to="/dashboard" replace />;
  }

  return <LoginPage />;
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<MarketingEntry />} />
      <Route path="/login" element={<LoginRoute />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/dashboard" element={<DashboardContent />} />
          <Route path="/scans" element={<NewScanContent />} />
          <Route path="/reports" element={<ScanMonitorContent />} />
          <Route
            path="/judge"
            element={<ComingSoonContent title="Judge" description="Calibration and judge quality views will live here." />}
          />
          <Route
            path="/monitoring"
            element={<ComingSoonContent title="Monitoring" description="System health and runtime observability will live here." />}
          />
          <Route
            path="/settings"
            element={<ComingSoonContent title="Settings" description="Workspace and integration settings will live here." />}
          />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default App;
