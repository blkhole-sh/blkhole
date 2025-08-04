import { createContext, createSignal, ParentProps, useContext, onMount } from "solid-js";
import { User } from "~/lib/model";

type AuthStore = {
	user: () => User | null;
	token: () => string | null;
	login: (email: string, password: string) => Promise<void>;
	logout: () => void;
	isAuthenticated: () => boolean;
};

const AuthContext = createContext<AuthStore>();

export const useAuth = (): AuthStore => {
	const context = useContext(AuthContext);
	if (!context) throw new Error("useAuth must be used within AuthProvider");
	return context;
};

const API_BASE = "http://127.0.0.1:8080/api";

const getCookie = (name: string): string | null => {
	if (typeof document === "undefined") return null;
	const match = document.cookie.match(new RegExp(`(^| )${name}=([^;]+)`));
	return match ? match[2] : null;
};

const setCookie = (name: string, value: string, hours = 24) => {
	if (typeof document === "undefined") return;
	const expires = new Date(Date.now() + hours * 3600000).toUTCString();
	document.cookie = `${name}=${value}; expires=${expires}; path=/; SameSite=Strict`;
};

const deleteCookie = (name: string) => {
	if (typeof document === "undefined") return;
	document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
};

export default function AuthProvider(props: ParentProps) {
	const [user, setUser] = createSignal<User | null>(null);
	const [token, setToken] = createSignal<string | null>(null);

	onMount(() => {
		const savedToken = getCookie("token");
		const savedUser = getCookie("user");
		
		if (savedToken && savedUser) {
			try {
				setToken(savedToken);
				setUser(JSON.parse(decodeURIComponent(savedUser)));
			} catch {
				// Invalid stored data, clear cookies
				deleteCookie("token");
				deleteCookie("user");
			}
		}
	});

	const login = async (email: string, password: string) => {
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
		
		setToken(userToken);
		setUser(userData);
	};

	const logout = () => {
		deleteCookie("token");
		deleteCookie("user");
		setToken(null);
		setUser(null);
	};

	const isAuthenticated = () => Boolean(token() && user());

	const store: AuthStore = {
		user,
		token,
		login,
		logout,
		isAuthenticated,
	};

	return (
		<AuthContext.Provider value={store}>
			{props.children}
		</AuthContext.Provider>
	);
}
