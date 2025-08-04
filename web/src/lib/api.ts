/**
 * API client for authenticated requests to the Leo DNS blocker backend
 */
import { Device, List, Quote, Schedule } from "./model";

const API_BASE = "http://127.0.0.1:8080/api";

/**
 * Get a cookie value by name
 * @param name - Cookie name to retrieve
 * @returns Cookie value or null if not found
 */
const getCookie = (name: string): string | null => {
	if (typeof document === "undefined") return null;
	const match = document.cookie.match(new RegExp(`(^| )${name}=([^;]+)`));
	return match?.[2] || null;
};

/**
 * Make an authenticated API request
 * Automatically includes JWT token in Authorization header and handles auth errors
 * @param endpoint - API endpoint path (e.g., "/quote")
 * @param options - Additional fetch options
 * @returns Parsed JSON response
 * @throws Error on API errors or authentication failure
 */
const api = async (endpoint: string, options?: RequestInit) => {
	const token = getCookie("token");
	const response = await fetch(`${API_BASE}${endpoint}`, {
		headers: {
			"Content-Type": "application/json",
			...(token && { Authorization: `Bearer ${token}` }),
		},
		...options,
	});

	if (response.status === 401) {
		// Clear expired/invalid tokens by setting them to expire in the past
		document.cookie = "token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/";
		document.cookie = "user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/";
		window.location.href = "/login";
		throw new Error("Unauthorized");
	}

	if (!response.ok) {
		throw new Error(`API Error: ${response.status}`);
	}

	return response.json();
};

/**
 * Get current user data from cookie
 * @returns Parsed user object
 * @throws Error if user data is missing or invalid
 */
const getCurrentUser = () => {
	const userCookie = getCookie("user");
	if (!userCookie) throw new Error("No user data");
	try {
		return JSON.parse(decodeURIComponent(userCookie));
	} catch {
		throw new Error("Invalid user data");
	}
};

// API endpoints for different resources

/** Get a random stoic quote */
export const getQuote = (): Promise<Quote> => api("/quote");

/** Get all devices for the current user */
export const getDevices = (): Promise<Device[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.hash}/devices`);
};

/** Get all domain lists for the current user */
export const getLists = (): Promise<List[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.hash}/lists`);
};

/** Get all schedules for the current user */
export const getSchedules = (): Promise<Schedule[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.hash}/schedules`);
};
