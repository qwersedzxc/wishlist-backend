import React, { useState } from 'react';
import { Dialog } from 'primereact/dialog';
import { InputText } from 'primereact/inputtext';
import { InputTextarea } from 'primereact/inputtextarea';
import { Calendar } from 'primereact/calendar';
import { Dropdown } from 'primereact/dropdown';
import { Checkbox } from 'primereact/checkbox';
import { Button } from 'primereact/button';
import { Chips } from 'primereact/chips';

function EnhancedWishlistDialog({ visible, onHide, onSubmit }) {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    eventDate: null,
    category: null,
    tags: [],
    isPublic: false,
    privacyLevel: 'private' // private, friends, link
  });

  const categories = [
    { label: 'День рождения', value: 'birthday' },
    { label: 'Новый год', value: 'new_year' },
    { label: 'Свадьба', value: 'wedding' },
    { label: 'Годовщина', value: 'anniversary' },
    { label: 'Выпускной', value: 'graduation' },
    { label: 'Другое', value: 'other' }
  ];

  const privacyOptions = [
    { label: 'Только я', value: 'private', icon: 'pi pi-lock' },
    { label: 'Друзья', value: 'friends', icon: 'pi pi-users' },
    { label: 'По ссылке', value: 'link', icon: 'pi pi-link' },
    { label: 'Публичный', value: 'public', icon: 'pi pi-globe' }
  ];

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit({
      ...formData,
      isPublic: formData.privacyLevel === 'public'
    });
    setFormData({
      title: '',
      description: '',
      eventDate: null,
      category: null,
      tags: [],
      isPublic: false,
      privacyLevel: 'private'
    });
  };

  const footer = (
    <div>
      <Button label="Отмена" icon="pi pi-times" onClick={onHide} className="p-button-text" />
      <Button label="Создать" icon="pi pi-check" onClick={handleSubmit} autoFocus />
    </div>
  );

  return (
    <Dialog
      header="Создать новый вишлист"
      visible={visible}
      style={{ width: '600px' }}
      onHide={onHide}
      footer={footer}
    >
      <form onSubmit={handleSubmit} className="p-fluid">
        <div className="field">
          <label htmlFor="title" className="block mb-2">
            Название события/праздника *
          </label>
          <InputText
            id="title"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            required
            placeholder="Мой день рождения 2024"
          />
        </div>

        <div className="grid">
          <div className="col-12 md:col-6">
            <div className="field">
              <label htmlFor="eventDate" className="block mb-2">
                Дата события
              </label>
              <Calendar
                id="eventDate"
                value={formData.eventDate}
                onChange={(e) => setFormData({ ...formData, eventDate: e.value })}
                dateFormat="dd.mm.yy"
                placeholder="Выберите дату"
                showIcon
              />
            </div>
          </div>

          <div className="col-12 md:col-6">
            <div className="field">
              <label htmlFor="category" className="block mb-2">
                Категория события
              </label>
              <Dropdown
                id="category"
                value={formData.category}
                options={categories}
                onChange={(e) => setFormData({ ...formData, category: e.value })}
                placeholder="Выберите категорию"
              />
            </div>
          </div>
        </div>

        <div className="field">
          <label htmlFor="description" className="block mb-2">
            Описание
          </label>
          <InputTextarea
            id="description"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            rows={3}
            placeholder="Расскажите о своем событии..."
          />
        </div>

        <div className="field">
          <label htmlFor="tags" className="block mb-2">
            Теги
          </label>
          <Chips
            id="tags"
            value={formData.tags}
            onChange={(e) => setFormData({ ...formData, tags: e.value })}
            placeholder="Добавьте теги (Enter для добавления)"
          />
          <small className="text-color-secondary">
            Например: подарки, электроника, книги
          </small>
        </div>

        <div className="field">
          <label className="block mb-3">Приватность</label>
          <div className="grid">
            {privacyOptions.map((option) => (
              <div key={option.value} className="col-6">
                <div className="field-radiobutton">
                  <input
                    type="radio"
                    id={option.value}
                    name="privacy"
                    value={option.value}
                    checked={formData.privacyLevel === option.value}
                    onChange={(e) => setFormData({ ...formData, privacyLevel: e.target.value })}
                  />
                  <label htmlFor={option.value} className="ml-2">
                    <i className={`${option.icon} mr-2`}></i>
                    {option.label}
                  </label>
                </div>
              </div>
            ))}
          </div>
        </div>

        {formData.privacyLevel === 'link' && (
          <div className="p-3 surface-100 border-round">
            <i className="pi pi-info-circle mr-2"></i>
            <span className="text-sm">
              Вишлист будет доступен только тем, у кого есть прямая ссылка
            </span>
          </div>
        )}
      </form>
    </Dialog>
  );
}

export default EnhancedWishlistDialog;