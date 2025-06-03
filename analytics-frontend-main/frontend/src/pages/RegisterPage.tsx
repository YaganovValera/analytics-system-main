import { useNavigate } from 'react-router-dom';
import api, { setAccessToken } from '@api/axios';
import AuthForm from '@components/AuthForm';
import { useAuth } from '@context/AuthContext';

function RegisterPage() {
  const navigate = useNavigate();
  const { setUser } = useAuth();

  const handleRegister = async (username: string, password: string) => {
    const res = await api.post('/register', { username, password });
    localStorage.setItem('refresh_token', res.data.refresh_token);
    setAccessToken(res.data.access_token);

    const userRes = await api.get('/me');
    setUser({
      user_id: userRes.data.user_id,
      roles: userRes.data.roles,
    });

    navigate('/me');
  };

  return <AuthForm mode="register" onSubmit={handleRegister} />;
}

export default RegisterPage;
