import { useMemo, useState } from 'react';
import type { TransportService } from '../../shared/types/TransportService';

/**
 * Props для компонента ServiceCard
 */
interface ServiceCardProps {
  /** Данные услуги для отображения */
  service: TransportService;
}

/**
 * Компонент карточки услуги
 * 
 * Отображает информацию об услуге в виде карточки с существующими стилями
 * Если изображения нет, подставляет изображение по-умолчанию
 * 
 * @param props - содержит объект service с данными услуги
 */
export function ServiceCard({ service }: ServiceCardProps) {
  // Получаем base URL из Vite (для GitHub Pages это /RIP-2-mod-/)
  const baseUrl = import.meta.env.BASE_URL || '/';
  // URL изображения по умолчанию если поле пустое
  const defaultImageUrl = `${baseUrl}assets/default.svg`;
  const imageUrl = service.imageUrl || defaultImageUrl;

  const [isAdding, setIsAdding] = useState(false);
  const addButtonLabel = useMemo(() => {
    if (isAdding) return 'Добавляем...';
    return 'Добавить в заявку';
  }, [isAdding]);

  const handleAddToDraft = async () => {
    if (isAdding) return;
    setIsAdding(true);
    try {
      const res = await fetch(`/api/logistic-requests/draft/services/${service.id}`, { method: 'POST' });
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      // Сообщаем шапке/виджету, что черновик обновился
      window.dispatchEvent(new Event('draft-logistic-request-updated'));
    } catch (e) {
      console.error('Failed to add transport service to draft logistic request:', e);
      alert('Не удалось добавить услугу в черновик заявки');
    } finally {
      setIsAdding(false);
    }
  };
  
  return (
    <div className="service-card">
      <img 
        src={imageUrl} 
        alt={service.name}
        className="service-image"
        onError={(e) => {
          // Fallback если изображение не загрузилось
          (e.target as HTMLImageElement).src = defaultImageUrl;
        }}
      />
      <div className="service-content">
        <h3 className="service-title">{service.name}</h3>
        <p className="service-description">{service.description}</p>
        <div className="service-actions">
          <a href={`#service-${service.id}`} className="details-link">подробнее</a>
          <button
            type="button"
            className="consult-btn"
            data-service-id={service.id}
            onClick={handleAddToDraft}
            disabled={isAdding}
            aria-disabled={isAdding}
          >
            {addButtonLabel}
          </button>
        </div>
      </div>
    </div>
  );
}