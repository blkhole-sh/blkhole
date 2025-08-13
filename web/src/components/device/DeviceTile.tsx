import { A } from "@solidjs/router";
import { Device } from "~/lib/model";
import Dot from "../icons/Dot";
import Edit from "../icons/Edit";
import Trash from "../icons/Trash";

interface Props {
	device: Device;
}

export default function DeviceTile(props: Props) {
	return (
		<li class="flex flex-wrap items-center justify-between gap-x-6 gap-y-4 py-2 sm:flex-nowrap">
			<div class="flex-1">
				<a
					href={`/devices/${props.device.id}/config`}
					class="text-sm/6 font-semibold text-gray-900"
					rel="external"
					download="leo.mobileconfig"
				>
					{props.device.name}
				</a>
				<div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
					<p>{props.device.os}</p>
					<Dot />
					<p>192.168.1.1</p>
				</div>
			</div>
			<div class="flex gap-x-4">
				<A href={`/devices/${props.device.id}`}>
					<Edit />
				</A>
				<Trash />
			</div>
		</li>
	);
}
