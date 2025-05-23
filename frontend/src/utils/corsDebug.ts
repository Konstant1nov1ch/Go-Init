/**
 * Утилита для проверки соединения
 * Позволяет протестировать соединение с бэкендом без Apollo Client
 */

import config from '../config';
import { formatRelative } from 'date-fns';
import { ru } from 'date-fns/locale';

/**
 * Простой вспомогательный метод для получения URL бэкенда
 */
function getHostBackendUrl(): string {
  // Определяем, запущены ли мы в Docker
  const inDocker = typeof window !== 'undefined' 
    ? window.location.hostname !== 'localhost'
    : false;
  
  // Выбираем хост на основе среды
  const host = inDocker ? 'host.docker.internal' : 'localhost';
  
  // Порт бэкенда
  const port = config.backendPort || 60013;
  
  // Формируем полный URL
  return `http://${host}:${port}`;
}

/**
 * Логирует информацию о текущей среде выполнения
 */
function logEnvironmentInfo(): void {
  if (typeof window !== 'undefined') {
    console.log('Environment Info:');
    console.log(`- Current hostname: ${window.location.hostname}`);
    console.log(`- Backend host: ${window.location.hostname !== 'localhost' ? 'host.docker.internal' : 'localhost'}`);
    console.log(`- Backend URL: ${getHostBackendUrl()}`);
    console.log(`- Running in Docker: ${window.location.hostname !== 'localhost'}`);
  }
}

/**
 * Проверка доступности бэкенда через простой fetch-запрос
 * @returns Promise с результатом проверки
 */
export async function testCorsConnection(): Promise<{
  success: boolean;
  message: string;
  details?: any;
}> {
  try {
    // Логируем информацию о среде выполнения
    logEnvironmentInfo();
    
    // Формируем URL для тестового запроса
    const url = config.apiUrl;
    console.log(`Тестирование соединения с: ${url}`);
    console.log(`Прямой URL бэкенда: ${getHostBackendUrl()}`);

    // Простой запрос для проверки подключения
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify({
        query: `{ __typename }`,
      }),
    });

    // Проверяем ответ
    if (response.ok) {
      const data = await response.json();
      return {
        success: true,
        message: 'Соединение с бэкендом установлено успешно',
        details: data,
      };
    } else {
      return {
        success: false,
        message: `Ошибка соединения: ${response.status} ${response.statusText}`,
        details: {
          status: response.status,
          statusText: response.statusText,
          headers: Object.fromEntries([...response.headers.entries()]),
        },
      };
    }
  } catch (error) {
    console.error('Ошибка при тестировании соединения:', error);
    
    // Проверяем, является ли ошибка CORS-ошибкой
    const isCorsError = error instanceof Error && 
      (error.message.includes('CORS') || 
       error.message.includes('Cross-Origin') ||
       error.message.includes('blocked'));
    
    return {
      success: false,
      message: isCorsError 
        ? `Обнаружена CORS-ошибка! ${config.inDocker ? 'Проверьте, что Docker правильно настроен для доступа к host.docker.internal' : 'Используйте Vite прокси (/graphql вместо http://localhost:60013/graphql)'}`
        : `Ошибка сети: ${error instanceof Error ? error.message : String(error)}`,
      details: error,
    };
  }
}

/**
 * Запускает тест соединения и выводит результат в консоль
 */
export async function debugCorsConnection(): Promise<void> {
  const now = new Date();
  const timestamp = formatRelative(now, new Date(), { locale: ru });
  
  console.group(`🔍 Диагностика соединения с API (${timestamp})`);
  console.log('Конфигурация:');
  console.log('- API URL:', config.apiUrl);
  console.log('- Бэкенд URL:', getHostBackendUrl());
  console.log('- В Docker:', config.inDocker);
  console.log('- Режим разработки:', import.meta.env.DEV ? 'Да' : 'Нет');
  
  try {
    const result = await testCorsConnection();
    if (result.success) {
      console.log('✅ Тест успешен:', result.message);
    } else {
      console.error('❌ Тест не пройден:', result.message);
      
      if (config.inDocker) {
        console.log('Советы для Docker:');
        console.log('1. Убедитесь, что host.docker.internal корректно разрешается в контейнере.');
        console.log('2. В docker-compose.yml должна быть настройка extra_hosts: ["host.docker.internal:host-gateway"]');
        console.log('3. Для Mac и Windows это обычно работает автоматически, для Linux нужны дополнительные настройки.');
      } else {
        console.log('Совет: В режиме разработки всегда используйте Vite прокси, который автоматически решает CORS-проблемы');
      }
    }
    console.log('Детали:', result.details);
  } catch (error) {
    console.error('❌ Критическая ошибка при тестировании:', error);
  } finally {
    console.groupEnd();
  }
}

// Автоматически запускаем диагностику в режиме разработки, если включен флаг
if (import.meta.env.DEV && (import.meta.env.VITE_ENABLE_CORS_DEBUG === 'true')) {
  window.addEventListener('load', () => {
    setTimeout(() => {
      debugCorsConnection();
    }, 1000);
  });
} 