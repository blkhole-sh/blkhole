import { createResource, For } from "solid-js";
import DeviceTile from "~/components/DeviceTile";
import PageShell from "~/components/PageShell";
import { getDevices } from "~/lib/api";

export default function Devices() {
	const [devices] = createResource(getDevices);

	return (
		<PageShell
			title="Devices"
			description="Connected devices routed through your blkhole server."
			cta="ADD DEVICE"
		>
			<div class="divide-y divide-zinc-200">
				<For each={devices()}>{(device) => <DeviceTile device={device} />}</For>
			</div>
		</PageShell>
	);
}
