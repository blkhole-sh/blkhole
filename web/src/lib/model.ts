// Define User interface
export interface User {
	hash: string;
	name: string;
	email: string;
	deviceHashes: string[];
	listIds: number[];
	scheduleIds: number[];
}

// Define Device interface
export interface Device {
	hash: string;
	name: string;
	os: string;
	userHash: string;
	scheduleIds: number[];
}

// Define Rule interface
export interface Rule {
	id: number;
	domain: string;
	listId: number;
	allowed: boolean;
}

// Define List interface
export interface List {
	id: number;
	name: string;
	description: string;
	source: string;
	userHash: string;
	rules: Rule[];
	scheduleIds: number[];
}

// Define Schedule interface
export interface Schedule {
	id: number;
	name: string;
	startTime: Date;
	endTime: Date;
	userHash: string;
	deviceHashes: string[];
	domains: string[];
	listIds: number[];
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
