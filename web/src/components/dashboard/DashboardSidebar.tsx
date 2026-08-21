import { useNavigate } from "@solidjs/router";
import { For, Show } from "solid-js";
import OSIcon from "~/components/icons/OSIcon";
import type { Device, DeviceActivity, List, Schedule } from "~/lib/model";
import { formatTime, isAllDay } from "~/lib/utils";

interface Props {
	devices: Device[] | undefined;
	activity: DeviceActivity[] | undefined;
	schedules: Schedule[] | undefined;
	lists: List[] | undefined;
	selectedDeviceId: string;
}

export default function DashboardSidebar(props: Props) {
	const navigate = useNavigate();
	const selectedSchedules = () =>
		(props.schedules ?? []).filter(
			(schedule) =>
				schedule.active &&
				(!props.selectedDeviceId ||
					schedule.deviceIds.some(
						(id) => String(id) === props.selectedDeviceId,
					)),
		);
	const selectedLists = () => {
		if (!props.selectedDeviceId) return [];
		const ids = new Set(
			selectedSchedules().flatMap((schedule) => schedule.listIds.map(String)),
		);
		return (props.lists ?? []).filter((list) => ids.has(String(list.id)));
	};
	const notReporting = () =>
		(props.activity ?? []).filter(
			(row) =>
				!row.lastQueryAt ||
				Date.now() - new Date(row.lastQueryAt).getTime() > 60 * 60 * 1000,
		);
	const deviceFor = (id: number) =>
		props.devices?.find((device) => String(device.id) === String(id));
	const lastSeen = (value?: string) => {
		if (!value) return "No queries recorded";
		const hours = Math.max(
			1,
			Math.floor((Date.now() - new Date(value).getTime()) / (60 * 60 * 1000)),
		);
		if (hours < 24)
			return `No queries for ${hours} ${hours === 1 ? "hour" : "hours"}`;
		const days = Math.floor(hours / 24);
		return `No queries for ${days} ${days === 1 ? "day" : "days"}`;
	};

	return (
		<aside class="pl-12 w-72 flex-shrink-0 border-l border-zinc-200">
			<section class="pb-8 border-b border-zinc-200">
				<h2 class="pb-4 font-medium text-zinc-700 text-sm tracking-wider">
					ACTIVE SCHEDULES
				</h2>
				<Show
					when={selectedSchedules().length > 0}
					fallback={
						<p class="text-sm tracking-wider text-zinc-400">
							No active schedules
						</p>
					}
				>
					<div class="flex flex-col gap-3">
						<For each={selectedSchedules()}>
							{(schedule) => (
								<button
									type="button"
									class="flex flex-col text-left cursor-pointer"
									onclick={() => navigate(`/schedules#schedule-${schedule.id}`)}
								>
									<span class="font-medium text-sm tracking-wider">
										{schedule.name}
									</span>
									<span class="text-zinc-500 text-sm tracking-wider">
										{isAllDay(schedule.startTime, schedule.endTime)
											? "All Day"
											: `${formatTime(schedule.startTime)} – ${formatTime(schedule.endTime)}`}
									</span>
								</button>
							)}
						</For>
					</div>
				</Show>
			</section>
			<Show
				when={!props.selectedDeviceId}
				fallback={
					<section class="pt-8">
						<h2 class="pb-4 font-medium text-zinc-700 text-sm tracking-wider">
							BLOCKLISTS IN EFFECT
						</h2>
						<Show
							when={selectedLists().length > 0}
							fallback={
								<p class="text-sm tracking-wider text-zinc-400">
									No active blocklists
								</p>
							}
						>
							<div class="flex flex-col gap-3">
								<For each={selectedLists()}>
									{(list) => (
										<button
											type="button"
											class="flex items-baseline justify-between gap-4 text-left cursor-pointer"
											onclick={() => navigate(`/lists#list-${list.id}`)}
										>
											<span class="min-w-0 truncate font-medium text-sm tracking-wider">
												{list.name}
											</span>
											<span class="flex-shrink-0 text-sm tracking-wider text-zinc-500">
												{list.rules.toLocaleString()} rules
											</span>
										</button>
									)}
								</For>
							</div>
						</Show>
					</section>
				}
			>
				<section class="pt-8">
					<h2 class="pb-4 font-medium text-zinc-700 text-sm tracking-wider">
						NOT REPORTING
					</h2>
					<Show
						when={notReporting().length > 0}
						fallback={
							<p class="text-sm tracking-wider text-zinc-400">
								All devices reporting
							</p>
						}
					>
						<div class="flex flex-col gap-3">
							<For each={notReporting()}>
								{(row) => (
									<button
										type="button"
										class="flex flex-row items-center gap-3 text-left cursor-pointer"
										onclick={() => navigate(`/devices#device-${row.deviceId}`)}
									>
										<OSIcon os={deviceFor(row.deviceId)?.os ?? ""} />
										<span class="min-w-0 flex flex-col">
											<span class="truncate font-medium text-sm tracking-wider">
												{row.deviceName}
											</span>
											<span class="text-zinc-500 text-sm tracking-wider">
												{lastSeen(row.lastQueryAt)}
											</span>
										</span>
									</button>
								)}
							</For>
						</div>
					</Show>
				</section>
			</Show>
		</aside>
	);
}
