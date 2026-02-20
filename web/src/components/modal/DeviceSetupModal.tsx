import { Show } from "solid-js";
import Modal from "./Modal";
import { Device } from "~/lib/model";

interface Props {
	open: boolean;
	device: Device | null;
	onClose: () => void;
}

export default function DeviceSetupModal(props: Props) {
	const configUrl = () => `/api/devices/${props.device?.id}/config`;

	const isApple = () => {
		const osLower = props.device?.os.toLowerCase();
		return (
			osLower?.includes("ios") ||
			osLower?.includes("macos") ||
			osLower?.includes("mac")
		);
	};

	const isIOS = () => props.device?.os.toLowerCase().includes("ios");
	const isMacOS = () => {
		const osLower = props.device?.os.toLowerCase();
		return osLower?.includes("macos") || osLower?.includes("mac");
	};
	const isAndroid = () => props.device?.os.toLowerCase().includes("android");
	const isWindows = () => props.device?.os.toLowerCase().includes("windows");
	const isLinux = () => props.device?.os.toLowerCase().includes("linux");

	return (
		<Modal
			title="Device Setup"
			open={props.open}
			onClose={props.onClose}
			onConfirm={props.onClose}
			confirmLabel="DONE"
		>
			<div class="flex flex-col gap-6">
				<Show when={isApple()}>
					<p class="text-sm text-zinc-700">
						Download and install the configuration profile to enable DNS
						filtering.
					</p>
					<a
						href={configUrl()}
						class="text-sm text-black font-medium underline hover:text-zinc-600"
						download
					>
						Download Configuration Profile
					</a>
					<div class="bg-zinc-50 p-4 rounded flex flex-col gap-3 text-sm">
						<p class="font-medium text-zinc-700">Setup Instructions:</p>
						<Show when={isIOS()}>
							<ol class="list-decimal list-inside space-y-2 text-zinc-600">
								<li>Tap the downloaded .mobileconfig file</li>
								<li>
									Go to{" "}
									<span class="font-medium">Settings → Profile Downloaded</span>
								</li>
								<li>
									Tap <span class="font-medium">Install</span> and follow the
									prompts
								</li>
							</ol>
						</Show>
						<Show when={isMacOS()}>
							<ol class="list-decimal list-inside space-y-2 text-zinc-600">
								<li>Open the downloaded .mobileconfig file</li>
								<li>
									Go to{" "}
									<span class="font-medium">
										System Settings → Privacy & Security → Profiles
									</span>
								</li>
								<li>
									Select the profile and click{" "}
									<span class="font-medium">Install</span>
								</li>
							</ol>
						</Show>
					</div>
				</Show>

				<Show when={isAndroid()}>
					<p class="text-sm text-zinc-700">
						Configure Private DNS to enable system-wide DNS filtering.
					</p>
					<div class="bg-zinc-50 p-4 rounded flex flex-col gap-3 text-sm">
						<p class="font-medium text-zinc-700">Setup Instructions:</p>
						<ol class="list-decimal list-inside space-y-2 text-zinc-600">
							<li>
								Go to{" "}
								<span class="font-medium">
									Settings → Network & Internet → Private DNS
								</span>
							</li>
							<li>
								Select{" "}
								<span class="font-medium">Private DNS provider hostname</span>
							</li>
							<li>
								Enter:{" "}
								<code class="bg-white px-2 py-0.5 rounded text-xs font-mono text-zinc-800">
									{props.device?.id}.dns.yourdomain.com
								</code>
							</li>
							<li>
								Tap <span class="font-medium">Save</span>
							</li>
						</ol>
					</div>
				</Show>

				<Show when={isWindows()}>
					<p class="text-sm text-zinc-700">
						Configure DNS over HTTPS to enable system-wide DNS filtering.
					</p>
					<div class="bg-zinc-50 p-4 rounded flex flex-col gap-3 text-sm">
						<p class="font-medium text-zinc-700">Setup Instructions:</p>
						<ol class="list-decimal list-inside space-y-2 text-zinc-600">
							<li>
								Go to{" "}
								<span class="font-medium">
									Settings → Network & Internet → Ethernet/WiFi
								</span>
							</li>
							<li>
								Click{" "}
								<span class="font-medium">DNS server assignment → Edit</span>
							</li>
							<li>
								Select <span class="font-medium">Manual</span> and toggle on{" "}
								<span class="font-medium">IPv4</span>
							</li>
							<li>
								Enter DNS server:{" "}
								<code class="bg-white px-2 py-0.5 rounded text-xs font-mono text-zinc-800">
									your-server-ip
								</code>
							</li>
							<li>
								Toggle <span class="font-medium">DNS over HTTPS</span> to{" "}
								<span class="font-medium">On (automatic template)</span>
							</li>
							<li>
								Click <span class="font-medium">Save</span>
							</li>
						</ol>
					</div>
				</Show>

				<Show when={isLinux()}>
					<p class="text-sm text-zinc-700">
						Configure DNS over TLS using systemd-resolved to enable system-wide
						DNS filtering.
					</p>
					<div class="bg-zinc-50 p-4 rounded flex flex-col gap-3 text-sm">
						<p class="font-medium text-zinc-700">Setup Instructions:</p>
						<ol class="list-decimal list-inside space-y-2 text-zinc-600">
							<li>
								Edit the configuration file:{" "}
								<code class="bg-white px-2 py-0.5 rounded text-xs font-mono text-zinc-800">
									/etc/systemd/resolved.conf
								</code>
							</li>
							<li>
								Add the following under{" "}
								<code class="bg-white px-2 py-0.5 rounded text-xs font-mono text-zinc-800">
									[Resolve]
								</code>{" "}
								section:
								<div class="bg-white p-3 rounded text-xs font-mono text-zinc-800 mt-2">
									DNS=your-server-ip
									<br />
									DNSOverTLS=yes
								</div>
							</li>
							<li>
								Save the file and restart the service:
								<div class="bg-white p-3 rounded text-xs font-mono text-zinc-800 mt-2">
									sudo systemctl restart systemd-resolved
								</div>
							</li>
						</ol>
					</div>
				</Show>
			</div>
		</Modal>
	);
}
