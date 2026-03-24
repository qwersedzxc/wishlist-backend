import React, { useState } from 'react';
import { resolveUrl } from '../utils';
import { Dialog } from 'primereact/dialog';
import { InputText } from 'primereact/inputtext';
import { InputTextarea } from 'primereact/inputtextarea';
import { Calendar } from 'primereact/calendar';
import { Dropdown } from 'primereact/dropdown';
import { Button } from 'primereact/button';
import { FileUpload } from 'primereact/fileupload';
import axios from 'axios';

function WishlistDialog({ visible, onHide, onSubmit }) {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    eventName: '',
    eventDate: null,
    privacyLevel: 'friends_only',
    imageUrl: ''
  });
  const [uploadedImage, setUploadedImage] = useState(null);
  const [previewImage, setPreviewImage] = useState(null);

  const privacyOptions = [
    { label: 'Публичный (виден всем)', value: 'public' },
    { label: 'Только друзья', value: 'friends_only' },
    { label: 'По ссылке', value: 'link_only' }
  ];

  const handleSubmit = (e) => {
    e.preventDefault();
    const submitData = {
      title: formData.title,
      description: formData.description || undefined,
      eventName: formData.eventName || undefined,
      eventDate: formData.eventDate ? formData.eventDate.toISOString().split('T')[0] : undefined,
      privacyLevel: formData.privacyLevel || 'friends_only',
      isPublic: formData.privacyLevel === 'public',
      imageUrl: uploadedImage || undefined,
    };
    onSubmit(submitData);
    setFormData({ title: '', description: '', eventName: '', eventDate: null, privacyLevel: 'friends_only', imageUrl: '' });
    setUploadedImage(null);
    setPreviewImage(null);
  };

  const handleImageUpload = async (event) => {
    const file = event.files[0];
    if (!file) return;
    
    console.log('Загружаем файл:', file.name, file.size, file.type);
    
    // Показываем локальное превью сразу
    const reader = new FileReader();
    reader.onloadend = () => {
      console.log('Превью готово');
      setPreviewImage(reader.result);
    };
    reader.readAsDataURL(file);
    
    const formDataUpload = new FormData();
    formDataUpload.append('file', file);
    try {
      const token = axios.defaults.headers.common['Authorization']?.replace('Bearer ', '');
      console.log('Отправляем на сервер...');
      const response = await fetch('http://localhost:8081/api/v1/upload/image', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formDataUpload
      });
      console.log('Ответ сервера:', response.status);
      if (response.ok) {
        const data = await response.json();
        console.log('Файл загружен:', data);
        setUploadedImage(data.url);
      } else {
        const errorText = await response.text();
        console.error('Ошибка загрузки:', response.status, errorText);
        alert('Ошибка загрузки изображения: ' + errorText);
      }
    } catch (error) {
      console.error('Ошибка загрузки изображения:', error);
      alert('Ошибка загрузки изображения: ' + error.message);
    }
  };

  const footer = (
    <div>
      <Button label="Отмена" icon="pi pi-times" onClick={onHide} className="p-button-text" />
      <Button label="Создать" icon="pi pi-check" onClick={handleSubmit} autoFocus />
    </div>
  );

  return (
    <Dialog header="Создать новый вишлист" visible={visible} style={{ width: '550px' }} onHide={onHide} footer={footer}>
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="title" className="block mb-2">Название вишлиста *</label>
          <InputText id="title" value={formData.title} onChange={(e) => setFormData({ ...formData, title: e.target.value })} required className="w-full" placeholder="Мой вишлист" />
        </div>

        <div className="field mt-4">
          <label htmlFor="eventName" className="block mb-2">Название события</label>
          <InputText id="eventName" value={formData.eventName} onChange={(e) => setFormData({ ...formData, eventName: e.target.value })} className="w-full" placeholder="День рождения, Новый год..." />
        </div>

        <div className="field mt-4">
          <label htmlFor="eventDate" className="block mb-2">Дата события</label>
          <Calendar id="eventDate" value={formData.eventDate} onChange={(e) => setFormData({ ...formData, eventDate: e.value })} className="w-full" placeholder="Выберите дату" dateFormat="dd.mm.yy" showIcon />
        </div>

        <div className="field mt-4">
          <label htmlFor="description" className="block mb-2">Описание</label>
          <InputTextarea id="description" value={formData.description} onChange={(e) => setFormData({ ...formData, description: e.target.value })} rows={3} className="w-full" placeholder="Описание вишлиста..." />
        </div>

        <div className="field mt-4">
          <label htmlFor="privacyLevel" className="block mb-2">Приватность</label>
          <Dropdown id="privacyLevel" value={formData.privacyLevel} options={privacyOptions} onChange={(e) => setFormData({ ...formData, privacyLevel: e.value })} className="w-full" />
        </div>

        <div className="field mt-4">
          <label className="block mb-2">Изображение вишлиста</label>
          <FileUpload mode="basic" name="file" accept="image/*" maxFileSize={5000000} onSelect={handleImageUpload} chooseLabel="Выбрать изображение" className="w-full" />
          {previewImage && (
            <div className="mt-2">
              <img src={previewImage} alt="Превью" style={{ maxWidth: '200px', maxHeight: '200px', objectFit: 'cover' }} />
            </div>
          )}
        </div>
      </form>
    </Dialog>
  );
}

export default WishlistDialog;
