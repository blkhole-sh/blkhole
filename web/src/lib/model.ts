// Define User interface - now uses DTO format with counts
export interface User {
	id: string;
	name: string;
	email: string;
	devices: number;
	lists: number;
	schedules: number;
}

// Define Device interface - now uses DTO format with schedule IDs and names
export interface Device {
	id: string;
	hash: string; // Keep hash for mobile config URLs
	name: string;
	os: string;
	userId: string;
	scheduleIds: number[];
	scheduleNames: string[];
}

// Define Rule interface
export interface Rule {
	id: number;
	domain: string;
	listId: number;
	allowed: boolean;
}

// Define List interface - now uses DTO format with counts
export interface List {
	id: number;
	name: string;
	description: string;
	source: string;
	userId: string;
	rules: number; // Count of rules in this list
	schedules: number; // Count of schedules using this list
}

// Define Schedule interface - now uses DTO format with names
export interface Schedule {
	id: number;
	name: string;
	startTime: string; // Backend sends as string, not Date
	endTime: string; // Backend sends as string, not Date
	active: boolean;
	userId: string;
	deviceIds: string[];
	deviceNames: string[];
	listIds: number[];
	listNames: string[];
	rules: number;
	monday: boolean;
	tuesday: boolean;
	wednesday: boolean;
	thursday: boolean;
	friday: boolean;
	saturday: boolean;
	sunday: boolean;
}

// Define Quote interface
export interface Quote {
	quote: string;
	author: string;
}

// Define Settings interface — server configuration exposed via GET /api/settings
export interface Settings {
	upstreamDns: string;
}

// Define StatCount interface for time-series data points
export interface StatCount {
	timestamp: string; // ISO 8601 timestamp
	count: number;
}

// Define QueryStats interface - API response for query statistics
export interface QueryStats {
	total: StatCount[];
	blocked: StatCount[];
}
