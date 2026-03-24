import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { resolveUrl } from './utils';
import { Button } from 'primereact/button';
import { DataView } from 'primereact/dataview';
import { Card } from 'primereact/card';
import { ProgressSpinner } from 'primereact/progressspinner';
import { Toast } from 'primereact/toast';
import { ConfirmDialog, confirmDialog } from 'primereact/confirmdialog';
import { TabView, TabPanel } from 'primereact/tabview';
import { Avatar } from 'primereact/avatar';
import AuthDialog from './components/AuthDialog';
import UserProfile from './components/UserProfile';
import UserProfileDialog from './components/UserProfileDialog';
import WishlistDialog from './components/WishlistDialog';
import ItemsDialog from './components/ItemsDialog';
import FriendsPanel from './components/FriendsPanel';
import FriendWishlists from './components/FriendWishlists';

import 'primereact/resources/themes/lara-light-blue/theme.css';
import 'primereact/resources/primereact.min.css';
import 'primeicons/primeicons.css';
import 'primeflex/primeflex.css';
import './App.css';

const API_BASE = 'http://localhost:8081/api/v1';

function MyWishlists({ user, toast, wishlistTemplate }) {
  const [wishlists, setWishlists] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) return;
    setLoading(true);
    axios.get(`${API_BASE}/wishlists?user_id=${user.id}&page=1&per_page=50`)
      .then(r => setWishlists(r.data.items || []))
      .catch(() => toast.current?.show({ severity: 'error', summary: 'Ошибка загрузки', life: 2000 }))
      .finally(() => setLoading(false));
  }, [user, toast]);

  if (loading) return <div className="flex justify-content-center p-4"><ProgressSpinner /></div>;

  if (wishlists.length === 0) return (
    <Card className="text-center">
      <h3>Нет вишлистов</h3>
      <p className="text-color-secondary">Вы ещё не создали ни одного вишлиста.</p>
    </Card>
  );

  return <DataView value={wishlists} itemTemplate={wishlistTemplate} layout="grid" />;
}

function App() {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [showAuthDialog, setShowAuthDialog] = useState(false);
  const [wishlists, setWishlists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [selectedWishlist, setSelectedWishlist] = useState(null);
  const [authorProfile, setAuthorProfile] = useState(null);
  const toast = React.useRef(null);

  // Восстанавливаем сессию при загрузке / обновлении страницы
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get('token');

    if (urlToken) {
      // После OAuth редиректа — сохраняем токен
      localStorage.setItem('token', urlToken);
      setToken(urlToken);
      axios.defaults.headers.common['Authorization'] = `Bearer ${urlToken}`;
      window.history.replaceState({}, document.title, '/');
      axios.get(`${API_BASE}/auth/me`).then(r => {
        setUser(r.data.user || r.data);
      }).catch(() => {
        localStorage.removeItem('token');
        setToken(null);
      });
      return;
    }

    const savedToken = localStorage.getItem('token');
    if (savedToken) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${savedToken}`;
      axios.get(`${API_BASE}/auth/me`).then(r => {
        setToken(savedToken);
        setUser(r.data.user || r.data);
      }).catch(() => {
        localStorage.removeItem('token');
        setToken(null);
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Настройка axios при смене токена
  useEffect(() => {
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      localStorage.setItem('token', token);
    } else {
      delete axios.defaults.headers.common['Authorization'];
    }

    // Добавляем interceptor для гарантии что токен всегда передается
    const interceptor = axios.interceptors.request.use(
      (config) => {
        const currentToken = localStorage.getItem('token');
        if (currentToken && !config.headers['Authorization']) {
          config.headers['Authorization'] = `Bearer ${currentToken}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    return () => {
      axios.interceptors.request.eject(interceptor);
    };
  }, [token]);

  useEffect(() => {
    loadWishlists();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const handleLogin = (userData, authToken) => {
    setUser(userData);
    setToken(authToken);
    // Устанавливаем заголовок сразу, не дожидаясь useEffect
    axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
    toast.current?.show({
      severity: 'success',
      summary: 'Добро пожаловать!',
      detail: `Вы вошли как ${userData.username}`,
      life: 3000
    });
  };

  const handleLogout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    sessionStorage.setItem('logged_out', 'true');
    toast.current?.show({
      severity: 'info',
      summary: 'До свидания!',
      detail: 'Вы вышли из системы',
      life: 3000
    });
  };

  const handleUpdateProfile = async (profileData) => {
    try {
      const response = await axios.patch(`${API_BASE}/auth/profile`, profileData);
      setUser(response.data.user);
      toast.current?.show({
        severity: 'success',
        summary: 'Успешно',
        detail: 'Профиль обновлён',
        life: 3000
      });
    } catch (error) {
      toast.current?.show({
        severity: 'error',
        summary: 'Ошибка',
        detail: error.response?.data?.error || 'Не удалось обновить профиль',
        life: 3000
      });
      throw error;
    }
  };

  const loadWishlists = async () => {
    if (!token) {
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      const response = await axios.get(`${API_BASE}/wishlists?page=1&per_page=50`);
      setWishlists(response.data.items || []);
    } catch (error) {
      console.error('Ошибка загрузки вишлистов:', error);
      if (error.response?.status === 401) {
        handleLogout();
        toast.current?.show({
          severity: 'warn',
          summary: 'Сессия истекла',
          detail: 'Пожалуйста, войдите снова',
          life: 3000
        });
      } else {
        toast.current?.show({
          severity: 'error',
          summary: 'Ошибка',
          detail: 'Не удалось загрузить вишлисты',
          life: 3000
        });
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCreateWishlist = async (data) => {
    try {
      console.log('User при создании:', user);
      console.log('Token:', localStorage.getItem('token') ? 'есть' : 'нет');
      console.log('Axios заголовок:', axios.defaults.headers.common['Authorization'] ? 'есть' : 'нет');
      
      const wishlistData = {
        title: data.title,
        description: data.description || undefined,
        eventName: data.eventName || undefined,
        eventDate: data.eventDate || undefined,
        privacyLevel: data.privacyLevel || 'friends_only',
        isPublic: data.isPublic || false,
        imageUrl: data.imageUrl || undefined,
      };
      console.log('Создаем вишлист с данными:', wishlistData);
      
      await axios.post(`${API_BASE}/wishlists`, wishlistData);
      
      setShowCreateDialog(false);
      loadWishlists();
      toast.current?.show({
        severity: 'success',
        summary: 'Успешно',
        detail: 'Вишлист создан',
        life: 3000
      });
    } catch (error) {
      console.error('Ошибка создания вишлиста:', error);
      console.error('Детали ошибки:', error.response?.data);
      console.error('Статус:', error.response?.status);
      console.error('Заголовки запроса:', error.config?.headers);
      const errMsg = error.response?.data?.error || error.response?.data?.message || `Ошибка ${error.response?.status || ''}: Не удалось создать вишлист`;
      toast.current?.show({
        severity: 'error',
        summary: 'Ошибка',
        detail: errMsg,
        life: 5000
      });
    }
  };

  const handleDeleteWishlist = (id) => {
    confirmDialog({
      message: 'Вы уверены, что хотите удалить этот вишлист?',
      header: 'Подтверждение удаления',
      icon: 'pi pi-exclamation-triangle',
      acceptLabel: 'Да',
      rejectLabel: 'Нет',
      accept: async () => {
        try {
          await axios.delete(`${API_BASE}/wishlists/${id}`);
          loadWishlists();
          toast.current?.show({
            severity: 'success',
            summary: 'Успешно',
            detail: 'Вишлист удален',
            life: 3000
          });
        } catch (error) {
          console.error('Ошибка удаления вишлиста:', error);
          toast.current?.show({
            severity: 'error',
            summary: 'Ошибка',
            detail: 'Не удалось удалить вишлист',
            life: 3000
          });
        }
      }
    });
  };

  const wishlistTemplate = (wishlist) => {
    return (
      <Card
        title={
          <div className="flex align-items-center gap-2">
            <span>{wishlist.title}</span>
          </div>
        }
        subTitle={wishlist.description}
        className="m-2"
        footer={
          <div className="flex gap-2">
            <Button
              label="Открыть"
              icon="pi pi-eye"
              onClick={() => setSelectedWishlist(wishlist)}
              className="p-button-primary"
            />
            <Button
              label="Удалить"
              icon="pi pi-trash"
              onClick={() => handleDeleteWishlist(wishlist.id)}
              className="p-button-danger"
            />
          </div>
        }
      >
        {wishlist.imageUrl && (
          <div className="mb-3">
            <img 
              src={resolveUrl(wishlist.imageUrl)} 
              alt={wishlist.title}
              style={{ 
                width: '100%', 
                maxHeight: '200px', 
                objectFit: 'cover',
                borderRadius: '4px'
              }}
              onError={(e) => {
                e.target.style.display = 'none';
              }}
            />
          </div>
        )}
        {wishlist.author && (
          <div
            className="flex align-items-center gap-2 mb-2"
            style={{ fontSize: '13px', color: '#666', cursor: 'pointer' }}
            onClick={() => setAuthorProfile(wishlist.author)}
          >
            <Avatar
              label={wishlist.author.username?.charAt(0).toUpperCase()}
              shape="circle"
              size="small"
              style={{ background: '#e0e0e0', color: '#333', width: '24px', height: '24px', fontSize: '11px' }}
            />
            <span>{wishlist.author.fullName || wishlist.author.username}</span>
          </div>
        )}
        <div className="flex align-items-center gap-3">
          <span className={`pi ${wishlist.isPublic ? 'pi-globe' : 'pi-lock'}`}></span>
          <span>{wishlist.isPublic ? 'Публичный' : 'Приватный'}</span>
        </div>
        {wishlist.eventName && (
          <div className="flex align-items-center gap-3 mt-2">
            <span className="pi pi-calendar"></span>
            <span>{wishlist.eventName}</span>
            {wishlist.eventDate && (
              <span className="text-color-secondary">
                ({new Date(wishlist.eventDate).toLocaleDateString('ru-RU')})
              </span>
            )}
          </div>
        )}
      </Card>
    );
  };

  return (
    <div className="app">
      <Toast ref={toast} />
      <ConfirmDialog />

      <div className="surface-0 shadow-2 p-4 border-round mb-4">
        <div className="flex justify-content-between align-items-center">
          <div className="flex align-items-center gap-3">
            <h1 className="m-0">Вишлисты</h1>
          </div>
          <div className="flex gap-2 align-items-center">
            {user ? (
              <>
                <Button
                  label="Создать вишлист"
                  icon="pi pi-plus"
                  onClick={() => setShowCreateDialog(true)}
                  className="p-button-lg"
                />
                <UserProfile
                  user={user}
                  onLogout={handleLogout}
                  onUpdateProfile={handleUpdateProfile}
                />
              </>
            ) : (
              <Button
                label="Войти"
                icon="pi pi-sign-in"
                onClick={() => setShowAuthDialog(true)}
                className="p-button-lg"
              />
            )}
          </div>
        </div>
      </div>

      {!user ? (
        <Card className="text-center">
          <i className="pi pi-user" style={{ fontSize: '4rem', color: '#ccc' }}></i>
          <h3>Добро пожаловать!</h3>
          <p className="text-color-secondary">Войдите или зарегистрируйтесь, чтобы создавать вишлисты</p>
          <Button
            label="Войти / Регистрация"
            icon="pi pi-sign-in"
            onClick={() => setShowAuthDialog(true)}
            size="large"
          />
        </Card>
      ) : (
        <TabView pt={{ panelContainer: { style: { padding: '1rem 0' } }, tab: { style: { marginRight: '8px' } } }}>
          <TabPanel header="Вишлисты">
            {loading ? (
              <div className="flex justify-content-center align-items-center" style={{ minHeight: '300px' }}>
                <ProgressSpinner />
              </div>
            ) : wishlists.length === 0 ? (
              <Card className="text-center">
                <h3>Нет вишлистов</h3>
                <p className="text-color-secondary">Создайте свой первый вишлист!</p>
                <Button label="Создать вишлист" icon="pi pi-plus" onClick={() => setShowCreateDialog(true)} />
              </Card>
            ) : (
              <DataView value={wishlists} itemTemplate={wishlistTemplate} layout="grid" />
            )}
          </TabPanel>

          <TabPanel header="Мои вишлисты">
            <MyWishlists
              user={user}
              toast={toast}
              wishlistTemplate={wishlistTemplate}
            />
          </TabPanel>

          <TabPanel header="Вишлисты друзей">
            <FriendWishlists toast={toast} />
          </TabPanel>

          <TabPanel header="Друзья">
            <FriendsPanel toast={toast} />
          </TabPanel>
        </TabView>
      )}

      <AuthDialog
        visible={showAuthDialog}
        onHide={() => setShowAuthDialog(false)}
      />

      <WishlistDialog
        visible={showCreateDialog}
        onHide={() => setShowCreateDialog(false)}
        onSubmit={handleCreateWishlist}
      />

      {selectedWishlist && (
        <ItemsDialog
          wishlist={selectedWishlist}
          visible={!!selectedWishlist}
          onHide={() => setSelectedWishlist(null)}
          toast={toast}
        />
      )}

      <UserProfileDialog profile={authorProfile} onHide={() => setAuthorProfile(null)} />
    </div>
  );
}

export default App;