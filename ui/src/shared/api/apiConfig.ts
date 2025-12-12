import axios from 'axios';
import { Configuration } from '@generated/configuration';
import { AlbumsApi, MediaApi, StatsApi } from '@generated/api';
import { handleUnauthorized } from '@shared/auth';

// Create axios instance with interceptors
const axiosInstance = axios.create();

axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 || error.response?.status === 403) {
      // Trigger logout flow: Keycloak -> Envoy /signout -> App home
      // Uses client_id since id_token_hint is not accessible (HttpOnly cookie)
      handleUnauthorized();
    }
    return Promise.reject(error);
  }
);

// API Configuration
const apiConfig = new Configuration({
  basePath: process.env.REACT_APP_API_URL || '/api/v1',
});

// API Instances
export const albumsApi = new AlbumsApi(apiConfig, undefined, axiosInstance);
export const mediaApi = new MediaApi(apiConfig, undefined, axiosInstance);
export const statsApi = new StatsApi(apiConfig, undefined, axiosInstance);

export { apiConfig };
