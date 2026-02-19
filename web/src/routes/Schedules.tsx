import { createResource, For } from "solid-js";
import ScheduleTile from "~/components/ScheduleTile";
import PageShell from "~/components/PageShell";
import { getSchedules } from "~/lib/api";

export default function Schedules() {
	const [schedules] = createResource(getSchedules);

	return (
		<PageShell
			title="Schedules"
			description="Time-based rules that apply blocklists to specific devices automatically."
			cta="ADD SCHEDULE"
		>
			<div class="divide-y divide-zinc-200">
				<For each={schedules()}>
					{(schedule) => <ScheduleTile schedule={schedule} />}
				</For>
			</div>
		</PageShell>
	);
}
