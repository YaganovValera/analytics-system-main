import { Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from '@pages/LoginPage';
import RegisterPage from '@pages/RegisterPage';
import MePage from '@pages/MePage';
import PrivateRoute from '@routes/PrivateRoute';
import { useAuth } from '@context/AuthContext';
import Header from '@components/Header';
import HistoricalCandlesPage from '@pages/candles/HistoricalCandlesPage';
import OfflineAnalysisPage from '@pages/analysis/OfflineAnalysisPage';
import OrderBookPage from '@pages/orderbook/OrderBookPage';
import AdminDashboardPage from '@pages/admin/AdminDashboardPage';


function App() {
  const { initialized } = useAuth();

  if (!initialized) return <p>Загрузка авторизации...</p>;

  return (
    <>
      <Header />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/me"
          element={
            <PrivateRoute>
              <MePage />
            </PrivateRoute>
          }
        />
        <Route
          path="/candles/historical"
          element={
            <PrivateRoute>
              <HistoricalCandlesPage />
            </PrivateRoute>
          }
        />
        <Route
          path="/analysis/offline"
          element={
            <PrivateRoute>
              <OfflineAnalysisPage />
            </PrivateRoute>}
        />
        <Route
          path="/orderbook/historical"
          element={
            <PrivateRoute>
              <OrderBookPage />
            </PrivateRoute>}
        />
        <Route
          path="/admin"
          element={
            <PrivateRoute>
              <AdminDashboardPage />
            </PrivateRoute>
          }
        />
        <Route path="/" element={<Navigate to="/login" replace />} />
      </Routes>
    </>
  );
}

export default App;
