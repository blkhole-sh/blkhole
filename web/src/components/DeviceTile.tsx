import { For, Show } from "solid-js";
import { Device } from "~/lib/model";
import ActionButton from "~/components/ActionButton";
import ActionLink from "~/components/ActionLink";
import Tag from "~/components/Tag";

interface Props {
	device: Device;
}

export default function DeviceTile(props: Props) {
	return (
		<div class="py-8 flex flex-col gap-6 font-inter">
			<p class="font-hedvig text-xl">{props.device.name}</p>
			<div class="flex flex-row">
				<div class="flex-1 grid grid-cols-3 items-start text-zinc-500">
					<p class="pb-2 text-sm tracking-wider">STATUS</p>
					<p class="pb-2 text-sm tracking-wider">OS</p>
					<p class="pb-2 text-sm tracking-wider">SCHEDULES</p>
					<div class="flex flex-row gap-2 items-center">
						<div class="w-2 h-2 bg-black rounded-full" />
						<p class="text-black">Connected</p>
					</div>
					<div class="w-fit justify-self-start">
						<Tag>{props.device.os}</Tag>
					</div>
					<div class="flex flex-row flex-wrap gap-2">
						<Show
							when={props.device.scheduleNames.length > 0}
							fallback={<p>—</p>}
						>
							<For each={props.device.scheduleNames}>
								{(name) => <Tag>{name}</Tag>}
							</For>
						</Show>
					</div>
				</div>
				<div class="flex flex-row gap-6 items-center">
					<ActionLink
						href={`/devices/${props.device.id}/config`}
						rel="external"
						download="blkhole.mobileconfig"
					>SETUP</ActionLink>
					<ActionButton>DELETE</ActionButton>
				</div>
			</div>
		</div>
	);
}
