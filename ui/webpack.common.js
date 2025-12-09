/* eslint-disable @typescript-eslint/no-var-requires */

const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const Dotenv = require('dotenv-webpack');
const webpack = require('webpack');
const { execSync } = require('child_process');

const ASSET_PATH = process.env.ASSET_PATH || '/';

// Get git commit hash from environment variable or git command (fallback for local dev)
const getGitCommit = () => {
  // First try environment variable (used in container builds)
  if (process.env.GIT_SHA) {
    return process.env.GIT_SHA;
  }
  
  // Fallback to git command for local development
  try {
    return execSync('git rev-parse --short HEAD', { encoding: 'utf8' }).trim();
  } catch (e) {
    return 'unknown';
  }
};

module.exports = (env) => {
  // AUTH_ENABLED: false in development, true in production (can be overridden via env var)
  // Only set if not already defined in environment (to avoid conflict with Dotenv plugin)
  const authEnabled = process.env.AUTH_ENABLED !== undefined
    ? process.env.AUTH_ENABLED
    : (env === 'production' ? 'true' : 'false');

  // AUTHZ_ENABLED: false in development, true in production (can be overridden via env var)
  const authzEnabled = process.env.AUTHZ_ENABLED !== undefined
    ? process.env.AUTHZ_ENABLED
    : (env === 'production' ? 'true' : 'false');

  // Set environment variables before Dotenv plugin runs
  process.env.AUTH_ENABLED = authEnabled;
  process.env.AUTHZ_ENABLED = authzEnabled;

  return {
    entry: './src/index.tsx',
    module: {
      rules: [
        {
          test: /\.(tsx|ts|jsx)?$/,
          use: [
            {
              loader: 'ts-loader',
              options: {
                transpileOnly: true,
                experimentalWatchApi: true,
              },
            },
          ],
        },
        {
          test: /\.(svg|ttf|eot|woff|woff2)$/,
          type: 'asset/resource',
        },
        {
          test: /\.(jpg|jpeg|png|gif)$/i,
          type: 'asset/inline',
          parser: {
            dataUrlCondition: {
              maxSize: 5000,
            },
          },
        },
      ],
    },
    output: {
      filename: '[name].bundle.js',
      path: path.resolve(__dirname, 'dist'),
      publicPath: ASSET_PATH,
      clean: true,
    },
    plugins: [
      new HtmlWebpackPlugin({
        template: path.resolve(__dirname, 'src', 'index.html'),
      }),
      new Dotenv({
        systemvars: true,
        silent: true,
      }),
      new webpack.DefinePlugin({
        'process.env.GIT_COMMIT': JSON.stringify(getGitCommit()),
      }),
    ],
    resolve: {
      extensions: ['.js', '.ts', '.tsx', '.jsx'],
      plugins: [
        new TsconfigPathsPlugin({
          configFile: path.resolve(__dirname, './tsconfig.json'),
        }),
      ],
      symlinks: false,
      cacheWithContext: false,
    },
  };
};