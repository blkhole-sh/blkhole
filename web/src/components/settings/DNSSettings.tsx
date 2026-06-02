import { createResource, Show } from "solid-js";
import Divider from "../ui/Divider";
import { getSettings } from "~/lib/api";

export default function DNSSettings() {

	const [settings] = createResource(getSettings);
	
	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-display text-2xl">DNS Configuration</h2>
			<div class="flex flex-col gap-1">
				<p class="font-medium text-zinc-700 text-sm tracking-wider">
					UPSTREAM DNS SERVER
				</p>
				<p class="py-2 text-sm leading-snug tracking-wider text-zinc-500">
					<Show when={!settings.loading} fallback="Loading...">
						{settings()?.upstreamDns ?? "—"}
					</Show>
				</p>
				<Divider />
			</div>
		</section>
	);
}
