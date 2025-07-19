import { A } from "@solidjs/router";
import { createResource, For, Show } from "solid-js";
import ScheduleTile from "~/components/schedule/ScheduleTile";
import { getSchedules } from "~/lib/api";

export default function Index() {
	const [schedules] = createResource(getSchedules);

	return (
		<div class="mt-2 flow-root">
			<div class="inline-block min-w-full align-middle">
				<ul role="list" class="divide-y divide-gray-100">
					<For each={schedules()}>
						{(schedule) => <ScheduleTile schedule={schedule} />}
					</For>
				</ul>
			</div>
		</div>
	);
}
