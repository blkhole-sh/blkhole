/**
 * Cookie management utilities
 */

/**
 * Get a cookie value by name
 * @param name - Cookie name to retrieve
 * @returns Cookie value or null if not found
 */
export const getCookie = (name: string): string | null => {
	const match = document.cookie.match(new RegExp(`(^| )${name}=([^;]+)`));
	return match?.[2] || null;
};

/**
 * Set a cookie with expiration
 * @param name - Cookie name
 * @param value - Cookie value
 * @param hours - Expiration in hours (default: 24)
 */
export const setCookie = (name: string, value: string, hours = 24) => {
	const expires = new Date(Date.now() + hours * 3600000).toUTCString();
	document.cookie = `${name}=${value}; expires=${expires}; path=/; SameSite=Strict`;
};

/**
 * Delete a cookie by setting it to expire in the past
 * @param name - Cookie name to delete
 */
export const deleteCookie = (name: string) => {
	document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
};

/**
 * Clear authentication cookies (token and user)
 */
export const clearAuthCookies = () => {
	deleteCookie("token");
	deleteCookie("user");
};

