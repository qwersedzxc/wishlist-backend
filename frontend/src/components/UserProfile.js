import React, { useState, useRef, useEffect } from 'react';
import { Avatar } from 'primereact/avatar';
import { Button } from 'primereact/button';
import { Dialog } from 'primereact/dialog';
import { InputText } from 'primereact/inputtext';
import { InputTextarea } from 'primereact/inputtextarea';

function UserProfile({ user, onLogout, onUpdateProfile }) {
  const [open, setOpen] = useState(false);
  const [showProfileDialog, setShowProfileDialog] = useState(false);
  const dropdownRef = useRef(null);

  const [profileData, setProfileData] = useState({
    fullName: '',
    birthDate: '',
    bio: '',
    phone: '',
    city: '',
  });

  // Синхронизируем данные при открытии диалога
  useEffect(() => {
    if (showProfileDialog && user) {
      setProfileData({
        fullName: user.fullName || '',
        birthDate: user.birthDate ? user.birthDate.substring(0, 10) : '',
        bio: user.bio || '',
        phone: user.phone || '',
        city: user.city || '',
      });
    }
  }, [showProfileDialog, user]);

  useEffect(() => {
    if (!open) return;
    const handler = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const handleUpdateProfile = async (e) => {
    e.preventDefault();
    try {
      const payload = {
        fullName: profileData.fullName || null,
        bio: profileData.bio || null,
        phone: profileData.phone || null,
        city: profileData.city || null,
        birthDate: profileData.birthDate || null,
      };
      await onUpdateProfile(payload);
      setShowProfileDialog(false);
    } catch {}
  };

  const menuItems = [
    { label: 'Мой профиль', icon: 'pi pi-user', action: () => { setShowProfileDialog(true); setOpen(false); } },
    { separator: true },
    { label: 'Выйти', icon: 'pi pi-sign-out', action: () => { onLogout(); setOpen(false); }, danger: true },
  ];

  const fieldStyle = { marginBottom: '16px' };
  const labelStyle = { display: 'block', marginBottom: '6px', fontSize: '13px', color: '#555' };

  return (
    <>
      <div ref={dropdownRef} style={{ position: 'relative' }}>
        <div
          className="flex align-items-center gap-2 cursor-pointer p-2 border-round hover:surface-100"
          style={{ cursor: 'pointer' }}
        >
          <div onClick={() => { setShowProfileDialog(true); setOpen(false); }}>
            <Avatar
              label={user?.username?.charAt(0).toUpperCase()}
              shape="circle"
              size="normal"
              style={{ background: '#000', color: '#fff', cursor: 'pointer' }}
            />
          </div>
          <span className="font-medium" onClick={() => setOpen(v => !v)}>{user?.username}</span>
          <i className={`pi pi-angle-${open ? 'up' : 'down'}`} onClick={() => setOpen(v => !v)}></i>
        </div>

        {open && (
          <div style={{
            position: 'absolute', right: 0, top: '100%', marginTop: '4px',
            background: '#fff', border: '1px solid #e0e0e0', borderRadius: '6px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.12)', minWidth: '180px', zIndex: 9999,
          }}>
            {menuItems.map((item, i) =>
              item.separator ? (
                <div key={i} style={{ borderTop: '1px solid #e0e0e0', margin: '4px 0' }} />
              ) : (
                <div
                  key={i}
                  onClick={item.action}
                  style={{
                    padding: '10px 16px', cursor: 'pointer', display: 'flex',
                    alignItems: 'center', gap: '10px',
                    color: item.danger ? '#e53935' : '#333', fontSize: '14px',
                  }}
                  onMouseEnter={e => e.currentTarget.style.background = '#f5f5f5'}
                  onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
                >
                  <i className={item.icon}></i>
                  {item.label}
                </div>
              )
            )}
          </div>
        )}
      </div>

      <Dialog
        header="Редактировать профиль"
        visible={showProfileDialog}
        style={{ width: '440px' }}
        onHide={() => setShowProfileDialog(false)}
      >
        {/* Информация о пользователе */}
        <div style={{ marginBottom: '20px', padding: '12px', background: '#f9f9f9', borderRadius: '6px' }}>
          <div style={{ fontSize: '13px', color: '#888', marginBottom: '4px' }}>Аккаунт</div>
          <div style={{ fontWeight: 500 }}>{user?.username}</div>
          <div style={{ fontSize: '13px', color: '#666' }}>{user?.email}</div>
        </div>

        <form onSubmit={handleUpdateProfile}>
          <div style={fieldStyle}>
            <label style={labelStyle}>Полное имя</label>
            <InputText
              style={{ width: '100%' }}
              value={profileData.fullName}
              onChange={e => setProfileData({ ...profileData, fullName: e.target.value })}
              placeholder="Иван Иванов"
            />
          </div>

          <div style={fieldStyle}>
            <label style={labelStyle}>Дата рождения</label>
            <InputText
              type="date"
              style={{ width: '100%' }}
              value={profileData.birthDate}
              onChange={e => setProfileData({ ...profileData, birthDate: e.target.value })}
            />
          </div>

          <div style={fieldStyle}>
            <label style={labelStyle}>Город</label>
            <InputText
              style={{ width: '100%' }}
              value={profileData.city}
              onChange={e => setProfileData({ ...profileData, city: e.target.value })}
              placeholder="Москва"
            />
          </div>

          <div style={fieldStyle}>
            <label style={labelStyle}>Телефон</label>
            <InputText
              style={{ width: '100%' }}
              value={profileData.phone}
              onChange={e => setProfileData({ ...profileData, phone: e.target.value })}
              placeholder="+7 900 000 00 00"
            />
          </div>

          <div style={fieldStyle}>
            <label style={labelStyle}>О себе</label>
            <InputTextarea
              style={{ width: '100%' }}
              rows={3}
              value={profileData.bio}
              onChange={e => setProfileData({ ...profileData, bio: e.target.value })}
              placeholder="Расскажите о себе..."
            />
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '8px' }}>
            <Button label="Отмена" className="p-button-text" type="button" onClick={() => setShowProfileDialog(false)} />
            <Button type="submit" label="Сохранить" />
          </div>
        </form>
      </Dialog>
    </>
  );
}

export default UserProfile;
