import { useLocation, useNavigate } from "@solidjs/router";
import { For } from "solid-js";
import { useAuth } from "~/context/AuthContext";
import ActionButton from "../ui/ActionButton";
import Logo from "../ui/Logo";

const routes = [
	{ href: "/", label: "DASHBOARD" },
	{ href: "/devices", label: "DEVICES" },
	{ href: "/lists", label: "BLOCKLISTS" },
	{ href: "/schedules", label: "SCHEDULES" },
	{ href: "/settings", label: "SETTINGS" },
];

export default function Sidebar() {
	const auth = useAuth();
	const location = useLocation();
	const navigate = useNavigate();

	const signOut = async () => {
		await auth.logout();
		navigate("/auth/signin");
	};

	return (
		<aside class="sticky top-0 h-screen w-60 flex-shrink-0 flex flex-col border-r border-zinc-200 bg-white">
			<div class="h-16 px-8 flex-shrink-0 flex items-center border-b border-zinc-200">
				<a href="/" class="text-black">
					<Logo class="h-[18px]" />
				</a>
			</div>
			<nav class="pt-12">
				<ul class="flex flex-col text-sm tracking-wider">
					<For each={routes}>
						{(route) => (
							<li>
								<a
									href={route.href}
									class="block py-2.5 pr-8 pl-7.5 border-l-2"
									classList={{
										"border-black text-black font-medium":
											location.pathname === route.href,
										"border-transparent text-zinc-500 hover:text-black":
											location.pathname !== route.href,
									}}
								>
									{route.label}
								</a>
							</li>
						)}
					</For>
				</ul>
			</nav>
			<div class="mt-auto px-8 pb-8">
				<ActionButton onclick={signOut} class="text-zinc-500 hover:text-black">
					SIGN OUT
				</ActionButton>
			</div>
		</aside>
	);
}
