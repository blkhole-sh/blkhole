/**
 * API client for authenticated requests to the Leo server backend
 */
import { Device, List, Quote, Schedule } from "./model";
import { getCookie, clearAuthCookies, setCookie } from "./cookies";

const API_BASE = "/api";

/**
 * Login user with email and password
 * @param email - User email
 * @param password - User password
 * @returns Promise containing user data and token
 * @throws Error on login failure
 */
export const login = async (email: string, password: string) => {
	const response = await fetch(`${API_BASE}/auth/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password }),
	});

	if (!response.ok) {
		throw new Error(await response.text());
	}

	const { user: userData, token: userToken } = await response.json();

	setCookie("token", userToken);
	setCookie("user", encodeURIComponent(JSON.stringify(userData)));

	return { user: userData, token: userToken };
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
		// Clear expired/invalid tokens
		clearAuthCookies();
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
