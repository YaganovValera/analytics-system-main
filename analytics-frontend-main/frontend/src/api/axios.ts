import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const setAccessToken = (token: string) => {
  localStorage.setItem('access_token', token);
  api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
};

export const clearAccessToken = () => {
  localStorage.removeItem('access_token');
  delete api.defaults.headers.common['Authorization'];
};

const storedAccess = localStorage.getItem('access_token');
if (storedAccess) {
  setAccessToken(storedAccess);
}

// === interceptor для auto-refresh ===
let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach((cb) => cb(newToken));
  refreshSubscribers = [];
}

function addRefreshSubscriber(callback: (token: string) => void) {
  refreshSubscribers.push(callback);
}

api.interceptors.response.use(
  response => response,
  async (error) => {
    const originalRequest = error.config;

    // если ошибка не 401 или уже был попытка — пробрасываем дальше
    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }

    // помечаем, что это повторная попытка
    originalRequest._retry = true;

    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      clearAccessToken();
      return Promise.reject(error);
    }

    if (isRefreshing) {
      // пока обновляется — ждём
      return new Promise((resolve) => {
        addRefreshSubscriber((token) => {
          originalRequest.headers['Authorization'] = `Bearer ${token}`;
          resolve(api(originalRequest));
        });
      });
    }

    isRefreshing = true;

    try {
      const res = await axios.post('http://localhost:8080/v1/refresh', { refresh_token: refreshToken });
      const { access_token, refresh_token: newRefresh } = res.data;

      setAccessToken(access_token);
      localStorage.setItem('refresh_token', newRefresh);

      onTokenRefreshed(access_token);
      return api(originalRequest);
    } catch (err) {
      clearAccessToken();
      return Promise.reject(err);
    } finally {
      isRefreshing = false;
    }
  }
);

export default api;
