import {
	createContext,
	createSignal,
	ParentProps,
	useContext,
	onMount,
} from "solid-js";
import { User } from "~/lib/model";
import { getCookie, clearAuthCookies } from "~/lib/cookies";
import { login as apiLogin } from "~/lib/api";

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
				clearAuthCookies();
			}
		}
	});

	const login = async (email: string, password: string) => {
		const { user: userData, token: userToken } = await apiLogin(email, password);
		setToken(userToken);
		setUser(userData);
	};

	const logout = () => {
		clearAuthCookies();
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
		<AuthContext.Provider value={store}>{props.children}</AuthContext.Provider>
	);
}
