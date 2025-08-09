// Define User interface - now uses DTO format with counts
export interface User {
	hash: string;
	name: string;
	email: string;
	devices: number;
	lists: number;
	schedules: number;
}

// Define Device interface - now uses DTO format with counts
export interface Device {
	hash: string;
	name: string;
	os: string;
	userHash: string;
	schedules: number;
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
	userHash: string;
	rules: number;    // Count of rules in this list
	schedules: number; // Count of schedules using this list
}

// Define Schedule interface - now uses DTO format with counts
export interface Schedule {
	id: number;
	name: string;
	startTime: string; // Backend sends as string, not Date
	endTime: string; // Backend sends as string, not Date
	userHash: string;
	devices: number;
	rules: number;
	lists: number;
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
