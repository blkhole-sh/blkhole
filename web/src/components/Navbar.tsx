import { useLocation, useNavigate } from "@solidjs/router";
import { useAuth } from "~/context/AuthContext";
import ActionButton from "./ActionButton";
import Logo from "./Logo";

export default function Navbar() {
	const auth = useAuth();
	const location = useLocation();
	const navigate = useNavigate();

	const isActive = (path: string) => location.pathname === path;

	const signOut = async () => {
		await auth.logout();
		navigate("/auth/signin");
	};

	return (
		<div class="sticky top-0 z-10 px-24 py-8 flex flex-row justify-between items-baseline font-inter text-sm text-zinc-500 bg-white">
			<div class="flex flex-row items-center gap-12">
				<a href="/" class="text-black">
					<Logo class="h-5" />
				</a>
				<nav>
					<ul class="flex flex-row gap-8 tracking-wider">
						<li>
							<a href="/" classList={{ "text-black underline": isActive("/") }}>
								DASHBOARD
							</a>
						</li>
						<li>
							<a
								href="/devices"
								classList={{ "text-black underline": isActive("/devices") }}
							>
								DEVICES
							</a>
						</li>
						<li>
							<a
								href="/lists"
								classList={{ "text-black underline": isActive("/lists") }}
							>
								BLOCKLISTS
							</a>
						</li>
						<li>
							<a
								href="/schedules"
								classList={{ "text-black underline": isActive("/schedules") }}
							>
								SCHEDULES
							</a>
						</li>
					</ul>
				</nav>
			</div>
			<div class="flex flex-row items-center gap-8">
				<ActionButton>{auth.user()?.email.toUpperCase()}</ActionButton>
				<ActionButton onclick={signOut}>SIGN OUT</ActionButton>
			</div>
		</div>
	);
}
