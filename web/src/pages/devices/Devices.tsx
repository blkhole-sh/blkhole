import { createResource, For } from "solid-js";
import DeviceTile from "~/components/device/DeviceTile";
import { getDevices } from "~/lib/api";

export default function Index() {
	const [devices] = createResource(getDevices);

	return (
		<div class="mt-2 flow-root">
			<div class="inline-block min-w-full align-middle">
				<ul role="list" class="divide-y divide-gray-100">
					<For each={devices()}>
						{(device) => <DeviceTile device={device} />}
					</For>
				</ul>
			</div>
		</div>
	);
}
