import { Navigate, Outlet, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import Layout from './Layout';
import DashboardContent from './DashboardContent';
import NewScanContent from './NewScanContent';
import ScanMonitorContent from './ScanMonitorContent';
import ReportCompareContent from './ReportCompareContent';
import JudgeContent from './JudgeContent';
import SettingsContent from './SettingsContent';
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
          <Route path="/reports" element={<ReportCompareContent />} />
          <Route path="/judge" element={<JudgeContent />} />
          <Route path="/monitoring" element={<ScanMonitorContent />} />
          <Route path="/settings" element={<SettingsContent />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default App;
