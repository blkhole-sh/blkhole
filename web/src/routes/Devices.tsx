import { createResource, createSignal, For, Show } from "solid-js";
import DeviceTile from "~/components/tile/DeviceTile";
import PageShell from "~/components/layout/PageShell";
import DeviceModal from "~/components/modal/DeviceModal";
import EmptyState from "~/components/ui/EmptyState";
import { getDevices } from "~/lib/api";
import { Device } from "~/lib/model";
import { useScrollToHash } from "~/lib/hooks";

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
		<PageShell
			title="Devices"
			description="Devices caught in your gravitational field."
			cta="ADD DEVICE"
			onCTA={handleAdd}
		>
			<Show
				when={devices() && devices()!.length > 0}
				fallback={
					<Show when={!devices.loading}>
						<EmptyState
							message="blkhole has nothing in orbit"
							subtitle="Add your first device to start blocking"
						/>
					</Show>
				}
			>
				<div class="divide-y divide-zinc-200">
					<For each={devices()}>
						{(device) => (
							<DeviceTile
								device={device}
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
	);
}
