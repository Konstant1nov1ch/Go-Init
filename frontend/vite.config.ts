import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Helper function for host resolution
function getHostBackendUrl() {
  const isDocker = process.env.DOCKER_ENV === 'true';
  const host = isDocker ? 'go_init_manager' : 'localhost';
  const port = 60013; // <--- ВАЖНО
  return `http://${host}:${port}`;
}



// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: false, // Allow Vite to find an available port automatically
    cors: true,
    proxy: {
      '/api': {
        target: getHostBackendUrl(),
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: (proxy, _options) => {
          // eslint-disable-next-line no-console
          proxy.on('error', (err, _req, _res) => {
            console.log('proxy error', err);
          });
          // eslint-disable-next-line no-console
          proxy.on('proxyReq', (_proxyReq, req, _res) => {
            console.log('Sending Request to the Target:', req.method, req.url);
          });
          // eslint-disable-next-line no-console
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
          });
        }
      },
      // Add direct proxy for GraphQL endpoint
      '/graphql': {
        target: getHostBackendUrl(),
        changeOrigin: true,
        secure: false,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('GraphQL proxy error:', err);
          });
        }
      }
    }
  },
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      '@apollo/client',
      'graphql'
    ],
    force: true
  }
})
