import React from 'react';
import { Dialog } from 'primereact/dialog';
import { Avatar } from 'primereact/avatar';

function UserProfileDialog({ profile, onHide }) {
  return (
    <Dialog
      header="Профиль пользователя"
      visible={!!profile}
      style={{ width: '380px' }}
      onHide={onHide}
    >
      {profile && (
        <div style={{ padding: '0.5rem 0' }}>
          <div style={{ textAlign: 'center', marginBottom: '20px' }}>
            <Avatar
              label={profile.username?.charAt(0).toUpperCase()}
              shape="circle"
              size="xlarge"
              style={{ background: '#000', color: '#fff', marginBottom: '12px' }}
            />
            <div style={{ fontWeight: 600, fontSize: '18px' }}>
              {profile.fullName || profile.username}
            </div>
            {profile.fullName && (
              <div style={{ color: '#888', fontSize: '13px', marginTop: '2px' }}>
                @{profile.username}
              </div>
            )}
          </div>

          {(profile.city || profile.phone || profile.birthDate || profile.bio) && (
            <div style={{ borderTop: '1px solid #eee', paddingTop: '16px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {profile.city && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '14px' }}>
                  <i className="pi pi-map-marker" style={{ color: '#888', width: '16px' }}></i>
                  <span>{profile.city}</span>
                </div>
              )}
              {profile.phone && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '14px' }}>
                  <i className="pi pi-phone" style={{ color: '#888', width: '16px' }}></i>
                  <span>{profile.phone}</span>
                </div>
              )}
              {profile.birthDate && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '14px' }}>
                  <i className="pi pi-calendar" style={{ color: '#888', width: '16px' }}></i>
                  <span>{new Date(profile.birthDate).toLocaleDateString('ru-RU')}</span>
                </div>
              )}
              {profile.bio && (
                <div style={{ display: 'flex', gap: '10px', fontSize: '14px' }}>
                  <i className="pi pi-info-circle" style={{ color: '#888', width: '16px', marginTop: '2px' }}></i>
                  <span style={{ color: '#444' }}>{profile.bio}</span>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </Dialog>
  );
}

export default UserProfileDialog;
