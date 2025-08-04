import { Router, Route } from "@solidjs/router";
import { Suspense } from "solid-js";
import "./app.css";

// Import your page components
import Login from "./pages/login/Login";
import Dashboard from "./pages/dashboard/Dashboard";
import Devices from "./pages/devices/Devices";
import DeviceDetail from "./pages/devices/DeviceDetail";
import NewDevice from "./pages/devices/NewDevice";
import Lists from "./pages/lists/Lists";
import ListDetail from "./pages/lists/ListDetail";
import NewList from "./pages/lists/NewList";
import Schedules from "./pages/schedules/Schedules";
import ScheduleDetail from "./pages/schedules/ScheduleDetail";
import NewSchedule from "./pages/schedules/NewSchedule";
import Blocked from "./pages/blocked/Blocked";
import AuthLayout from "./components/layouts/AuthLayout";
import AuthProvider from "./context/AuthContext";

export default function App() {
	return (
		<AuthProvider>
			<Router>
				<Suspense>
					<Route path="/login" component={Login} />
					<Route path="/blocked" component={Blocked} />
					<Route path="/" component={AuthLayout}>
						<Route path="/" component={Dashboard} />
						<Route path="/devices" component={Devices} />
						<Route path="/devices/:id" component={DeviceDetail} />
						<Route path="/devices/new" component={NewDevice} />
						<Route path="/lists" component={Lists} />
						<Route path="/lists/:id" component={ListDetail} />
						<Route path="/lists/new" component={NewList} />
						<Route path="/schedules" component={Schedules} />
						<Route path="/schedules/:id" component={ScheduleDetail} />
						<Route path="/schedules/new" component={NewSchedule} />
					</Route>
				</Suspense>
			</Router>
		</AuthProvider>
	);
}

