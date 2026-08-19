import {
	createEffect,
	createResource,
	createSignal,
	For,
	onCleanup,
	Show,
	untrack,
} from "solid-js";
import {
	createBrowserPairing,
	getBrowserClients,
	revokeBrowserClient,
} from "~/lib/api";
import {
	detectBrowserExtension,
	getExtensionInstallTarget,
	pairBrowserExtension,
} from "~/lib/browserExtension";
import type { Device } from "~/lib/model";
import ActionButton from "../ui/ActionButton";

interface Props {
	device: Device;
	open: boolean;
}

type ExtensionState =
	| "checking"
	| "not-detected"
	| "unpaired"
	| "paired"
	| "pairing"
	| "error";

const formatDate = (value: string | null) =>
	value
		? new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(
				new Date(value),
			)
		: "Never";

export default function BrowserExtensionSetup(props: Props) {
	const installTarget = getExtensionInstallTarget();
	const [state, setState] = createSignal<ExtensionState>("checking");
	const [version, setVersion] = createSignal<string>();
	const [error, setError] = createSignal("");
	const [clientError, setClientError] = createSignal("");
	const [revoking, setRevoking] = createSignal<string>();
	let detectionRun = 0;
	const [clients, { refetch }] = createResource(
		() => (props.open ? props.device.id : undefined),
		getBrowserClients,
	);

	const detect = async () => {
		if (!props.open || state() === "pairing") return;
		const run = ++detectionRun;
		setState("checking");
		setError("");
		const extension = await detectBrowserExtension();
		if (run !== detectionRun || !props.open) return;
		if (!extension) {
			setVersion(undefined);
			setState("not-detected");
			return;
		}
		setVersion(extension.version);
		setState(
			extension.pairedDeviceId === String(props.device.id)
				? "paired"
				: "unpaired",
		);
	};

	createEffect(() => {
		if (!props.open) return;
		untrack(() => void detect());
		const handleFocus = () => void detect();
		const handleVisibility = () => {
			if (document.visibilityState === "visible") void detect();
		};
		window.addEventListener("focus", handleFocus);
		document.addEventListener("visibilitychange", handleVisibility);
		onCleanup(() => {
			detectionRun++;
			window.removeEventListener("focus", handleFocus);
			document.removeEventListener("visibilitychange", handleVisibility);
		});
	});

	const pair = async () => {
		if (state() === "pairing") return;
		setState("pairing");
		setError("");
		try {
			const pairing = await createBrowserPairing(props.device.id);
			const result = await pairBrowserExtension(
				pairing.pairingToken,
				window.location.origin,
			);
			if (!result.success) {
				throw new Error(result.error ?? "Could not connect the extension.");
			}
			setState("paired");
			await refetch();
		} catch (cause) {
			setError(
				cause instanceof Error
					? cause.message
					: "Could not connect the extension.",
			);
			setState("error");
		}
	};

	const revoke = async (clientId: string) => {
		if (revoking()) return;
		setRevoking(clientId);
		setClientError("");
		try {
			await revokeBrowserClient(props.device.id, clientId);
			await refetch();
			void detect();
		} catch {
			setClientError("Could not revoke this browser.");
		} finally {
			setRevoking(undefined);
		}
	};

	return (
		<section class="border-t border-zinc-200 pt-8 flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<p class="font-medium text-zinc-700 text-sm tracking-wider">
					BROWSER EXTENSION
				</p>
				<p class="text-sm text-zinc-500">
					Show a local blkhole page when this browser opens a blocked domain.
				</p>
			</div>

			<Show when={state() === "checking"}>
				<button
					type="button"
					disabled
					class="w-full px-4 py-3 text-sm font-medium tracking-wider bg-zinc-200 text-zinc-500"
				>
					CHECKING EXTENSION…
				</button>
			</Show>

			<Show when={state() === "not-detected"}>
				<Show
					when={installTarget.url}
					fallback={
						<button
							type="button"
							disabled
							class="w-full px-4 py-3 text-sm font-medium tracking-wider bg-zinc-200 text-zinc-500"
						>
							STORE LISTING NOT AVAILABLE
						</button>
					}
				>
					<a
						href={installTarget.url}
						target="_blank"
						rel="noreferrer"
						class="block w-full px-4 py-3 text-center text-sm font-medium tracking-wider bg-black text-white"
					>
						INSTALL FROM {installTarget.name.toUpperCase()}
					</a>
				</Show>
				<div class="bg-zinc-50 p-4 flex flex-col gap-2 text-sm text-zinc-600">
					<p>
						After installation, return to this page and approve access to{" "}
						<span class="font-medium">{window.location.origin}</span>.
					</p>
					<p>
						For a self-hosted instance, open the blkhole extension settings from
						the browser toolbar and set the web app origin to{" "}
						<span class="font-medium">{window.location.origin}</span> before
						checking again.
					</p>
					<Show when={installTarget.browser === "safari"}>
						<p>
							Enable blkhole in Safari settings, then grant website access for
							this site.
						</p>
					</Show>
					<p>
						A disabled extension or missing site permission looks the same as an
						extension that is not installed.
					</p>
				</div>
				<ActionButton onclick={() => void detect()}>CHECK AGAIN</ActionButton>
			</Show>

			<Show when={state() === "unpaired" || state() === "error"}>
				<button
					type="button"
					onclick={() => void pair()}
					class="w-full px-4 py-3 text-sm font-medium tracking-wider bg-black text-white cursor-pointer"
				>
					CONNECT EXTENSION
				</button>
				<Show when={error()}>
					<p class="text-sm text-red-700">{error()}</p>
				</Show>
			</Show>

			<Show when={state() === "pairing"}>
				<button
					type="button"
					disabled
					class="w-full px-4 py-3 text-sm font-medium tracking-wider bg-zinc-200 text-zinc-500"
				>
					CONNECTING…
				</button>
			</Show>

			<Show when={state() === "paired"}>
				<div class="w-full px-4 py-3 text-center text-sm font-medium tracking-wider bg-emerald-100 text-emerald-800">
					CONNECTED{version() ? ` · V${version()}` : ""}
				</div>
			</Show>

			<div class="mt-4 flex flex-col gap-3">
				<p class="font-medium text-zinc-700 text-sm tracking-wider">
					PAIRED BROWSERS
				</p>
				<Show when={clients.loading}>
					<p class="text-sm text-zinc-500">Loading…</p>
				</Show>
				<Show when={clients.error}>
					<p class="text-sm text-red-700">Could not load paired browsers.</p>
				</Show>
				<Show when={clientError()}>
					<p class="text-sm text-red-700">{clientError()}</p>
				</Show>
				<Show
					when={(clients()?.length ?? 0) > 0}
					fallback={
						<Show when={!clients.loading && !clients.error}>
							<p class="text-sm text-zinc-500">No browsers paired yet.</p>
						</Show>
					}
				>
					<div class="divide-y divide-zinc-200 border-y border-zinc-200">
						<For each={clients()}>
							{(client) => (
								<div class="py-3 flex items-start justify-between gap-4">
									<div class="flex flex-col gap-1 text-sm">
										<p class="font-medium text-zinc-700">
											{client.name || client.browser}
										</p>
										<p class="text-zinc-500">
											{client.browser} · Added {formatDate(client.createdAt)} ·
											Last active {formatDate(client.lastActiveAt)}
										</p>
									</div>
									<ActionButton
										onclick={() => void revoke(client.id)}
										class="text-red-700 disabled:text-zinc-400"
									>
										{revoking() === client.id ? "REVOKING…" : "REVOKE"}
									</ActionButton>
								</div>
							)}
						</For>
					</div>
				</Show>
			</div>
		</section>
	);
}
