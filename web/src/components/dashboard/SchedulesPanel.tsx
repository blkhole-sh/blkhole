import { For, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import type { Schedule } from "~/lib/model";

interface Props {
	schedules: Schedule[] | undefined;
}

const formatTime = (t: string) => {
	const [hours, minutes] = t.split(":");
	const d = new Date();
	d.setHours(Number(hours), Number(minutes), 0, 0);
	return d.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
};

export default function SchedulesPanel(props: Props) {
	const navigate = useNavigate();
	const activeSchedules = () => props.schedules?.filter((s) => s.active) || [];

	return (
		<div class="pb-8">
			<p class="pb-4 font-medium text-zinc-500 text-sm tracking-wider">
				ACTIVE SCHEDULES
			</p>
			<div class="flex flex-col gap-3">
				<Show
					when={activeSchedules().length > 0}
					fallback={<p class="text-zinc-400 text-sm">No active schedules</p>}
				>
					<For each={activeSchedules()}>
						{(schedule) => (
							<div
								class="flex flex-col flex-1 cursor-pointer"
								onclick={() => navigate(`/schedules#schedule-${schedule.id}`)}
							>
								<p class="font font-medium text-sm tracking-wider">
									{schedule.name}
								</p>
								<p class="text-zinc-500 text-sm tracking-wider">
									{formatTime(schedule.startTime)} –{" "}
									{formatTime(schedule.endTime)}
								</p>
							</div>
						)}
					</For>
				</Show>
			</div>
		</div>
	);
}
