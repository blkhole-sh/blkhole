import { ParentProps } from "solid-js";
import SideBar from "~/components/sidebar/SideBar";
import AuthContextProvider from "~/context/AuthContext";

export default function AuthLayout(props: ParentProps) {
	return (
		<AuthContextProvider>
			<div class="font-inter flex flex-row h-screen w-full overflow-hidden">
				<div class="w-72 h-full">
					<SideBar />
				</div>
				<div class="flex-1 overflow-hidden">{props.children}</div>
			</div>
		</AuthContextProvider>
	);
}
