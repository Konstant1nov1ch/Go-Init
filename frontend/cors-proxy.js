// Simple CORS proxy server for development
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');

const app = express();
const PORT = 8080;

// Enable CORS for all routes
app.use(cors({
  origin: '*', // Allow all origins
  methods: ['GET', 'POST', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With'],
  credentials: true
}));

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} [${req.method}] ${req.url}`);
  next();
});

// Proxy all requests to the target backend server
const apiProxy = createProxyMiddleware({
  target: 'http://localhost:60013',
  changeOrigin: true,
  pathRewrite: {
    '^/api': '' // Remove /api prefix when forwarding
  },
  onProxyRes: (proxyRes, req, res) => {
    // Add CORS headers to the response
    proxyRes.headers['Access-Control-Allow-Origin'] = '*';
    proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, OPTIONS';
    proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization';
    
    // Log the response status
    console.log(`${new Date().toISOString()} [${req.method}] ${req.url} - ${proxyRes.statusCode}`);
  },
  onError: (err, req, res) => {
    console.error('Proxy Error:', err);
    res.status(500).send('Proxy Error');
  }
});

// Apply the proxy middleware to all routes
app.use('/', apiProxy);

// Start the server
app.listen(PORT, () => {
  console.log(`CORS Proxy Server running at http://localhost:${PORT}`);
  console.log(`Proxying requests to http://localhost:60013`);
  console.log('Use this URL in your config.ts file:');
  console.log(`apiUrl: 'http://localhost:${PORT}/graphql'`);
}); 