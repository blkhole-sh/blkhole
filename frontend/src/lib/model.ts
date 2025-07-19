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

// Define Domain interface
export interface Domain {
	id: number;
	name: string;
	listIds: number[];
	scheduleIds: number[];
}

// Define List interface
export interface List {
	id: number;
	name: string;
	description?: string;
	source?: string;
	userHash: string;
	domainIds: number[];
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
	domainIds: number[];
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
