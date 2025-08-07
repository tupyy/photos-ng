import { Configuration } from '@generated/configuration';
import { AlbumsApi, MediaApi, StatsApi } from '@generated/api';

// API Configuration
const apiConfig = new Configuration({
  basePath: process.env.REACT_APP_API_URL || '/api/v1',
});

// API Instances
export const albumsApi = new AlbumsApi(apiConfig);
export const mediaApi = new MediaApi(apiConfig);
export const statsApi = new StatsApi(apiConfig);

export { apiConfig };
