import { useNavigate } from "@solidjs/router";
import { createEffect, createSignal, For, Index, Show } from "solid-js";
import DeleteModal from "~/components/modal/DeleteModal";
import ActionButton from "~/components/ui/ActionButton";
import Switch from "~/components/ui/Switch";
import Tag from "~/components/ui/Tag";
import { deleteSchedule, updateSchedule } from "~/lib/api";
import type { Schedule } from "~/lib/model";
import { formatTime, isAllDay } from "~/lib/utils";

interface Props {
	schedule: Schedule;
	first?: boolean;
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

export default function ScheduleTile(props: Props) {
	const navigate = useNavigate();
	const [deleteOpen, setDeleteOpen] = createSignal(false);

	const [optimisticActive, setOptimisticActive] = createSignal<
		boolean | undefined
	>();

	const active = () => optimisticActive() ?? props.schedule.active;

	createEffect(() => {
		const pa = props.schedule.active;
		if (optimisticActive() !== undefined && pa === optimisticActive()) {
			setOptimisticActive(undefined);
		}
	});

	const handleToggleActive = async () => {
		const next = !active();
		setOptimisticActive(next);
		try {
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
				next,
				days,
				props.schedule.listIds,
				props.schedule.deviceIds,
			);
			props.onUpdated();
		} catch {
			setOptimisticActive(undefined);
		}
	};

	return (
		<div
			id={`schedule-${props.schedule.id}`}
			class="py-5 flex flex-row items-center gap-6"
			classList={{ "opacity-50": !active(), "pt-0": props.first }}
		>
			<Switch checked={active()} onChange={handleToggleActive} />
			<p class="w-28 flex-shrink-0 font-medium tracking-wider truncate">
				{props.schedule.name}
			</p>
			<div class="flex flex-row items-center gap-4 flex-shrink-0">
				<p class="w-24 flex-shrink-0 text-sm text-zinc-500 whitespace-nowrap">
					{isAllDay(props.schedule.startTime, props.schedule.endTime)
						? "All Day"
						: `${formatTime(props.schedule.startTime)} – ${formatTime(props.schedule.endTime)}`}
				</p>
				<div class="w-36 flex-shrink-0 flex flex-row gap-1.5 text-sm tracking-wider text-zinc-300">
					<For each={days}>
						{({ label, key }) => (
							<span
								class="w-3.5 text-center"
								classList={{ "text-black": props.schedule[key] }}
							>
								{label}
							</span>
						)}
					</For>
				</div>
			</div>
			<div class="ml-4 flex-1 min-w-0 flex flex-row flex-wrap justify-start gap-2">
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
				<Show
					when={
						props.schedule.deviceNames.length === 0 &&
						props.schedule.listNames.length === 0
					}
				>
					<p class="text-zinc-500">—</p>
				</Show>
			</div>
			<div class="w-36 flex-shrink-0 flex flex-row justify-end gap-6">
				<ActionButton onclick={() => props.onEdit(props.schedule)}>
					EDIT
				</ActionButton>
				<Show when={!props.schedule.isDefault}>
					<ActionButton onclick={() => setDeleteOpen(true)}>
						DELETE
					</ActionButton>
				</Show>
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
