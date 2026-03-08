import { Show, createSignal, Index } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { Device } from "~/lib/model";
import ActionButton from "~/components/ui/ActionButton";
import Tag from "~/components/ui/Tag";
import OSIcon from "~/components/icons/OSIcon";
import { deleteDevice } from "~/lib/api";
import DeleteModal from "../modal/DeleteModal";
import DeviceSetupModal from "../modal/DeviceSetupModal";
import { useScrollToHash } from "~/lib/hooks";

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
		<div id={`device-${props.device.id}`} class="py-8 flex flex-col gap-4">
			<p class="font-medium tracking-wider">{props.device.name}</p>
			<div class="flex flex-row">
				<div class="flex-1 grid grid-cols-3 items-start text-zinc-500">
					<p class="pb-2 text-sm tracking-wider">STATUS</p>
					<p class="pb-2 text-sm tracking-wider">OS</p>
					<p class="pb-2 text-sm tracking-wider">SCHEDULES</p>
					<p>Connected</p>
					<OSIcon os={props.device.os} />
					<div class="flex flex-row flex-wrap gap-2">
						<Show
							when={props.device.scheduleNames.length > 0}
							fallback={<p>—</p>}
						>
							<Index each={props.device.scheduleNames}>
								{(name, i) => (
									<Tag
										onclick={() =>
											navigate(
												`/schedules#schedule-${props.device.scheduleIds[i]}`,
											)
										}
									>
										{name()}
									</Tag>
								)}
							</Index>
						</Show>
					</div>
				</div>
				<div class="flex flex-row gap-6 items-center">
					<ActionButton onclick={() => setSetupDeviceOpen(true)}>
						SETUP
					</ActionButton>
					<ActionButton onclick={() => props.onEdit(props.device)}>
						EDIT
					</ActionButton>
					<ActionButton onclick={() => setDeleteOpen(true)}>
						DELETE
					</ActionButton>
				</div>
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
