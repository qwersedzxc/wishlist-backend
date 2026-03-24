import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import { resolveUrl } from '../utils';
import { Dialog } from 'primereact/dialog';
import { Button } from 'primereact/button';
import { InputText } from 'primereact/inputtext';
import { InputNumber } from 'primereact/inputnumber';
import { ProgressSpinner } from 'primereact/progressspinner';
import { confirmDialog } from 'primereact/confirmdialog';
import { FileUpload } from 'primereact/fileupload';
import ItemDetailDialog from './ItemDetailDialog';

const API_BASE = 'http://localhost:8081/api/v1';

function ItemsDialog({ wishlist, visible, onHide, toast, readOnly = false }) {
  const canAdd = !readOnly;
  
  // Определяем является ли текущий пользователь владельцем вишлиста
  const getCurrentUserId = () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return null;
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.user_id;
    } catch {
      return null;
    }
  };
  
  const currentUserId = getCurrentUserId();
  const isOwner = currentUserId && wishlist.userId && currentUserId === wishlist.userId;
  
  console.log('ItemsDialog:', { currentUserId, wishlistUserId: wishlist.userId, isOwner });

  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showAddForm, setShowAddForm] = useState(false);
  const [uploadLoading, setUploadLoading] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);
  const fileUploadRef = useRef(null);
  const [formData, setFormData] = useState({ title: '', url: '', imageUrl: '', price: null, priority: 5 });

  useEffect(() => {
    if (visible) loadItems();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible, wishlist.id]);

  const getAuthHeaders = () => {
    const token = localStorage.getItem('token');
    return token ? { Authorization: `Bearer ${token}` } : {};
  };

  const loadItems = async () => {
    try {
      setLoading(true);
      const r = await axios.get(`${API_BASE}/wishlists/${wishlist.id}/items?page=1&per_page=100`, {
        headers: getAuthHeaders()
      });
      setItems(r.data.items || []);
    } catch {
      toast.current?.show({ severity: 'error', summary: 'Ошибка загрузки', life: 3000 });
    } finally {
      setLoading(false);
    }
  };

  const handleAddItem = async (e) => {
    e.preventDefault();
    try {
      await axios.post(`${API_BASE}/wishlists/${wishlist.id}/items`, 
        { ...formData, wishlistId: wishlist.id },
        { headers: getAuthHeaders() }
      );
      setShowAddForm(false);
      setFormData({ title: '', url: '', imageUrl: '', price: null, priority: 5 });
      loadItems();
      toast.current?.show({ severity: 'success', summary: 'Элемент добавлен', life: 3000 });
    } catch {
      toast.current?.show({ severity: 'error', summary: 'Ошибка', life: 3000 });
    }
  };

  const handleImageUpload = async (event) => {
    const file = event.files[0];
    if (!file) return;
    console.log('ItemsDialog: Загружаем файл:', file.name, file.size, file.type);
    setUploadLoading(true);
    const fd = new FormData();
    fd.append('file', file);
    try {
      console.log('ItemsDialog: Отправляем на сервер...');
      const r = await axios.post(`${API_BASE}/upload/image`, fd, { 
        headers: { 
          'Content-Type': 'multipart/form-data',
          ...getAuthHeaders()
        } 
      });
      console.log('ItemsDialog: Файл загружен:', r.data);
      setFormData(prev => ({ ...prev, imageUrl: r.data.url }));
      if (fileUploadRef.current) fileUploadRef.current.clear();
      toast.current?.show({ severity: 'success', summary: 'Фото загружено', life: 3000 });
    } catch (error) {
      console.error('ItemsDialog: Ошибка загрузки фото:', error);
      toast.current?.show({ severity: 'error', summary: 'Ошибка загрузки фото', life: 3000 });
    } finally {
      setUploadLoading(false);
    }
  };

  const handleTogglePurchased = async (item) => {
    try {
      await axios.patch(`${API_BASE}/wishlists/${wishlist.id}/items/${item.id}`, 
        { isPurchased: !item.isPurchased },
        { headers: getAuthHeaders() }
      );
      loadItems();
    } catch {}
  };

  const handleReserve = async (item, isIncognito = false) => {
    try {
      const headers = getAuthHeaders();
      console.log('Reserve headers:', headers);
      await axios.post(`${API_BASE}/wishlists/${wishlist.id}/items/${item.id}/reserve`, 
        { isIncognito },
        { headers }
      );
      loadItems();
      toast.current?.show({ severity: 'success', summary: isIncognito ? 'Забронировано инкогнито' : 'Забронировано', life: 3000 });
    } catch (error) {
      console.error('Reserve error:', error.response?.data || error.message);
      toast.current?.show({ severity: 'error', summary: 'Ошибка бронирования', life: 3000 });
    }
  };

  const handleUnreserve = async (item) => {
    try {
      await axios.delete(`${API_BASE}/wishlists/${wishlist.id}/items/${item.id}/reserve`, {
        headers: getAuthHeaders()
      });
      loadItems();
      toast.current?.show({ severity: 'success', summary: 'Бронирование снято', life: 3000 });
    } catch (error) {
      toast.current?.show({ severity: 'error', summary: 'Ошибка', life: 3000 });
    }
  };

  const handleDeleteItem = (itemId) => {
    confirmDialog({
      message: 'Удалить этот элемент?',
      header: 'Подтверждение',
      acceptLabel: 'Да',
      rejectLabel: 'Нет',
      accept: async () => {
        try {
          await axios.delete(`${API_BASE}/wishlists/${wishlist.id}/items/${itemId}`, {
            headers: getAuthHeaders()
          });
          loadItems();
        } catch {}
      }
    });
  };

  return (
    <Dialog visible={visible} onHide={onHide} style={{ width: '600px' }} header={null} closable>
      <div className="flex align-items-center justify-content-between mb-4">
        <div className="flex align-items-center gap-3">
          {wishlist.imageUrl ? (
            <img src={resolveUrl(wishlist.imageUrl)} alt={wishlist.title} style={{ width: '56px', height: '56px', objectFit: 'cover', borderRadius: '8px', border: '1px solid #e0e0e0' }} />
          ) : (
            <div style={{ width: '56px', height: '56px', background: '#f0f0f0', borderRadius: '8px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <span style={{ color: '#999', fontSize: '12px' }}>нет фото</span>
            </div>
          )}
          <span style={{ fontSize: '1.2rem', fontWeight: 600 }}>{wishlist.title}</span>
        </div>
        {canAdd && (
          <Button
            label={showAddForm ? 'Отмена' : 'Добавить подарок'}
            icon={showAddForm ? 'pi pi-times' : 'pi pi-plus'}
            onClick={() => setShowAddForm(v => !v)}
            className={showAddForm ? 'p-button-text' : ''}
            style={{ background: showAddForm ? 'transparent' : '#000', color: showAddForm ? '#333' : '#fff', border: '1px solid #000' }}
          />
        )}
      </div>

      {canAdd && showAddForm && (
        <form onSubmit={handleAddItem} className="mb-4 p-3" style={{ background: '#f9f9f9', borderRadius: '8px' }}>
          <div className="flex flex-column gap-3">
            <InputText value={formData.title} onChange={e => setFormData({ ...formData, title: e.target.value })} placeholder="Название *" required className="w-full" />
            <InputText value={formData.url} onChange={e => setFormData({ ...formData, url: e.target.value })} placeholder="Ссылка на товар" className="w-full" />
            <InputNumber value={formData.price} onValueChange={e => setFormData({ ...formData, price: e.value })} placeholder="Цена (₽)" className="w-full" minFractionDigits={0} />
            <FileUpload ref={fileUploadRef} mode="basic" name="file" accept="image/*" maxFileSize={5000000} onSelect={handleImageUpload} chooseLabel="Фото товара" disabled={uploadLoading} />
            {formData.imageUrl && (
              <img src={resolveUrl(formData.imageUrl)} alt="preview" style={{ width: '80px', height: '80px', objectFit: 'cover', borderRadius: '6px' }} />
            )}
            <Button type="submit" label="Добавить" icon="pi pi-check" disabled={uploadLoading} />
          </div>
        </form>
      )}

      {loading ? (
        <div className="flex justify-content-center p-4"><ProgressSpinner /></div>
      ) : items.length === 0 ? (
        <p className="text-center text-color-secondary">Список пуст. Добавьте первый элемент!</p>
      ) : (
        <div className="flex flex-column gap-2">
          {items.map(item => {
            console.log('Item:', item.id, { isOwner, isReserved: item.isReserved, reservedByMe: item.reservedByMe });
            return (
            <div key={item.id} style={{
              display: 'flex', alignItems: 'center', gap: '12px',
              padding: '10px 12px', border: '1px solid #e0e0e0', borderRadius: '8px',
              background: item.isPurchased ? '#f5f5f5' : '#fff',
              opacity: item.isPurchased ? 0.7 : 1,
            }}>
              {item.imageUrl ? (
                <img 
                  src={resolveUrl(item.imageUrl)} 
                  alt={item.title} 
                  style={{ 
                    width: '44px', 
                    height: '44px', 
                    objectFit: 'cover', 
                    borderRadius: '6px', 
                    flexShrink: 0,
                    cursor: 'pointer'
                  }}
                  onClick={() => setSelectedItem(item)}
                />
              ) : (
                <div 
                  style={{ 
                    width: '44px', 
                    height: '44px', 
                    background: '#f0f0f0', 
                    borderRadius: '6px', 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center', 
                    flexShrink: 0,
                    cursor: 'pointer'
                  }}
                  onClick={() => setSelectedItem(item)}
                >
                  <i className="pi pi-image" style={{ color: '#ccc' }}></i>
                </div>
              )}
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 600, textDecoration: item.isPurchased ? 'line-through' : 'none', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {item.title}
                  {item.isReserved && !item.reservedByMe && <span style={{ marginLeft: '8px', fontSize: '12px', color: '#ff9800' }}>🔒 Забронировано</span>}
                  {item.reservedByMe && <span style={{ marginLeft: '8px', fontSize: '12px', color: '#4caf50' }}>✓ Вы забронировали</span>}
                </div>
                {item.price && <div style={{ fontSize: '13px', color: '#666' }}>{item.price} ₽</div>}
              </div>
              {item.url && (
                <a href={item.url} target="_blank" rel="noopener noreferrer">
                  <Button icon="pi pi-external-link" className="p-button-text p-button-sm" />
                </a>
              )}
              {!isOwner && (
                <>
                  {item.reservedByMe ? (
                    <Button
                      icon="pi pi-times"
                      className="p-button-sm p-button-warning"
                      tooltip="Снять бронь"
                      onClick={() => handleUnreserve(item)}
                    />
                  ) : !item.isReserved ? (
                    <>
                      <Button
                        icon="pi pi-bookmark"
                        className="p-button-sm p-button-success"
                        tooltip="Забронировать"
                        onClick={() => handleReserve(item, false)}
                      />
                      <Button
                        icon="pi pi-eye-slash"
                        className="p-button-sm p-button-help"
                        tooltip="Забронировать инкогнито"
                        onClick={() => handleReserve(item, true)}
                      />
                    </>
                  ) : null}
                  <Button
                    icon={item.isPurchased ? 'pi pi-undo' : 'pi pi-gift'}
                    className={`p-button-sm ${item.isPurchased ? 'p-button-warning' : 'p-button-info'}`}
                    tooltip={item.isPurchased ? 'Отменить' : 'Подарю это'}
                    onClick={() => handleTogglePurchased(item)}
                  />
                </>
              )}
              {isOwner && (
                <>
                  <Button icon={item.isPurchased ? 'pi pi-undo' : 'pi pi-check'} className={`p-button-sm ${item.isPurchased ? 'p-button-warning' : 'p-button-success'}`} onClick={() => handleTogglePurchased(item)} />
                  <Button icon="pi pi-trash" className="p-button-sm p-button-danger p-button-text" onClick={() => handleDeleteItem(item.id)} />
                </>
              )}
            </div>
            );
          })}
        </div>
      )}

      <ItemDetailDialog
        item={selectedItem}
        visible={!!selectedItem}
        onHide={() => setSelectedItem(null)}
      />
    </Dialog>
  );
}

export default ItemsDialog;
