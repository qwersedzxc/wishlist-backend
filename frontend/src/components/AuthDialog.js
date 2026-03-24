import React from 'react';
import { Dialog } from 'primereact/dialog';
import { Button } from 'primereact/button';

function AuthDialog({ visible, onHide }) {
  const handleGitHubLogin = () => {
    const afterLogout = sessionStorage.getItem('logged_out') === 'true';
    sessionStorage.removeItem('logged_out');
    const url = afterLogout
      ? 'http://localhost:8081/api/v1/auth/oauth/login?prompt=true'
      : 'http://localhost:8081/api/v1/auth/oauth/login';
    window.location.href = url;
  };

  return (
    <Dialog
      header="Вход"
      visible={visible}
      style={{ width: '360px' }}
      onHide={onHide}
    >
      <div style={{ textAlign: 'center', padding: '1rem 0' }}>
        <p style={{ marginBottom: '1.5rem', color: '#555' }}>
          Войдите через GitHub, чтобы продолжить
        </p>
        <Button
          label="Войти через GitHub"
          icon="pi pi-github"
          onClick={handleGitHubLogin}
          style={{ width: '100%', background: '#24292e', border: 'none' }}
        />
      </div>
    </Dialog>
  );
}

export default AuthDialog;
