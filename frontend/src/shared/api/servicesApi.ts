import type { TransportService, TransportServiceFilters, TransportServiceRaw } from '../types/TransportService';
import { getApiUrl, getServerIPAddress, isTauriApp } from '../config/apiConfig';

/**
 * Конвертирует сырой формат Service из API (snake_case) в camelCase
 * Также преобразует полный URL MinIO в относительный путь для прокси или IP для Tauri
 */
function convertService(raw: TransportServiceRaw): TransportService {
  // Преобразуем полный URL в относительный путь для прокси или IP для Tauri
  let imageUrl = raw.image_url;
  if (imageUrl) {
    if (imageUrl.startsWith('http://localhost:9003/')) {
      // Извлекаем путь из URL (например, /lab1/fura.jpg)
      const urlObj = new URL(imageUrl);
      const imagePath = urlObj.pathname; // /lab1/fura.jpg
      
      if (isTauriApp()) {
        // Для Tauri используем полный URL с IP сервера и путем /lab1/...
        const serverIP = getServerIPAddress();
        if (serverIP) {
          // Убираем trailing slash если есть
          const baseUrl = serverIP.replace(/\/$/, '');
          imageUrl = `${baseUrl}${imagePath}`;
          console.log(`[Tauri] Converted image URL: ${raw.image_url} -> ${imageUrl}`);
        } else {
          // Fallback: используем относительный путь
          imageUrl = imagePath;
          console.warn('[Tauri] Server IP not configured, using relative path');
        }
      } else {
        // Для веб-версии используем относительный путь через прокси бэкенда
        // Бэкенд проксирует /lab1/* к MinIO
        // В DEV нельзя префиксить BASE_URL (иначе получаем /RIP-2-mod-/lab1/... -> 404)
        // В PROD (GitHub Pages) наоборот нужен BASE_URL.
        if (import.meta.env.DEV) {
          imageUrl = imagePath; // например: /lab1/fura.jpg
        } else {
          const baseUrl = import.meta.env.BASE_URL || '/';
          imageUrl = `${baseUrl}${imagePath.replace(/^\//, '')}`;
        }
        console.log(`[Web] Converted image URL: ${raw.image_url} -> ${imageUrl}`);
      }
    }
  }
  
  return {
    id: raw.id,
    name: raw.name,
    description: raw.description,
    price: raw.price,
    imageUrl: imageUrl,
    deliveryDays: raw.delivery_days,
    maxWeight: raw.max_weight,
    maxVolume: raw.max_volume,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at
  };
}

/**
 * Mock данные для fallback когда бэкенд недоступен
 * Имитируют реальные данные услуг по транспортировке грузов
 */
const MOCK_SERVICES: TransportService[] = [
  {
    id: 1,
    name: 'Фура',
    description: 'Грузоперевозки на фурах для больших объемов. Идеально для габаритных грузов.',
    price: 50000,
    deliveryDays: 7,
    maxWeight: 20000,
    maxVolume: 80,
    imageUrl: undefined, // Используется default.svg
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z'
  },
  {
    id: 2,
    name: 'Малотоннажный',
    description: 'Быстрые грузоперевозки на малотоннажных автомобилях. Подходит для малых партий.',
    price: 15000,
    deliveryDays: 3,
    maxWeight: 3000,
    maxVolume: 15,
    imageUrl: undefined,
    createdAt: '2024-01-10T10:00:00Z',
    updatedAt: '2024-01-10T10:00:00Z'
  },
  {
    id: 3,
    name: 'Авиа',
    description: 'Скоростная доставка грузов по воздуху. Самый быстрый способ доставки.',
    price: 150000,
    deliveryDays: 1,
    maxWeight: 5000,
    maxVolume: 25,
    imageUrl: undefined,
    createdAt: '2024-01-20T10:00:00Z',
    updatedAt: '2024-01-20T10:00:00Z'
  },
  {
    id: 4,
    name: 'Поезд',
    description: 'Надежные железнодорожные перевозки. Оптимально для больших объемов на дальние расстояния.',
    price: 80000,
    deliveryDays: 14,
    maxWeight: 40000,
    maxVolume: 100,
    imageUrl: undefined,
    createdAt: '2024-01-05T10:00:00Z',
    updatedAt: '2024-01-05T10:00:00Z'
  },
  {
    id: 5,
    name: 'Корабль',
    description: 'Морские грузоперевозки. Экономичный вариант для международной доставки.',
    price: 120000,
    deliveryDays: 30,
    maxWeight: 100000,
    maxVolume: 500,
    imageUrl: undefined,
    createdAt: '2024-01-12T10:00:00Z',
    updatedAt: '2024-01-12T10:00:00Z'
  },
  {
    id: 6,
    name: 'Мультимодальный',
    description: 'Комбинированная доставка разными видами транспорта. Максимальная гибкость.',
    price: 100000,
    deliveryDays: 10,
    maxWeight: 30000,
    maxVolume: 150,
    imageUrl: undefined,
    createdAt: '2024-01-18T10:00:00Z',
    updatedAt: '2024-01-18T10:00:00Z'
  }
];

/**
 * Конвертирует объект фильтров в query string для URL
 * @param filters - параметры фильтрации
 * @returns query string (например: "?search=truck&minPrice=1000")
 */
function buildQueryString(filters: TransportServiceFilters): string {
  const params = new URLSearchParams();
  
  if (filters.search) params.append('search', filters.search);
  if (filters.minPrice !== undefined) params.append('minPrice', filters.minPrice.toString());
  if (filters.maxPrice !== undefined) params.append('maxPrice', filters.maxPrice.toString());
  if (filters.dateFrom) params.append('dateFrom', filters.dateFrom);
  if (filters.dateTo) params.append('dateTo', filters.dateTo);
  
  const queryString = params.toString();
  return queryString ? `?${queryString}` : '';
}

/**
 * Имитирует серверную фильтрацию на mock данных
 * Используется когда бэкенд недоступен
 */
function filterMockServices(services: TransportService[], filters: TransportServiceFilters): TransportService[] {
  return services.filter(service => {
    // Поиск по названию и описанию
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      const matchesSearch = 
        service.name.toLowerCase().includes(searchLower) ||
        service.description.toLowerCase().includes(searchLower);
      if (!matchesSearch) return false;
    }
    
    // Фильтр по цене
    if (filters.minPrice !== undefined && service.price < filters.minPrice) return false;
    if (filters.maxPrice !== undefined && service.price > filters.maxPrice) return false;
    
    // Фильтр по дате создания
    // dateFrom и dateTo приходят в формате YYYY-MM-DD, сравниваем как строки
    // createdAt в формате ISO (2024-01-15T10:00:00Z), берем только дату
    if (filters.dateFrom) {
      const serviceDate = service.createdAt.split('T')[0]; // YYYY-MM-DD
      if (serviceDate < filters.dateFrom) return false;
    }
    if (filters.dateTo) {
      const serviceDate = service.createdAt.split('T')[0]; // YYYY-MM-DD
      if (serviceDate > filters.dateTo) return false;
    }
    
    return true;
  });
}

/**
 * Получение списка услуг с сервера или mock данных
 * 
 * @param filters - параметры фильтрации (опционально)
 * @returns Promise с массивом услуг
 * 
 * Логика работы:
 * 1. Делает fetch запрос к бэкенду через proxy (/api/services)
 * 2. Если сервер недоступен → возвращает mock данные
 * 3. Если сервер доступен → серверная фильтрация, иначе локальная на mock
 */
export async function fetchTransportServices(filters: TransportServiceFilters = {}): Promise<TransportService[]> {
  try {
    const queryString = buildQueryString(filters);
    const response = await fetch(getApiUrl(`/api/transport-services${queryString}`));
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    
    // Бэкенд может вернуть либо transport_services (доменно), либо services (старый формат)
    let services: TransportServiceRaw[] = [];
    if (data.transport_services && Array.isArray(data.transport_services)) {
      services = data.transport_services;
    } else if (data.services && Array.isArray(data.services)) {
      services = data.services;
    } else if (Array.isArray(data)) {
      services = data;
    } else {
      throw new Error('Unexpected response format');
    }
    
    // Конвертируем из snake_case в camelCase
    return services.map(convertService);
  } catch (error) {
    console.warn('Failed to fetch services from backend, using mock data:', error);
    
    // Fallback на mock данные при недоступности сервера
    // Применяем локальную фильтрацию если сервер не ответил
    return filterMockServices(MOCK_SERVICES, filters);
  }
}

/**
 * Получение одной услуги по ID
 */
export async function fetchTransportService(id: number): Promise<TransportService | null> {
  try {
    const response = await fetch(getApiUrl(`/api/transport-services/${id}`));
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    
    // Если ответ содержит поле service
    if (data.service) {
      return convertService(data.service);
    }
    
    return null;
  } catch (error) {
    console.warn(`Failed to fetch service ${id} from backend:`, error);
    
    // Fallback на mock данные
    return MOCK_SERVICES.find(s => s.id === id) || null;
  }
}

