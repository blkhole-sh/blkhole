import { createContext, ParentProps, Show, useContext } from "solid-js";

export const AuthContext = createContext<string>();

export function useAuth(): string {
	const context = useContext(AuthContext);

	if (!context) {
		throw new Error("useAuth must be used within an AuthContextProvider");
	}

	return context;
}

function getCookie(name: string): string | null {
	if (typeof document === "undefined") {
		return null;
	}
	const value = `; ${document.cookie}`;
	const parts = value.split(`; ${name}=`);
	if (parts.length === 2) {
		return parts.pop()?.split(";").shift() || null;
	}
	return null;
}

export default function AuthContextProvider(props: ParentProps) {
	const userHash =
		getCookie("userHash") ||
		"CAJIV6QHSXUTNMG4GS2STKIOLLX64O5SS3Q6YLWN6UEC45ZVFXKA====";

	return (
		<AuthContext.Provider value={userHash}>
			{props.children}
		</AuthContext.Provider>
	);
}
