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
	const userHash = "XZ7J2U3R2MUJWPBYV36XYDCTQ47LBFKE7GHNFIXKWAU2GETX2HIA====";

	return (
		<AuthContext.Provider value={userHash}>
			{props.children}
		</AuthContext.Provider>
	);
}
