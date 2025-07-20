import { createContext, ParentProps, useContext } from "solid-js";

export const AuthContext = createContext<string>();

export function useAuth(): string {
	const context = useContext(AuthContext);

	if (!context) {
		throw new Error("useAuth must be used within an AuthContextProvider");
	}

	return context;
}

export default function AuthContextProvider(props: ParentProps) {
	const userHash = "CAJIV6QHSXUTNMG4GS2STKIOLLX64O5SS3Q6YLWN6UEC45ZVFXKA====";

	return (
		<AuthContext.Provider value={userHash}>
			{props.children}
		</AuthContext.Provider>
	);
}
