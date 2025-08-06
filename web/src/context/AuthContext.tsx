import {
	createContext,
	createSignal,
	ParentProps,
	useContext,
	onMount,
} from "solid-js";
import { User } from "~/lib/model";
import { login as apiLogin, logout as apiLogout, refreshAuth } from "~/lib/api";

type AuthStore = {
	user: () => User | null;
	login: (email: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
	isAuthenticated: () => boolean;
	checkAuth: () => Promise<void>;
};

const AuthContext = createContext<AuthStore>();

export const useAuth = (): AuthStore => {
	const context = useContext(AuthContext);
	if (!context) throw new Error("useAuth must be used within AuthProvider");
	return context;
};

export default function AuthProvider(props: ParentProps) {
	const [user, setUser] = createSignal<User | null>(null);

	onMount(async () => {
		// Try to load user from localStorage
		const savedUser = localStorage.getItem("user");
		if (savedUser) {
			try {
				setUser(JSON.parse(savedUser));
			} catch {
				// Invalid stored data, remove it
				localStorage.removeItem("user");
			}
		}
	});

	const login = async (email: string, password: string) => {
		const { user: userData } = await apiLogin(email, password);
		setUser(userData);
	};

	const logout = async () => {
		await apiLogout();
		setUser(null);
	};

	const checkAuth = async () => {
		try {
			const { user: userData } = await refreshAuth();
			setUser(userData);
		} catch {
			// Not authenticated or refresh failed
			setUser(null);
			localStorage.removeItem("user");
		}
	};

	const isAuthenticated = () => Boolean(user());

	const store: AuthStore = {
		user,
		login,
		logout,
		isAuthenticated,
		checkAuth,
	};

	return (
		<AuthContext.Provider value={store}>{props.children}</AuthContext.Provider>
	);
}
