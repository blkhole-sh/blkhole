import { createResource, createSignal, onCleanup, onMount } from "solid-js";
import DevicesPanel from "~/components/dashboard/DevicesPanel";
import QueriesPanel from "~/components/dashboard/QueriesPanel";
import SchedulesPanel from "~/components/dashboard/SchedulesPanel";
import StatsPanel from "~/components/dashboard/StatsPanel";
import ChevronDown from "~/components/icons/ChevronDown";
import PageShell from "~/components/layout/PageShell";
import { getDevices, getQueryStats, getSchedules } from "~/lib/api";

type TimeRange = "24h" | "7d" | "30d";

export default function Dashboard() {
	const [tick, setTick] = createSignal(0);
	const [range, setRange] = createSignal<TimeRange>("24h");
	const [stats] = createResource(
		() => ({ tick: tick(), range: range() }),
		(source) => getQueryStats(source.range),
	);
	const [schedules] = createResource(getSchedules);
	const [devices] = createResource(getDevices);

	onMount(() => {
		const id = setInterval(() => setTick((t) => t + 1), 10_000);
		onCleanup(() => clearInterval(id));
	});

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
			actions={
				<div class="flex flex-shrink-0 flex-row items-center gap-2">
					<select
						aria-label="Range"
						value={range()}
						onChange={(event) =>
							setRange(event.currentTarget.value as TimeRange)
						}
						class="text-sm font-medium tracking-wider text-zinc-500 outline-none bg-transparent appearance-none cursor-pointer"
					>
						<option value="24h">24 HOURS</option>
						<option value="7d">7 DAYS</option>
						<option value="30d">30 DAYS</option>
					</select>
					<ChevronDown class="size-4 flex-shrink-0 text-zinc-400 pointer-events-none" />
				</div>
			}
		>
			<div class="flex flex-1 flex-row items-stretch divide-x divide-zinc-200">
				<div class="pr-12 flex flex-col flex-1 divide-y divide-zinc-200">
					<StatsPanel stats={stats()} />
					<QueriesPanel stats={stats()} range={range()} />
				</div>
				<div class="pl-12 w-60 flex flex-col divide-y divide-zinc-200">
					<SchedulesPanel schedules={schedules()} />
					<DevicesPanel devices={devices()} />
				</div>
			</div>
		</PageShell>
	);
}
