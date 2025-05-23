import React, { useState } from 'react';
import config from '../config';

// Стили для компонента
const styles = {
  container: {
    padding: '20px',
    border: '1px solid #eee',
    borderRadius: '8px',
    backgroundColor: '#f9f9f9',
    maxWidth: '800px',
    margin: '0 auto'
  },
  header: {
    fontSize: '1.2rem',
    fontWeight: 'bold',
    marginBottom: '15px'
  },
  buttonRow: {
    display: 'flex',
    gap: '10px',
    marginBottom: '20px'
  },
  button: {
    padding: '8px 15px',
    borderRadius: '4px',
    cursor: 'pointer',
    fontWeight: 'bold',
    border: 'none'
  },
  primaryButton: {
    backgroundColor: '#4361ee',
    color: 'white'
  },
  secondaryButton: {
    backgroundColor: '#eee',
    border: '1px solid #ddd',
    color: '#333'
  },
  result: {
    padding: '15px',
    borderRadius: '4px',
    border: '1px solid #eee',
    marginTop: '15px',
    maxHeight: '300px',
    overflow: 'auto',
    backgroundColor: '#fff'
  },
  success: {
    backgroundColor: 'rgba(39, 174, 96, 0.1)',
    borderColor: '#27ae60'
  },
  error: {
    backgroundColor: 'rgba(231, 76, 60, 0.1)',
    borderColor: '#e74c3c'
  },
  pre: {
    whiteSpace: 'pre-wrap',
    margin: 0,
    fontFamily: 'monospace',
    fontSize: '13px'
  },
  url: {
    fontFamily: 'monospace',
    padding: '8px',
    backgroundColor: '#eee',
    borderRadius: '4px',
    fontSize: '13px',
    marginBottom: '15px',
    wordBreak: 'break-all' as const
  },
  label: {
    fontWeight: 'bold',
    marginBottom: '5px',
    display: 'block'
  }
};

interface TestResult {
  success: boolean;
  data?: any;
  error?: string;
  headers?: Record<string, string>;
  requestDuration?: number;
}

/**
 * Компонент для тестирования API соединения
 */
const ApiTester: React.FC = () => {
  const [result, setResult] = useState<TestResult | null>(null);
  const [loading, setLoading] = useState(false);
  
  const testWithFetch = async () => {
    setLoading(true);
    setResult(null);
    
    const startTime = performance.now();
    
    try {
      // Простой тестовый запрос к GraphQL API
      const response = await fetch(config.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({
          query: `{ __typename }`
        })
      });
      
      const duration = performance.now() - startTime;
      
      // Получаем заголовки для анализа
      const headers: Record<string, string> = {};
      response.headers.forEach((value, key) => {
        headers[key] = value;
      });
      
      if (response.ok) {
        const data = await response.json();
        setResult({
          success: true,
          data,
          headers,
          requestDuration: duration
        });
      } else {
        setResult({
          success: false,
          error: `Ошибка: ${response.status} ${response.statusText}`,
          headers,
          requestDuration: duration
        });
      }
    } catch (error) {
      const duration = performance.now() - startTime;
      
      setResult({
        success: false,
        error: error instanceof Error ? error.message : String(error),
        requestDuration: duration
      });
    } finally {
      setLoading(false);
    }
  };
  
  const testWithApollo = async () => {
    setLoading(true);
    setResult(null);
    
    try {
      // Импортируем Apollo клиент динамически
      const { client } = await import('../graphql/client');
      const { gql } = await import('@apollo/client');
      
      const startTime = performance.now();
      
      const response = await client.query({
        query: gql`
          query TestQuery {
            __typename
          }
        `,
        fetchPolicy: 'network-only'
      });
      
      const duration = performance.now() - startTime;
      
      setResult({
        success: true,
        data: response.data,
        requestDuration: duration
      });
    } catch (error) {
      setResult({
        success: false,
        error: error instanceof Error ? error.message : String(error)
      });
    } finally {
      setLoading(false);
    }
  };
  
  const toggleCorsMode = () => {
    // Проверяем, есть ли параметр CORS в URL
    const url = new URL(window.location.href);
    const hasCors = url.searchParams.has('cors');
    
    // Переключаем режим CORS
    if (hasCors) {
      url.searchParams.delete('cors');
    } else {
      url.searchParams.set('cors', '1');
    }
    
    // Перезагружаем страницу с новыми параметрами
    window.location.href = url.toString();
  };
  
  return (
    <div style={styles.container}>
      <div style={styles.header}>Тест API соединения</div>
      
      <div style={styles.url}>
        <div style={styles.label}>Текущий URL API:</div>
        {config.apiUrl}
        {config.apiUrl.includes('8080') && (
          <div style={{color: 'green', marginTop: '5px'}}>
            ✓ Используется CORS прокси
          </div>
        )}
      </div>
      
      <div style={styles.buttonRow}>
        <button 
          style={{...styles.button, ...styles.primaryButton}}
          onClick={testWithFetch}
          disabled={loading}
        >
          {loading ? 'Тестирование...' : 'Тест с Fetch API'}
        </button>
        
        <button 
          style={{...styles.button, ...styles.primaryButton}}
          onClick={testWithApollo}
          disabled={loading}
        >
          {loading ? 'Тестирование...' : 'Тест с Apollo'}
        </button>
        
        <button 
          style={{...styles.button, ...styles.secondaryButton}}
          onClick={toggleCorsMode}
        >
          {config.apiUrl.includes('8080') 
            ? 'Отключить CORS прокси' 
            : 'Включить CORS прокси'}
        </button>
      </div>
      
      {result && (
        <div style={{
          ...styles.result,
          ...(result.success ? styles.success : styles.error)
        }}>
          <div style={styles.label}>
            {result.success ? 'Успешное соединение' : 'Ошибка соединения'}
            {result.requestDuration && (
              <span style={{fontWeight: 'normal', marginLeft: '10px'}}>
                ({result.requestDuration.toFixed(0)}мс)
              </span>
            )}
          </div>
          
          {result.error && (
            <div style={{marginBottom: '10px', color: '#e74c3c'}}>
              {result.error}
            </div>
          )}
          
          {result.data && (
            <>
              <div style={styles.label}>Данные:</div>
              <pre style={styles.pre}>
                {JSON.stringify(result.data, null, 2)}
              </pre>
            </>
          )}
          
          {result.headers && (
            <>
              <div style={styles.label}>Заголовки ответа:</div>
              <pre style={styles.pre}>
                {Object.entries(result.headers)
                  .filter(([key]) => key.toLowerCase().includes('access-control') || 
                                   key.toLowerCase().includes('content') ||
                                   key.toLowerCase().includes('origin'))
                  .map(([key, value]) => `${key}: ${value}`)
                  .join('\n')}
              </pre>
              
              {!Object.keys(result.headers).some(h => 
                h.toLowerCase().includes('access-control-allow-origin')) && (
                <div style={{color: '#e74c3c', marginTop: '10px'}}>
                  ⚠️ Отсутствует заголовок Access-Control-Allow-Origin!
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default ApiTester; 