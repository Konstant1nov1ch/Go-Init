/**
 * Утилита для определения правильного URL бэкенда в зависимости от среды выполнения
 * Учитывает особенности Docker, где для доступа к хост-машине используется host.docker.internal
 */

/**
 * Возвращает полный URL для доступа к бэкенду с учетом среды выполнения
 * @returns URL бэкенда
 */
export function getHostBackendUrl(): string {
  // Определяем, запущены ли мы в Docker
  const inDocker = typeof window !== 'undefined' 
    ? window.location.hostname !== 'localhost'
    : process.env.DOCKER_ENV === 'true';
  
  // Выбираем хост на основе среды
  const host = inDocker ? 'host.docker.internal' : 'localhost';
  
  // Порт бэкенда
  const port = 60013;
  
  // Формируем полный URL
  return `http://${host}:${port}`;
}

/**
 * Возвращает только хост для доступа к бэкенду
 * @returns Хост бэкенда (без протокола и порта)
 */
export function getBackendHost(): string {
  return typeof window !== 'undefined' && window.location.hostname !== 'localhost'
    ? 'host.docker.internal'
    : 'localhost';
}

/**
 * Логирует информацию о текущей среде выполнения
 */
export function logEnvironmentInfo(): void {
  if (typeof window !== 'undefined') {
    console.log('Environment Info:');
    console.log(`- Current hostname: ${window.location.hostname}`);
    console.log(`- Backend host: ${getBackendHost()}`);
    console.log(`- Backend URL: ${getHostBackendUrl()}`);
    console.log(`- Running in Docker: ${window.location.hostname !== 'localhost'}`);
  }
} 