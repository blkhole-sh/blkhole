import { ParentProps } from "solid-js";
import AuthContextProvider from "~/context/AuthContext";
import { TabBarLayout } from "~/components/tabbar/TabBarLayout";

export default function AuthLayout(props: ParentProps) {
	const tabs = [
		{ title: "Devices", path: "/devices" },
		{ title: "Lists", path: "/lists" },
		{ title: "Schedules", path: "/schedules" },
	];

	return (
		<AuthContextProvider>
			<TabBarLayout tabs={tabs}>{props.children}</TabBarLayout>
		</AuthContextProvider>
	);
}
