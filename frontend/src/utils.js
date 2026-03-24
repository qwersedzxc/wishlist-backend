// Возвращает полный URL изображения.
// Если URL уже абсолютный (https://...) — возвращает как есть.
// Если относительный (/uploads/...) — добавляет localhost для обратной совместимости.
export function resolveUrl(url) {
  if (!url) return '';
  if (url.startsWith('http://') || url.startsWith('https://')) return url;
  return `http://localhost:8081${url}`;
}
