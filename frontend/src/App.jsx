import { useState } from 'react';
import Layout from './Layout';
import DashboardContent from './DashboardContent';
import NewScanContent from './NewScanContent';
import ScanMonitorContent from './ScanMonitorContent';

function App() {
  const [currentPage, setCurrentPage] = useState('dashboard');

  const getActiveIndex = () => {
    const pages = ['dashboard', 'scans', 'reports', 'judge', 'monitoring', 'settings'];
    return pages.indexOf(currentPage);
  };

  const renderContent = () => {
    switch(currentPage) {
      case 'dashboard':
        return <DashboardContent />;
      case 'scans':
        return <NewScanContent />;
      case 'reports':
        return <ScanMonitorContent />;
      default:
        return <DashboardContent />;
    }
  };

  return (
    <Layout activeIndex={getActiveIndex()} onNavigate={setCurrentPage}>
      {renderContent()}
    </Layout>
  );
}

export default App;
