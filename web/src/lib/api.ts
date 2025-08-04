import { useAuth } from "~/context/AuthContext";
import { Device, List, Quote, Schedule } from "./model";

// Load random stoic quote from backend
export async function getQuote(): Promise<Quote> {
	return (await fetch("http://127.0.0.1:8080/quote")).json();
}

export async function getDevices(): Promise<Device[]> {
	const userHash = useAuth();
	return (
		await fetch(`http://127.0.0.1:8080/api/users/${userHash}/devices`)
	).json();
}

export async function getLists(): Promise<List[]> {
	const userHash = useAuth();
	return (
		await fetch(`http://127.0.0.1:8080/api/users/${userHash}/lists`)
	).json();
}

export async function getSchedules(): Promise<Schedule[]> {
	const userHash = useAuth();
	return (
		await fetch(`http://127.0.0.1:8080/api/users/${userHash}/schedules`)
	).json();
}
