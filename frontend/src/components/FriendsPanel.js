import React, { useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import { Button } from 'primereact/button';
import { InputText } from 'primereact/inputtext';
import { TabView, TabPanel } from 'primereact/tabview';
import { ProgressSpinner } from 'primereact/progressspinner';
import { Avatar } from 'primereact/avatar';
import UserProfileDialog from './UserProfileDialog';

const API_BASE = 'http://localhost:8081/api/v1';

// Получаем headers с токеном авторизации
function getAuthHeaders() {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// Получаем ID текущего пользователя из токена
function getCurrentUserId() {
  try {
    const token = localStorage.getItem('token');
    if (!token) return null;
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.user_id;
  } catch {
    return null;
  }
}

function FriendsPanel({ toast }) {
  const [friends, setFriends] = useState([]);
  const [pendingRequests, setPendingRequests] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [searching, setSearching] = useState(false);
  const [sentRequests, setSentRequests] = useState(new Set());
  const [viewProfile, setViewProfile] = useState(null);
  const searchTimer = useRef(null);
  const currentUserId = getCurrentUserId();

  const openProfile = async (userId) => {
    try {
      const r = await axios.get(`${API_BASE}/users/${userId}`, {
        headers: getAuthHeaders()
      });
      setViewProfile(r.data);
    } catch {}
  };

  const loadFriends = useCallback(async () => {
    try {
      const r = await axios.get(`${API_BASE}/friends`, {
        headers: getAuthHeaders()
      });
      setFriends(r.data.friends || []);
    } catch {}
  }, []);

  const loadPending = useCallback(async () => {
    try {
      const r = await axios.get(`${API_BASE}/friends/requests`, {
        headers: getAuthHeaders()
      });
      setPendingRequests(r.data.requests || []);
    } catch {}
  }, []);

  useEffect(() => {
    loadFriends();
    loadPending();
  }, [loadFriends, loadPending]);

  // Live-поиск с дебаунсом 400ms
  useEffect(() => {
    if (!searchQuery.trim()) {
      setSearchResults([]);
      return;
    }
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(async () => {
      setSearching(true);
      try {
        const r = await axios.get(`${API_BASE}/users/search?q=${encodeURIComponent(searchQuery)}`);
        setSearchResults(r.data.users || []);
      } catch {
        toast.current?.show({ severity: 'error', summary: 'Ошибка поиска', life: 2000 });
      } finally {
        setSearching(false);
      }
    }, 400);
    return () => clearTimeout(searchTimer.current);
  }, [searchQuery, toast]);

  const friendIds = new Set(friends.map(f => f.id));

  const sendRequest = async (friendId) => {
    try {
      const headers = getAuthHeaders();
      console.log('Sending friend request with headers:', headers);
      await axios.post(`${API_BASE}/friends/request`, { friendId }, { headers });
      setSentRequests(prev => new Set([...prev, friendId]));
      toast.current?.show({ severity: 'success', summary: 'Запрос отправлен', life: 2000 });
    } catch (e) {
      console.error('Friend request error:', e.response?.data || e.message);
      toast.current?.show({ severity: 'error', summary: e.response?.data?.error || 'Ошибка', life: 2000 });
    }
  };

  const acceptRequest = async (friendshipId) => {
    try {
      await axios.post(`${API_BASE}/friends/requests/${friendshipId}/accept`, {}, {
        headers: getAuthHeaders()
      });
      toast.current?.show({ severity: 'success', summary: 'Запрос принят', life: 2000 });
      loadFriends();
      loadPending();
    } catch {}
  };

  const rejectRequest = async (friendshipId) => {
    try {
      await axios.post(`${API_BASE}/friends/requests/${friendshipId}/reject`, {}, {
        headers: getAuthHeaders()
      });
      toast.current?.show({ severity: 'info', summary: 'Запрос отклонён', life: 2000 });
      loadPending();
    } catch {}
  };

  const removeFriend = async (friendId) => {
    try {
      await axios.delete(`${API_BASE}/friends/${friendId}`, {
        headers: getAuthHeaders()
      });
      toast.current?.show({ severity: 'info', summary: 'Друг удалён', life: 2000 });
      loadFriends();
    } catch {}
  };

  const getSearchButtonState = (userId) => {
    if (userId === currentUserId) return { disabled: true, label: 'Это вы', icon: 'pi pi-user' };
    if (friendIds.has(userId)) return { disabled: true, label: 'Уже друг', icon: 'pi pi-check' };
    if (sentRequests.has(userId)) return { disabled: true, label: 'Запрос отправлен', icon: 'pi pi-clock' };
    return { disabled: false, label: 'Добавить', icon: 'pi pi-user-plus' };
  };

  return (
    <div>
      <TabView>
        <TabPanel header={`Мои друзья (${friends.length})`}>
          {friends.length === 0 ? (
            <p className="text-color-secondary text-center mt-4">
              Нет друзей. Найдите пользователей во вкладке "Найти друзей".
            </p>
          ) : (
            <div className="flex flex-column gap-2 mt-2">
              {friends.map(f => (
                <div key={f.friendshipId} className="flex align-items-center justify-content-between p-3 border-1 border-round surface-border">
                  <div className="flex align-items-center gap-3">
                    <Avatar
                      label={f.username?.charAt(0).toUpperCase()}
                      shape="circle"
                      style={{ background: '#e0e0e0', color: '#333', cursor: 'pointer' }}
                      onClick={() => openProfile(f.id)}
                    />
                    <div>
                      <div className="font-semibold">{f.username}</div>
                      <div className="text-color-secondary text-sm">{f.email}</div>
                    </div>
                  </div>
                  <Button
                    icon="pi pi-user-minus"
                    label="Удалить"
                    className="p-button-text p-button-danger p-button-sm"
                    onClick={() => removeFriend(f.id)}
                  />
                </div>
              ))}
            </div>
          )}
        </TabPanel>

        <TabPanel header={`Запросы${pendingRequests.length > 0 ? ` (${pendingRequests.length})` : ''}`}>
          {pendingRequests.length === 0 ? (
            <p className="text-color-secondary text-center mt-4">Нет входящих запросов.</p>
          ) : (
            <div className="flex flex-column gap-2 mt-2">
              {pendingRequests.map(req => (
                <div key={req.friendshipId} className="flex align-items-center justify-content-between p-3 border-1 border-round surface-border">
                  <div className="flex align-items-center gap-3">
                    <Avatar
                      label={req.username?.charAt(0).toUpperCase()}
                      shape="circle"
                      style={{ background: '#e0e0e0', color: '#333', cursor: 'pointer' }}
                      onClick={() => openProfile(req.id)}
                    />
                    <div>
                      <div className="font-semibold">{req.username}</div>
                      <div className="text-color-secondary text-sm">{req.email}</div>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button icon="pi pi-check" label="Принять" className="p-button-success p-button-sm" onClick={() => acceptRequest(req.friendshipId)} />
                    <Button icon="pi pi-times" label="Отклонить" className="p-button-text p-button-danger p-button-sm" onClick={() => rejectRequest(req.friendshipId)} />
                  </div>
                </div>
              ))}
            </div>
          )}
        </TabPanel>

        <TabPanel header="Найти друзей">
          <div className="mt-2">
            <span className="p-input-icon-left w-full">
              <i className={searching ? 'pi pi-spin pi-spinner' : 'pi pi-search'} />
              <InputText
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                placeholder="Введите имя или email..."
                className="w-full"
              />
            </span>

            <div className="flex flex-column gap-2 mt-3">
              {searchResults
                .filter(u => u.id !== currentUserId)
                .map(u => {
                  const btn = getSearchButtonState(u.id);
                  return (
                    <div key={u.id} className="flex align-items-center justify-content-between p-3 border-1 border-round surface-border">
                      <div className="flex align-items-center gap-3">
                        <Avatar
                          label={u.username?.charAt(0).toUpperCase()}
                          shape="circle"
                          style={{ background: '#e0e0e0', color: '#333', cursor: 'pointer' }}
                          onClick={() => openProfile(u.id)}
                        />
                        <div>
                          <div className="font-semibold">{u.username}</div>
                          <div className="text-color-secondary text-sm">{u.email}</div>
                        </div>
                      </div>
                      <Button
                        icon={btn.icon}
                        label={btn.label}
                        disabled={btn.disabled}
                        className="p-button-sm"
                        onClick={() => sendRequest(u.id)}
                      />
                    </div>
                  );
                })}

              {!searching && searchQuery.trim() && searchResults.filter(u => u.id !== currentUserId).length === 0 && (
                <p className="text-color-secondary text-center mt-3">Пользователи не найдены.</p>
              )}

              {!searchQuery.trim() && (
                <p className="text-color-secondary text-center mt-3">
                  Начните вводить имя или email для поиска.
                </p>
              )}
            </div>
          </div>
        </TabPanel>
      </TabView>
      <UserProfileDialog profile={viewProfile} onHide={() => setViewProfile(null)} />
    </div>
  );
}

export default FriendsPanel;
