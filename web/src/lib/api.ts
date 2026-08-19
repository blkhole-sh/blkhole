/**
 * API client for authenticated requests to the blkhole server backend
 */
import type {
	BrowserClient,
	BrowserPairing,
	Device,
	List,
	QueryStats,
	Quote,
	Schedule,
	Settings,
} from "./model";

const API_BASE = "/api";

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
	const { user: userData } = responseData;

	// Store user data in localStorage (not sensitive)
	localStorage.setItem("user", JSON.stringify(userData));

	return { user: userData };
};

/**
 * Register user with email and password
 * Backend will set secure HttpOnly cookies automatically
 * @param email - User email
 * @param password - User password
 * @returns Promise containing user data
 * @throws Error on registration failure
 */
export const register = async (email: string, password: string) => {
	const response = await fetch(`${API_BASE}/auth/register`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password }),
		credentials: "include", // Include cookies in request
	});

	if (!response.ok) {
		throw new Error(await response.text());
	}

	const responseData = await response.json();
	const { user: userData } = responseData;

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
const api = async <T>(endpoint: string, options?: RequestInit): Promise<T> => {
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
			} catch {
				// Refresh failed, clear local data and redirect
				localStorage.removeItem("user");
				window.location.href = "/auth/signin";
				throw new Error("Authentication failed");
			}
		}

		if (!response.ok) {
			throw new Error(`API Error: ${response.status}`);
		}

		// Handle 204 No Content responses (e.g., successful DELETE)
		if (response.status === 204) {
			return;
		}

		return response.json() as Promise<T>;
	};

	return makeRequest() as Promise<T>;
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

/** Create a new device for the current user */
export const createDevice = (name: string, os: string): Promise<Device> => {
	const user = getCurrentUser();
	return api("/devices", {
		method: "PUT",
		body: JSON.stringify({ name, os, userId: user.id }),
	});
};

/** Update a device by ID */
export const updateDevice = (
	id: string,
	name: string,
	os: string,
): Promise<Device> =>
	api(`/devices/${id}`, {
		method: "PATCH",
		body: JSON.stringify({ name, os }),
	});

/** Delete a device by ID */
export const deleteDevice = (id: string): Promise<void> =>
	api(`/devices/${id}`, { method: "DELETE" });

/** Create a short-lived browser extension pairing for a device. */
export const createBrowserPairing = (id: string): Promise<BrowserPairing> =>
	api(`/devices/${id}/browser-pairings`, { method: "POST" });

/** List browser installations paired with a device. */
export const getBrowserClients = (id: string): Promise<BrowserClient[]> =>
	api(`/devices/${id}/browser-clients`);

/** Revoke a browser installation paired with a device. */
export const revokeBrowserClient = (
	deviceId: string,
	clientId: string,
): Promise<void> =>
	api(`/devices/${deviceId}/browser-clients/${clientId}`, {
		method: "DELETE",
	});

/** Update a blocklist by ID */
export const updateList = (
	id: number,
	name: string,
	description: string,
	source: string,
): Promise<List> =>
	api(`/lists/${id}`, {
		method: "PATCH",
		body: JSON.stringify({ name, description, source }),
	});

/** Delete a blocklist by ID */
export const deleteList = (id: number): Promise<void> =>
	api(`/lists/${id}`, { method: "DELETE" });

/** Update a schedule by ID */
export const updateSchedule = (
	id: number,
	name: string,
	startTime: string,
	endTime: string,
	active: boolean,
	days: {
		monday: boolean;
		tuesday: boolean;
		wednesday: boolean;
		thursday: boolean;
		friday: boolean;
		saturday: boolean;
		sunday: boolean;
	},
	listIds: number[],
	deviceIds: string[],
): Promise<Schedule> =>
	api(`/schedules/${id}`, {
		method: "PATCH",
		body: JSON.stringify({
			name,
			startTime,
			endTime,
			active,
			listIds,
			deviceIds,
			...days,
		}),
	});

/** Delete a schedule by ID */
export const deleteSchedule = (id: number): Promise<void> =>
	api(`/schedules/${id}`, { method: "DELETE" });

/** Create a new blocklist for the current user */
export const createList = (
	name: string,
	description: string,
	source: string,
): Promise<List> => {
	const user = getCurrentUser();
	return api("/lists", {
		method: "PUT",
		body: JSON.stringify({ name, description, source, userId: user.id }),
	});
};

/** Create a new schedule for the current user */
export const createSchedule = (
	name: string,
	startTime: string,
	endTime: string,
	days: {
		monday: boolean;
		tuesday: boolean;
		wednesday: boolean;
		thursday: boolean;
		friday: boolean;
		saturday: boolean;
		sunday: boolean;
	},
	listIds: number[],
	deviceIds: string[],
): Promise<Schedule> => {
	const user = getCurrentUser();
	return api("/schedules", {
		method: "PUT",
		body: JSON.stringify({
			name,
			startTime,
			endTime,
			active: true,
			userId: user.id,
			listIds,
			deviceIds,
			...days,
		}),
	});
};

/** Get query statistics for the current user - returns total and blocked query counts over time */
export const getQueryStats = (
	range: "24h" | "7d" | "30d" = "24h",
): Promise<QueryStats> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}/stats/queries?range=${range}`);
};

/** Get server settings (upstream DNS, etc.) — no auth required */
export const getSettings = (): Promise<Settings> => {
	return fetch(`${API_BASE}/settings`).then((r) => {
		if (!r.ok) throw new Error(`Failed to fetch settings: ${r.status}`);
		return r.json();
	});
};

/**
 * Change the current user's password
 * @param currentPassword - The user's current password for verification
 * @param newPassword - The new password to set
 */
export const changePassword = (
	currentPassword: string,
	newPassword: string,
): Promise<void> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}`, {
		method: "PATCH",
		body: JSON.stringify({ currentPassword, newPassword }),
	});
};

/** Delete the current user's account */
export const deleteAccount = (): Promise<void> => {
	const user = getCurrentUser();
	return api(`/users/${user.id}`, { method: "DELETE" });
};
