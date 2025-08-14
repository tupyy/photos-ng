/* eslint-disable @typescript-eslint/no-var-requires */

const path = require('path');
const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');
const { stylePaths } = require('./stylePaths');

const HOST = process.env.HOST || 'localhost';
const PORT = process.env.PORT || '3000';

// Proxy configuration
const PROXY_MODE = process.env.PROXY_MODE || 'local'; // 'local' or 'remote'
const getProxyTarget = () => {
  switch (PROXY_MODE) {
    case 'remote':
      return 'https://photos.tls.tupangiu.ro';
    case 'local':
    default:
      return 'http://localhost:8080';
  }
};

const proxyTarget = getProxyTarget();
console.log(`🔗 Proxy mode: ${PROXY_MODE} -> ${proxyTarget}`);

module.exports = merge(common('development'), {
  mode: 'development',
  devtool: 'eval-source-map',
  devServer: {
    host: HOST,
    port: PORT,
    historyApiFallback: true,
    open: true,
    static: {
      directory: path.resolve(__dirname, 'dist'),
    },
    client: {
      overlay: true,
    },
    proxy: [
      {
        context: ['/api'],
        target: proxyTarget,
        changeOrigin: true,
        secure: PROXY_MODE === 'remote', // Enable SSL verification for remote
        logLevel: 'debug',
        onProxyReq: (proxyReq, req, res) => {
          console.log(`📡 Proxying ${req.method} ${req.url} -> ${proxyTarget}${req.url}`);
        },
        onError: (err, req, res) => {
          console.error(`❌ Proxy error for ${req.url}:`, err.message);
        }
      }
    ],
  },
  module: {
    rules: [
      {
        test: /\.css$/,
        include: [...stylePaths],
        use: ['style-loader', 'css-loader'],
      },
    ],
  },
});