import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { ApolloProvider } from '@apollo/client';
import { SnackbarProvider } from 'notistack';
import { ThemeProvider, createTheme } from '@mui/material';
import { client } from './graphql/client';
import TemplateList from './pages/TemplateList';
import TemplateDetails from './pages/TemplateDetails';
import CreateTemplate from './pages/CreateTemplate';
import { Logo } from './components/logo';
import './App.css';

import { debugCorsConnection } from './utils/corsDebug';
import { useEffect, useState } from 'react';

/* ---- фирменная тема Go-cyan ---- */
const theme = createTheme({
  palette: {
    primary: { main: '#00ADD8' },
    success: { main: '#2e7d32' },
    error:   { main: '#d32f2f' }
  },
  typography: { fontFamily: 'Inter, sans-serif' }
});

export default function App() {
  const [apiStatus, setApiStatus] = useState<{ success?: boolean; message?: string } | null>(null);

  /* Dev-корс дебаг */
  useEffect(() => {
    if (import.meta.env.DEV) {
      (window as any).debugApi = async () => debugCorsConnection();
      console.info('%c🔍 Пиши debugApi() в консоли, чтобы проверить API',
        'color:#00ADD8;font-weight:bold');
    }
  }, []);

  return (
    <ApolloProvider client={client}>
      <ThemeProvider theme={theme}>
        <SnackbarProvider maxSnack={3}>
          <Router>
            <div className="app-container">
              {/* ---------- SIDEBAR ---------- */}
              <aside className="app-sidebar">
                <div className="sidebar-header">
                  {/* PNG-герой на десктопе */}
                  <Logo size={120} png className="logo-desktop" />
                  {/* Маленькое SVG на мобиле */}
                  <Logo size={48} className="logo-mobile" />
                </div>

                <nav className="sidebar-nav">
                  <Link to="/create" className="sidebar-button primary">
                    Gen New Template
                  </Link>
                  <Link to="/" className="sidebar-button">
                    Recent Templates
                  </Link>

                  {import.meta.env.DEV && (
                    <button
                      className={`sidebar-button ${
                        apiStatus?.success
                          ? 'success'
                          : apiStatus?.success === false
                          ? 'error'
                          : ''
                      }`}
                      onClick={async () => {
                        try {
                          const { testCorsConnection } = await import('./utils/corsDebug');
                          const res = await testCorsConnection();
                          setApiStatus(res);
                        } catch (e: any) {
                          setApiStatus({ success: false, message: e.message });
                        }
                      }}
                    >
                      Проверить API
                    </button>
                  )}
                </nav>
              </aside>

              {/* ---------- MAIN ---------- */}
              <main className="app-content">
                {apiStatus && (
                  <div
                    className={`cors-notification ${
                      apiStatus.success ? 'success' : 'error'
                    }`}
                  >
                    {apiStatus.success
                      ? '✅ API доступен'
                      : `❌ Проблема с API: ${apiStatus.message}`}
                  </div>
                )}

                <Routes>
                  <Route path="/" element={<TemplateList />} />
                  <Route path="/template/:id" element={<TemplateDetails />} />
                  <Route path="/create" element={<CreateTemplate />} />
                </Routes>
              </main>

              {/* ---------- FOOTER ---------- */}
              <footer className="app-footer">
                <p>&copy; 2025 Go init — template generator</p>
              </footer>
            </div>
          </Router>
        </SnackbarProvider>
      </ThemeProvider>
    </ApolloProvider>
  );
}
