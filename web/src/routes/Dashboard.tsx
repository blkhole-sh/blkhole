import { createResource } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import StatsPanel from "~/components/dashboard/StatsPanel";
import QueriesPanel from "~/components/dashboard/QueriesPanel";
import SchedulesPanel from "~/components/dashboard/SchedulesPanel";
import DevicesPanel from "~/components/dashboard/DevicesPanel";
import { getQueryStats, getSchedules, getDevices } from "~/lib/api";

export default function Dashboard() {
	const [stats] = createResource(() => getQueryStats("24h"));
	const [schedules] = createResource(getSchedules);
	const [devices] = createResource(getDevices);

	const formatDate = () => {
		return new Date().toLocaleDateString(undefined, {
			weekday: "long",
			year: "numeric",
			month: "long",
			day: "numeric",
		});
	};

	return (
		<PageShell
			title="Dashboard"
			description={`Your DNS universe, observed on ${formatDate()}.`}
		>
			<div class="py-8 flex flex-1 flex-row items-stretch divide-x divide-zinc-200">
				<div class="pr-16 flex flex-col flex-1 divide-y divide-zinc-200">
					<StatsPanel stats={stats()} />
					<QueriesPanel stats={stats()} />
				</div>
				<div class="pl-16 w-xs flex flex-col flex-col divide-y divide-zinc-200">
					<SchedulesPanel schedules={schedules()} />
					<DevicesPanel devices={devices()} />
				</div>
			</div>
		</PageShell>
	);
}
