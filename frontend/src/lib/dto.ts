// Define UserDTO interface
export interface UserDTO {
	hash: string;
	name: string;
	email: string;
}

// Define DeviceDTO interface
export interface DeviceDTO {
	hash: string;
	name: string;
	os: string;
}

// Define ListDTO interface
export interface ListDTO {
	id: number;
	name: string;
}

// Define ScheduleDTO interface
export interface ScheduleDTO {
	id: number;
	name: string;
	startTime: Date;
	endTime: Date;
	monday: boolean;
	tuesday: boolean;
	wednesday: boolean;
	thursday: boolean;
	friday: boolean;
	saturday: boolean;
	sunday: boolean;
}
