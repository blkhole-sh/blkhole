import { render } from "solid-js/web";
import { Router, Route } from "@solidjs/router";
import { Suspense } from "solid-js";
import "virtual:uno.css";
import "./app.css";

import SignIn from "./routes/auth/SignIn";
import SignUp from "./routes/auth/SignUp";
import Dashboard from "./routes/Dashboard";
import Devices from "./routes/Devices";
import Lists from "./routes/Lists";
import Schedules from "./routes/Schedules";
import AuthLayout from "./components/AuthLayout";
import AuthProvider from "./context/AuthContext";
import Blocked from "./routes/Blocked";

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
					</Route>
				</Suspense>
			</Router>
		</AuthProvider>
	),
	root!,
);
