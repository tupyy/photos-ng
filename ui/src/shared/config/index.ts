export const TIMELINE_PAGE_SIZE = 100;

// Authentication mode - when false, user fetch is skipped and profile menu is hidden
// Set via webpack: false in dev, true in prod (can override with AUTH_ENABLED env var)
export const AUTH_ENABLED = process.env.AUTH_ENABLED === 'true';

// Authorization mode - when false, all permissions are granted
// Set via webpack: false in dev, true in prod (can override with AUTHZ_ENABLED env var)
export const AUTHZ_ENABLED = process.env.AUTHZ_ENABLED === 'true';
