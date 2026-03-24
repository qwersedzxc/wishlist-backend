import React from 'react';
import { resolveUrl } from '../utils';
import { Dialog } from 'primereact/dialog';
import { Avatar } from 'primereact/avatar';
import { Button } from 'primereact/button';

function ItemDetailDialog({ item, visible, onHide }) {
  if (!item) return null;

  return (
    <Dialog
      visible={visible}
      onHide={onHide}
      style={{ width: '500px' }}
      header="Детали подарка"
      closable
    >
      <div className="flex flex-column gap-3">
        {/* Изображение */}
        {item.imageUrl ? (
          <img
            src={resolveUrl(item.imageUrl)}
            alt={item.title}
            style={{
              width: '100%',
              maxHeight: '300px',
              objectFit: 'contain',
              borderRadius: '8px',
              border: '1px solid #e0e0e0'
            }}
          />
        ) : (
          <div style={{
            width: '100%',
            height: '200px',
            background: '#f0f0f0',
            borderRadius: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <i className="pi pi-image" style={{ fontSize: '3rem', color: '#ccc' }}></i>
          </div>
        )}

        {/* Название */}
        <div>
          <h3 style={{ margin: '0 0 8px 0', fontSize: '1.5rem' }}>{item.title}</h3>
          {item.description && (
            <p style={{ margin: 0, color: '#666', fontSize: '14px' }}>{item.description}</p>
          )}
        </div>

        {/* Цена */}
        {item.price && (
          <div style={{
            padding: '12px',
            background: '#f9f9f9',
            borderRadius: '6px',
            fontSize: '1.2rem',
            fontWeight: 600
          }}>
            {item.price} ₽
          </div>
        )}

        {/* Ссылка */}
        {item.url && (
          <a
            href={item.url}
            target="_blank"
            rel="noopener noreferrer"
            style={{ textDecoration: 'none' }}
          >
            <Button
              label="Открыть ссылку на товар"
              icon="pi pi-external-link"
              className="w-full"
              style={{ background: '#000', border: '1px solid #000' }}
            />
          </a>
        )}

        {/* Информация о бронировании */}
        {item.isReserved && item.reservedBy && !item.isIncognitoReservation && (
          <div style={{
            padding: '12px',
            background: '#fff3e0',
            borderRadius: '6px',
            border: '1px solid #ffb74d'
          }}>
            <div style={{ fontWeight: 600, marginBottom: '8px', color: '#f57c00' }}>
              🔒 Забронировано
            </div>
            <div className="flex align-items-center gap-2">
              <Avatar
                image={item.reservedBy.avatarUrl ? resolveUrl(item.reservedBy.avatarUrl) : null}
                label={!item.reservedBy.avatarUrl ? item.reservedBy.username[0].toUpperCase() : null}
                shape="circle"
                size="normal"
                style={{ background: '#666', color: '#fff' }}
              />
              <div>
                <div style={{ fontWeight: 600 }}>
                  {item.reservedBy.fullName || item.reservedBy.username}
                </div>
                <div style={{ fontSize: '12px', color: '#666' }}>
                  @{item.reservedBy.username}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Инкогнито бронирование */}
        {item.isReserved && item.isIncognitoReservation && (
          <div style={{
            padding: '12px',
            background: '#f3e5f5',
            borderRadius: '6px',
            border: '1px solid #ba68c8',
            color: '#7b1fa2',
            fontWeight: 600
          }}>
            🔒 Забронировано инкогнито
          </div>
        )}

        {/* Забронировано мной */}
        {item.reservedByMe && (
          <div style={{
            padding: '12px',
            background: '#e8f5e9',
            borderRadius: '6px',
            border: '1px solid #66bb6a',
            color: '#2e7d32',
            fontWeight: 600
          }}>
            ✓ Вы забронировали этот подарок
          </div>
        )}

        {/* Куплено */}
        {item.isPurchased && (
          <div style={{
            padding: '12px',
            background: '#e3f2fd',
            borderRadius: '6px',
            border: '1px solid #42a5f5',
            color: '#1565c0',
            fontWeight: 600
          }}>
            🎁 Подарок куплен
          </div>
        )}
      </div>
    </Dialog>
  );
}

export default ItemDetailDialog;
