import { Link } from 'react-router-dom';
import { useAuth } from '@context/AuthContext';
import './Header.css';

function Header() {
  const { user, isAuthenticated, logout } = useAuth();

  if (!isAuthenticated) return null;

  const handleLogout = () => {
    logout();
  };

  return (
    <header className="site-header">
      <nav>
        <Link to="/me">Профиль</Link>
        {user?.roles.includes('admin') && <Link to="/admin">Админка</Link>}
        <Link to="/candles/historical">Аналитика</Link>
        <Link to="/analysis/offline">Анализ csv файлов</Link>
        <Link to="/orderbook/historical">Orderbook</Link>
      </nav>
      <div className="user-info">
        <span>👤 {user?.roles}</span>
        <button className="logout-button" onClick={handleLogout}>Выйти</button>
      </div>
    </header>
  );
}

export default Header;
