/**
 * Authentication utilities for Envoy + Keycloak OIDC flow
 *
 * Handles the "logout loop" problem when HttpOnly cookies prevent
 * JavaScript from accessing id_token_hint.
 */

// Configuration from environment variables
const AUTH_CONFIG = {
  keycloakLogoutUrl: process.env.REACT_APP_KEYCLOAK_LOGOUT_URL || '',
  envoySignoutUrl: process.env.REACT_APP_ENVOY_SIGNOUT_URL || '',
  clientId: process.env.REACT_APP_OIDC_CLIENT_ID || '',
} as const;

// Session storage key to prevent redirect loops
const LOGOUT_IN_PROGRESS_KEY = 'auth_logout_in_progress';
const LOGOUT_TIMESTAMP_KEY = 'auth_logout_timestamp';
const LOGOUT_TIMEOUT_MS = 30000; // 30 seconds max for logout flow

/**
 * Check if a logout is already in progress to prevent loops
 */
function isLogoutInProgress(): boolean {
  const inProgress = sessionStorage.getItem(LOGOUT_IN_PROGRESS_KEY);
  const timestamp = sessionStorage.getItem(LOGOUT_TIMESTAMP_KEY);

  if (!inProgress || !timestamp) {
    return false;
  }

  // Check if the logout flow has timed out (stuck state)
  const elapsed = Date.now() - parseInt(timestamp, 10);
  if (elapsed > LOGOUT_TIMEOUT_MS) {
    clearLogoutState();
    return false;
  }

  return true;
}

/**
 * Mark that a logout is in progress
 */
function setLogoutInProgress(): void {
  sessionStorage.setItem(LOGOUT_IN_PROGRESS_KEY, 'true');
  sessionStorage.setItem(LOGOUT_TIMESTAMP_KEY, Date.now().toString());
}

/**
 * Clear the logout state (call this on successful app load)
 */
export function clearLogoutState(): void {
  sessionStorage.removeItem(LOGOUT_IN_PROGRESS_KEY);
  sessionStorage.removeItem(LOGOUT_TIMESTAMP_KEY);
}

/**
 * Perform a complete logout flow:
 * 1. Redirect to Keycloak logout endpoint (kills SSO session)
 * 2. Keycloak redirects to Envoy /signout (clears proxy cookies)
 * 3. Envoy redirects to app home
 *
 * Uses client_id since id_token_hint is not accessible (HttpOnly cookie)
 */
export function performLogout(): void {
  // Prevent redirect loops
  if (isLogoutInProgress()) {
    console.warn('[Auth] Logout already in progress, skipping redirect');
    return;
  }

  // Check if we're already on the signout path
  if (window.location.pathname === '/signout') {
    console.warn('[Auth] Already on signout path, skipping redirect');
    return;
  }

  setLogoutInProgress();

  // Build the Keycloak logout URL
  // post_logout_redirect_uri -> Envoy /signout to clear proxy cookies
  const logoutUrl = new URL(AUTH_CONFIG.keycloakLogoutUrl);
  logoutUrl.searchParams.set('client_id', AUTH_CONFIG.clientId);
  logoutUrl.searchParams.set('post_logout_redirect_uri', AUTH_CONFIG.envoySignoutUrl);

  console.info('[Auth] Initiating logout redirect to Keycloak');
  window.location.href = logoutUrl.toString();
}

/**
 * Handle 401 Unauthorized responses
 * Returns true if logout was triggered, false if skipped
 */
export function handleUnauthorized(): boolean {
  if (isLogoutInProgress()) {
    console.warn('[Auth] 401 received but logout already in progress');
    return false;
  }

  console.info('[Auth] 401 Unauthorized - triggering logout');
  performLogout();
  return true;
}

/**
 * Get auth configuration (useful for debugging)
 */
export function getAuthConfig() {
  return { ...AUTH_CONFIG };
}
