import { useNavigate } from "@solidjs/router";
import { createEffect, type ParentProps, Show } from "solid-js";
import Footer from "~/components/layout/Footer";
import Navbar from "~/components/layout/Navbar";
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
			<div class="min-h-screen flex flex-col">
				<Navbar />
				<div class="flex-1 flex flex-col">{props.children}</div>
				<Footer />
			</div>
		</Show>
	);
}
