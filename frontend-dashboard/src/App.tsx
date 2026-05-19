import { useState, useEffect } from 'react';
import Layout from './components/Layout';
import Login from './components/Login';
import type { User } from './types';

export default function App() {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const stored = localStorage.getItem('ondc_user');
    if (stored) {
      try { setUser(JSON.parse(stored)); } catch { /* ignore */ }
    }
  }, []);

  const handleLogin = (u: User, token: string) => {
    localStorage.setItem('ondc_token', token);
    localStorage.setItem('ondc_user', JSON.stringify(u));
    setUser(u);
  };

  const handleLogout = () => {
    localStorage.removeItem('ondc_token');
    localStorage.removeItem('ondc_user');
    setUser(null);
  };

  if (!user) return <Login onLogin={handleLogin} />;
  return <Layout user={user} onLogout={handleLogout} />;
}
