import {
	createContext,
	createSignal,
	type ParentProps,
	useContext,
} from "solid-js";
import {
	login as apiLogin,
	logout as apiLogout,
	register as apiRegister,
	refreshAuth,
} from "~/lib/api";
import type { User } from "~/lib/model";

type AuthStore = {
	user: () => User | null;
	login: (email: string, password: string) => Promise<void>;
	register: (email: string, password: string) => Promise<void>;
	logout: () => Promise<void>;
	isAuthenticated: () => boolean;
	checkAuth: () => Promise<void>;
	updateUser: (user: User) => void;
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

	const register = async (email: string, password: string) => {
		const { user: userData } = await apiRegister(email, password);
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
	const updateUser = (updated: User) => {
		localStorage.setItem("user", JSON.stringify(updated));
		setUser(updated);
	};

	const store: AuthStore = {
		user,
		login,
		register,
		logout,
		isAuthenticated,
		checkAuth,
		updateUser,
	};

	return (
		<AuthContext.Provider value={store}>{props.children}</AuthContext.Provider>
	);
}
