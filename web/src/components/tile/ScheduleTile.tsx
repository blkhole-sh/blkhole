import { For, Show, createSignal, Index } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { Schedule } from "~/lib/model";
import ActionButton from "~/components/ui/ActionButton";
import Tag from "~/components/ui/Tag";
import Switch from "~/components/ui/Switch";
import { deleteSchedule, updateSchedule } from "~/lib/api";
import DeleteModal from "~/components/modal/DeleteModal";
import { useScrollToHash } from "~/lib/hooks";

interface Props {
	schedule: Schedule;
	onEdit: (schedule: Schedule) => void;
	onDeleted: () => void;
	onUpdated: () => void;
}

const days = [
	{ label: "M", key: "monday" },
	{ label: "T", key: "tuesday" },
	{ label: "W", key: "wednesday" },
	{ label: "T", key: "thursday" },
	{ label: "F", key: "friday" },
	{ label: "S", key: "saturday" },
	{ label: "S", key: "sunday" },
] as const;

const formatTime = (t: string) => {
	const [hours, minutes] = t.split(":");
	const d = new Date();
	d.setHours(Number(hours), Number(minutes), 0, 0);
	return d.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
};

export default function ScheduleTile(props: Props) {
	const navigate = useNavigate();
	const [deleteOpen, setDeleteOpen] = createSignal(false);

	const handleToggleActive = async () => {
		const days = {
			monday: props.schedule.monday,
			tuesday: props.schedule.tuesday,
			wednesday: props.schedule.wednesday,
			thursday: props.schedule.thursday,
			friday: props.schedule.friday,
			saturday: props.schedule.saturday,
			sunday: props.schedule.sunday,
		};
		await updateSchedule(
			props.schedule.id,
			props.schedule.name,
			props.schedule.startTime,
			props.schedule.endTime,
			!props.schedule.active,
			days,
			props.schedule.listIds,
			props.schedule.deviceIds,
		);
		props.onUpdated();
	};

	return (
		<div id={`schedule-${props.schedule.id}`} class="py-8 flex flex-col gap-6">
			<div class="flex flex-col gap-1">
				<div class="flex flex-row justify-between items-center">
					<p
						class="font-medium tracking-wider"
						classList={{ "opacity-50": !props.schedule.active }}
					>
						{props.schedule.name}
					</p>
					<Show when={!props.schedule.isDefault}>
						<Switch
							checked={props.schedule.active}
							onChange={handleToggleActive}
						/>
					</Show>
				</div>
				<p
					class="text-zinc-500 text-sm"
					classList={{ "opacity-50": !props.schedule.active }}
				>
					{props.schedule.active ? "Active" : "Inactive"}
				</p>
			</div>
			<div class="flex flex-row">
				<div
					class="flex-1 grid grid-cols-3 items-start text-zinc-500"
					classList={{ "opacity-50": !props.schedule.active }}
				>
					<p class="pb-2 text-sm tracking-wider">TIMING</p>
					<p class="pb-2 text-sm tracking-wider">DEVICES</p>
					<p class="pb-2 text-sm tracking-wider">BLOCKLISTS</p>
					<div class="max-w-md">
						<p class="pb-4 text-black">
							{props.schedule.startTime === props.schedule.endTime
								? "All Day"
								: `${formatTime(props.schedule.startTime)} – ${formatTime(props.schedule.endTime)}`}
						</p>
						<div class="flex flex-row gap-3 text-zinc-300">
							<For each={days}>
								{({ label, key }) => (
									<span classList={{ "text-black": props.schedule[key] }}>
										{label}
									</span>
								)}
							</For>
						</div>
					</div>
					<div class="flex flex-row flex-wrap gap-2 max-w-sm">
						<Show
							when={props.schedule.deviceNames.length > 0}
							fallback={<p>-</p>}
						>
							<Index each={props.schedule.deviceNames}>
								{(name, i) => (
									<Tag
										onclick={() =>
											navigate(`/devices#device-${props.schedule.deviceIds[i]}`)
										}
									>
										{name()}
									</Tag>
								)}
							</Index>
						</Show>
					</div>
					<div class="flex flex-row flex-wrap gap-2 max-w-sm">
						<Show
							when={props.schedule.listNames.length > 0}
							fallback={<p>-</p>}
						>
							<Index each={props.schedule.listNames}>
								{(name, i) => (
									<Tag
										onclick={() =>
											navigate(`/lists#list-${props.schedule.listIds[i]}`)
										}
									>
										{name()}
									</Tag>
								)}
							</Index>
						</Show>
					</div>
				</div>
				<div class="flex flex-row gap-6">
					<ActionButton onclick={() => props.onEdit(props.schedule)}>
						EDIT
					</ActionButton>
					<Show when={!props.schedule.isDefault}>
						<ActionButton onclick={() => setDeleteOpen(true)}>
							DELETE
						</ActionButton>
					</Show>
				</div>
			</div>
			<DeleteModal
				open={deleteOpen()}
				name={props.schedule.name}
				onClose={() => setDeleteOpen(false)}
				onConfirm={async () => {
					await deleteSchedule(props.schedule.id);
					setDeleteOpen(false);
					props.onDeleted();
				}}
			/>
		</div>
	);
}
