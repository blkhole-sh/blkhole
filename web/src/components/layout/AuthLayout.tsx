import { useNavigate } from "@solidjs/router";
import { createEffect, type ParentProps, Show } from "solid-js";
import Footer from "~/components/layout/Footer";
import Sidebar from "~/components/layout/Sidebar";
import { useAuth } from "~/context/AuthContext";

export default function AuthLayout(props: ParentProps) {
	const auth = useAuth();
	const navigate = useNavigate();

	createEffect(() => {
		if (!auth.isAuthenticated()) {
			navigate("/auth/signin", { replace: true });
		}
	});

	return (
		<Show when={auth.isAuthenticated()}>
			<div class="min-h-screen flex flex-row">
				<Sidebar />
				<div class="min-w-0 flex-1 flex flex-col">
					<div class="flex-1 flex flex-col">{props.children}</div>
					<Footer />
				</div>
			</div>
		</Show>
	);
}
