import { useNavigate } from "@solidjs/router";
import { createSignal, Index, Show } from "solid-js";
import OSIcon from "~/components/icons/OSIcon";
import ActionButton from "~/components/ui/ActionButton";
import Tag from "~/components/ui/Tag";
import { deleteDevice } from "~/lib/api";
import { useScrollToHash } from "~/lib/hooks";
import type { Device } from "~/lib/model";
import DeleteModal from "../modal/DeleteModal";
import DeviceSetupModal from "../modal/DeviceSetupModal";

interface Props {
	device: Device;
	onEdit: (device: Device) => void;
	onDeleted: () => void;
}

export default function DeviceTile(props: Props) {
	const navigate = useNavigate();
	useScrollToHash();
	const [setupDeviceOpen, setSetupDeviceOpen] = createSignal(false);
	const [deleteOpen, setDeleteOpen] = createSignal(false);

	return (
		<div
			id={`device-${props.device.id}`}
			class="py-5 flex flex-row items-center gap-8"
		>
			<div class="flex-shrink-0 flex flex-row items-center gap-4 min-h-6.5">
				<OSIcon os={props.device.os} />
				<p class="w-44 flex-shrink-0 font-medium tracking-wider truncate">
					{props.device.name}
				</p>
			</div>
			<p class="w-24 flex-shrink-0 text-sm text-zinc-500">Connected</p>
			<div class="flex-1 min-w-0 flex flex-row flex-wrap justify-center gap-2">
				<Show
					when={props.device.scheduleNames.length > 0}
					fallback={<p class="text-zinc-500">—</p>}
				>
					<Index each={props.device.scheduleNames}>
						{(name, i) => (
							<Tag
								onclick={() =>
									navigate(`/schedules#schedule-${props.device.scheduleIds[i]}`)
								}
							>
								{name()}
							</Tag>
						)}
					</Index>
				</Show>
			</div>
			<div class="w-48 flex-shrink-0 flex flex-row justify-end gap-6">
				<ActionButton onclick={() => setSetupDeviceOpen(true)}>
					SETUP
				</ActionButton>
				<ActionButton onclick={() => props.onEdit(props.device)}>
					EDIT
				</ActionButton>
				<ActionButton onclick={() => setDeleteOpen(true)}>DELETE</ActionButton>
			</div>
			<DeviceSetupModal
				open={setupDeviceOpen()}
				device={props.device}
				onClose={() => setSetupDeviceOpen(false)}
			/>
			<DeleteModal
				open={deleteOpen()}
				name={props.device.name}
				onClose={() => setDeleteOpen(false)}
				onConfirm={async () => {
					await deleteDevice(props.device.id);
					setDeleteOpen(false);
					props.onDeleted();
				}}
			/>
		</div>
	);
}
