import {
	createContext,
	createSignal,
	ParentProps,
	useContext,
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

const loadUser = (): User | null => {
	const saved = localStorage.getItem("user");
	if (!saved) return null;
	try {
		return JSON.parse(saved);
	} catch {
		localStorage.removeItem("user");
		return null;
	}
};

export default function AuthProvider(props: ParentProps) {
	const [user, setUser] = createSignal<User | null>(loadUser());

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
