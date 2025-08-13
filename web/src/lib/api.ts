/**
 * API client for authenticated requests to the Leo server backend
 */
import { Device, List, Quote, Schedule } from "./model";

const API_BASE = import.meta.env.DEV ? "http://localhost:8080/api" : "/api";

/**
 * Login user with email and password
 * Backend will set secure HttpOnly cookies automatically
 * @param email - User email
 * @param password - User password
 * @returns Promise containing user data
 * @throws Error on login failure
 */
export const login = async (email: string, password: string) => {
	const response = await fetch(`${API_BASE}/auth/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password }),
		credentials: "include", // Include cookies in request
	});

	if (!response.ok) {
		throw new Error(await response.text());
	}

	const responseData = await response.json();
	console.log("Login response:", responseData);
	const { user: userData } = responseData;
	console.log("User data:", userData);
	console.log("User id:", userData?.id);

	// Store user data in localStorage (not sensitive)
	localStorage.setItem("user", JSON.stringify(userData));

	return { user: userData };
};

/**
 * Logout user by calling backend logout endpoint
 * @returns Promise that resolves when logout is complete
 */
export const logout = async () => {
	try {
		await fetch(`${API_BASE}/auth/logout`, {
			method: "POST",
			credentials: "include",
		});
	} catch (error) {
		// Logout endpoint might fail, but we still want to clear local data
		console.warn("Logout endpoint failed:", error);
	}

	// Clear local user data
	localStorage.removeItem("user");
};

/**
 * Refresh authentication tokens
 * @returns Promise containing updated user data
 */
export const refreshAuth = async () => {
	const response = await fetch(`${API_BASE}/auth/refresh`, {
		method: "POST",
		credentials: "include",
	});

	if (!response.ok) {
		throw new Error("Token refresh failed");
	}

	const responseData = await response.json();
	const { user: userData } = responseData;
	localStorage.setItem("user", JSON.stringify(userData));
	return { user: userData };
};

/**
 * Make an authenticated API request
 * Uses HttpOnly cookies for authentication, handles token refresh automatically
 * @param endpoint - API endpoint path (e.g., "/quote")
 * @param options - Additional fetch options
 * @returns Parsed JSON response
 * @throws Error on API errors or authentication failure
 */
const api = async (endpoint: string, options?: RequestInit): Promise<any> => {
	const makeRequest = async (includeRefresh = true) => {
		const response = await fetch(`${API_BASE}${endpoint}`, {
			headers: {
				"Content-Type": "application/json",
			},
			credentials: "include", // Include HttpOnly cookies
			...options,
		});

		if (response.status === 401 && includeRefresh) {
			// Try to refresh tokens
			try {
				await refreshAuth();
				// Retry the original request once
				return makeRequest(false);
			} catch (refreshError) {
				// Refresh failed, clear local data and redirect
				localStorage.removeItem("user");
				window.location.href = "/login";
				throw new Error("Authentication failed");
			}
		}

		if (!response.ok) {
			throw new Error(`API Error: ${response.status}`);
		}

		return response.json();
	};

	return makeRequest();
};

/**
 * Get current user data from localStorage
 * @returns Parsed user object
 * @throws Error if user data is missing or invalid
 */
const getCurrentUser = () => {
	const userData = localStorage.getItem("user");
	if (!userData) throw new Error("No user data");
	try {
		return JSON.parse(userData);
	} catch {
		throw new Error("Invalid user data");
	}
};

// API endpoints for different resources

/** Get a random stoic quote */
export const getQuote = (): Promise<Quote> => api("/quote");

/** Get all devices for the current user - returns DeviceDTO with schedule counts */
export const getDevices = (): Promise<Device[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}/devices`);
};

/** Get all domain lists for the current user - returns ListDTO with rule and schedule counts */
export const getLists = (): Promise<List[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}/lists`);
};

/** Get all schedules for the current user - returns ScheduleDTO with device, domain, and list counts */
export const getSchedules = (): Promise<Schedule[]> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}/schedules`);
};
