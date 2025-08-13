// Define UserDTO interface
export interface UserDTO {
	id: string;
	name: string;
	email: string;
}

// Define DeviceDTO interface
export interface DeviceDTO {
	id: string;
	hash: string; // Keep hash for mobile config URLs
	name: string;
	os: string;
}

// Define RuleDTO interface
export interface RuleDTO {
	domain: string;
	listId: number;
	allowed: boolean;
}

// Define ListDTO interface
export interface ListDTO {
	name: string;
	description: string;
	source: string;
}

// Define ScheduleDTO interface
export interface ScheduleDTO {
	name: string;
	startTime: Date;
	endTime: Date;
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
