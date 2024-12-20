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
  const userHash = "VN7KGKKMROH233SESMCRTJASL3GUUCO7JJ2E2YSOVA2OAL64S7UA====";

  return (
    <AuthContext.Provider value={userHash}>
      {props.children}
    </AuthContext.Provider>
  );
}
