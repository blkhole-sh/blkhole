import { render } from "solid-js/web";
import { Router, Route } from "@solidjs/router";
import { Suspense, lazy } from "solid-js";
import "virtual:uno.css";
import "./app.css";

// Lazy load routes for code splitting
const SignIn = lazy(() => import("./routes/auth/SignIn"));
const SignUp = lazy(() => import("./routes/auth/SignUp"));
const Dashboard = lazy(() => import("./routes/Dashboard"));
const Devices = lazy(() => import("./routes/Devices"));
const Lists = lazy(() => import("./routes/Lists"));
const Schedules = lazy(() => import("./routes/Schedules"));
const Blocked = lazy(() => import("./routes/Blocked"));
const Settings = lazy(() => import("./routes/Settings"));

// Keep layout/context eager loaded (needed immediately)
import AuthLayout from "./components/layout/AuthLayout";
import AuthProvider from "./context/AuthContext";

const root = document.getElementById("root");

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
	throw new Error(
		"Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?",
	);
}

render(
	() => (
		<AuthProvider>
			<Router>
				<Suspense>
					<Route path="/auth/signin" component={SignIn} />
					<Route path="/auth/signup" component={SignUp} />
					<Route path="/blocked" component={Blocked} />
					<Route path="/" component={AuthLayout}>
						<Route path="" component={Dashboard} />
						<Route path="/devices" component={Devices} />
						<Route path="/lists" component={Lists} />
						<Route path="/schedules" component={Schedules} />
						<Route path="/settings" component={Settings} />
					</Route>
				</Suspense>
			</Router>
		</AuthProvider>
	),
	root!,
);
