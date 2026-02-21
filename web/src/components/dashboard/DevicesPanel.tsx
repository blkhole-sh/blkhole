import { For, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import type { Device } from "~/lib/model";
import OSIcon from "~/components/icons/OSIcon";

interface Props {
	devices: Device[] | undefined;
}

export default function DevicesPanel(props: Props) {
	const navigate = useNavigate();

	return (
		<div class="pt-8">
			<p class="pb-4 font-medium text-zinc-500 text-sm tracking-wider">
				DEVICES IN USE
			</p>
			<div class="flex flex-col gap-3">
				<Show
					when={props.devices && props.devices.length > 0}
					fallback={<p class="text-zinc-400 text-sm">No devices configured</p>}
				>
					<For each={props.devices}>
						{(device) => (
							<div
								class="flex flex-row gap-3 items-center cursor-pointer"
								onclick={() => navigate(`/devices#device-${device.id}`)}
							>
								<OSIcon os={device.os} />
								<div class="flex flex-col gap-1">
									<p class="font-display text">{device.name}</p>
									<p class="text-zinc-500 text-sm">Connected</p>
								</div>
							</div>
						)}
					</For>
				</Show>
			</div>
		</div>
	);
}
