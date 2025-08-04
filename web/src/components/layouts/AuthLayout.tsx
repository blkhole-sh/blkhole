import { ParentProps, Show, onMount } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { useAuth } from "~/context/AuthContext";
import { TabBarLayout } from "~/components/tabbar/TabBarLayout";

export default function AuthLayout(props: ParentProps) {
	const auth = useAuth();
	const navigate = useNavigate();

	onMount(() => {
		if (!auth.isAuthenticated()) {
			navigate("/login");
		}
	});

	return (
		<Show when={auth.isAuthenticated()} fallback={<div>Loading...</div>}>
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
