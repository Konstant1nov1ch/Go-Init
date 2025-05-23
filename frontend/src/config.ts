const config = {
  apiUrl: '/graphql', // <--- Лови прокси внутри Vite
  inDocker: !['localhost', '127.0.0.1'].includes(window.location.hostname),

  backendHost: !['localhost', '127.0.0.1'].includes(window.location.hostname)
    ? 'go_init_manager'
    : 'localhost',

  backendPort: 8080,

  version: '1.0.0',

  auth: {
    enabled: false,
    tokenKey: 'auth_token',
  },

  debug: {
    logRequests: !import.meta.env.PROD,
  },
};

export default config;
