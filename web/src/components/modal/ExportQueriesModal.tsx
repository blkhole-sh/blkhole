import { createEffect, createSignal, For, Show } from "solid-js";
import { exportQueryLogs } from "~/lib/api";
import type { Device } from "~/lib/model";
import ActionButton from "../ui/ActionButton";
import Divider from "../ui/Divider";
import Modal from "./Modal";

type ExportRange = "1h" | "24h" | "7d" | "30d";

interface Props {
	open: boolean;
	devices: Device[] | undefined;
	onClose: () => void;
}

export default function ExportQueriesModal(props: Props) {
	const [selected, setSelected] = createSignal<string[]>([]);
	const [range, setRange] = createSignal<ExportRange>("24h");
	const [error, setError] = createSignal("");

	createEffect(() => {
		if (props.open) {
			setSelected((props.devices ?? []).map((device) => String(device.id)));
			setRange("24h");
			setError("");
		}
	});

	const allSelected = () => selected().length === (props.devices?.length ?? 0);
	const toggle = (id: string) =>
		setSelected((ids) =>
			ids.includes(id) ? ids.filter((item) => item !== id) : [...ids, id],
		);
	const toggleAll = () =>
		setSelected(
			allSelected()
				? []
				: (props.devices ?? []).map((device) => String(device.id)),
		);
	const handleExport = async () => {
		if (selected().length === 0) {
			setError("Select at least one device");
			return;
		}
		try {
			setError("");
			await exportQueryLogs(selected(), range());
			props.onClose();
		} catch {
			setError("Failed to export queries.");
		}
	};

	return (
		<Modal
			title="Export Queries"
			open={props.open}
			onClose={props.onClose}
			onConfirm={handleExport}
			confirmLabel="EXPORT CSV"
		>
			<div class="flex flex-col gap-8">
				<section class="flex flex-col gap-1">
					<div class="flex items-baseline justify-between gap-4">
						<p class="font-medium text-zinc-500 text-sm tracking-wider">
							DEVICES
						</p>
						<ActionButton onclick={toggleAll}>
							{allSelected() ? "CLEAR ALL" : "SELECT ALL"}
						</ActionButton>
					</div>
					<div class="py-2 flex flex-col">
						<For each={props.devices}>
							{(device) => (
								<label class="py-2 flex items-center gap-4 border-b border-zinc-100 cursor-pointer">
									<input
										type="checkbox"
										checked={selected().includes(String(device.id))}
										onChange={() => toggle(String(device.id))}
									/>
									<span class="text-sm tracking-wider">{device.name}</span>
								</label>
							)}
						</For>
					</div>
					<Divider />
				</section>
				<section class="flex flex-col gap-1">
					<label
						for="export-range"
						class="font-medium text-zinc-500 text-sm tracking-wider"
					>
						RANGE
					</label>
					<select
						id="export-range"
						value={range()}
						onChange={(event) =>
							setRange(event.currentTarget.value as ExportRange)
						}
						class="w-full py-2 text-sm font-medium tracking-wider outline-none bg-transparent cursor-pointer"
					>
						<option value="1h">1 HOUR</option>
						<option value="24h">24 HOURS</option>
						<option value="7d">7 DAYS</option>
						<option value="30d">30 DAYS</option>
					</select>
					<Divider />
				</section>
				<Show when={error()}>
					<p class="text-sm text-red-700">{error()}</p>
				</Show>
			</div>
		</Modal>
	);
}
