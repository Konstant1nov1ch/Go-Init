import { ApolloClient, InMemoryCache, createHttpLink, ApolloLink, from } from '@apollo/client';
import { onError } from '@apollo/client/link/error';
import config from '../config';

// Создание HTTP-линка с конфигурацией из config.ts
const httpLink = createHttpLink({
  uri: config.apiUrl,
  credentials: 'include', // Включаем передачу куки для CORS запросов
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  }
});

// Middleware для добавления токена аутентификации в заголовки
const authMiddleware = new ApolloLink((operation, forward) => {
  if (config.auth.enabled) {
    const token = localStorage.getItem(config.auth.tokenKey);
    if (token) {
      operation.setContext({
        headers: {
          authorization: `Bearer ${token}`,
        },
      });
    }
  }
  
  // Добавляем логирование запросов в режиме отладки
  if (config.debug.logRequests) {
    console.log(`[GraphQL Request] ${operation.operationName}:`, 
      operation.variables ? JSON.stringify(operation.variables) : 'No variables');
  }
  
  return forward(operation);
});

// Обработка ошибок
const errorLink = onError(({ graphQLErrors, networkError, operation }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      );
    });
  }

  if (networkError) {
    console.error(`[Network error for operation ${operation.operationName}]:`, networkError);
    
    // Проверка на сетевые ошибки
    if (networkError.message && networkError.message.includes('Failed to fetch')) {
      console.error('[API Error] Не удалось подключиться к API серверу');
      console.error(`Попытка запроса на ${config.apiUrl}`);
      console.error('Убедитесь, что:');
      console.error('1. Сервер запущен и доступен');
      console.error('2. В режиме разработки используется Vite прокси на /graphql'); 
    }
  }
});

// Создание Apollo Client
export const client = new ApolloClient({
  link: from([errorLink, authMiddleware, httpLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all',
    },
    query: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all',
    },
    mutate: {
      errorPolicy: 'all',
    },
  },
}); 