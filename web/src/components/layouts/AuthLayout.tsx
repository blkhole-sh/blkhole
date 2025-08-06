import { ParentProps, Show, createEffect } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { useAuth } from "~/context/AuthContext";
import { TabBarLayout } from "~/components/tabbar/TabBarLayout";

export default function AuthLayout(props: ParentProps) {
	const auth = useAuth();
	const navigate = useNavigate();

	// Reactive auth check - runs whenever auth state changes
	createEffect(() => {
		if (!auth.isAuthenticated()) {
			navigate("/login", { replace: true });
		}
	});

	return (
		<Show when={auth.isAuthenticated()} fallback={<div>Redirecting...</div>}>
			<TabBarLayout tabs={[
				{ title: "Devices", path: "/devices" },
				{ title: "Lists", path: "/lists" },
				{ title: "Schedules", path: "/schedules" },
			]}>
				{props.children}
			</TabBarLayout>
		</Show>
	);
}
