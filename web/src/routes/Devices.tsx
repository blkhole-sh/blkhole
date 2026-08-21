import { createResource, createSignal, For, Show } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import DeviceModal from "~/components/modal/DeviceModal";
import DeviceTile from "~/components/tile/DeviceTile";
import EmptyState from "~/components/ui/EmptyState";
import { getDevices } from "~/lib/api";
import { useScrollToHash } from "~/lib/hooks";
import type { Device } from "~/lib/model";

export default function Devices() {
	useScrollToHash();
	const [modalOpen, setModalOpen] = createSignal(false);
	const [editingDevice, setEditingDevice] = createSignal<Device | null>(null);
	const [devices, { refetch }] = createResource(getDevices);

	const handleAdd = () => {
		setEditingDevice(null);
		setModalOpen(true);
	};

	const handleEdit = (device: Device) => {
		setEditingDevice(device);
		setModalOpen(true);
	};

	const handleSaved = () => {
		setModalOpen(false);
		setEditingDevice(null);
		refetch();
	};

	return (
		<div class="h-screen flex flex-col overflow-hidden">
			<PageShell
				title="Devices"
				description="Devices caught in your gravitational field."
				cta="ADD DEVICE"
				onCTA={handleAdd}
				contentClass="p-12 overflow-y-auto"
			>
				<Show
					when={(devices()?.length ?? 0) > 0}
					fallback={
						<Show when={!devices.loading}>
							<EmptyState
								message="blkhole has nothing in orbit"
								subtitle="Add your first device to start blocking"
							/>
						</Show>
					}
				>
					<div class="grid grid-cols-[max-content_6rem_minmax(0,1fr)_12rem] gap-x-8 divide-y divide-zinc-200">
						<For each={devices()}>
							{(device, index) => (
								<DeviceTile
									device={device}
									first={index() === 0}
									onEdit={handleEdit}
									onDeleted={refetch}
								/>
							)}
						</For>
					</div>
				</Show>
				<DeviceModal
					open={modalOpen()}
					device={editingDevice()}
					onClose={() => setModalOpen(false)}
					onSaved={handleSaved}
				/>
			</PageShell>
		</div>
	);
}
