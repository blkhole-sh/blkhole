import {
	createEffect,
	createResource,
	createSignal,
	For,
	onCleanup,
	onMount,
	Show,
} from "solid-js";
import ActivityPanel from "~/components/dashboard/ActivityPanel";
import DashboardSidebar from "~/components/dashboard/DashboardSidebar";
import DomainsPanel from "~/components/dashboard/DomainsPanel";
import QueriesPanel from "~/components/dashboard/QueriesPanel";
import QueryLogPanel from "~/components/dashboard/QueryLogPanel";
import StatsPanel from "~/components/dashboard/StatsPanel";
import ChevronDown from "~/components/icons/ChevronDown";
import PageShell from "~/components/layout/PageShell";
import ExportQueriesModal from "~/components/modal/ExportQueriesModal";
import {
	getDevices,
	getLists,
	getQueryLogs,
	getQueryStats,
	getSchedules,
} from "~/lib/api";

type TimeRange = "24h" | "7d" | "30d";
const LOG_PAGE_SIZE = 8;

export default function Dashboard() {
	const [tick, setTick] = createSignal(0);
	const [range, setRange] = createSignal<TimeRange>("24h");
	const [deviceId, setDeviceId] = createSignal("");
	const [logPage, setLogPage] = createSignal(0);
	const [exportOpen, setExportOpen] = createSignal(false);
	const [stats] = createResource(
		() => ({ tick: tick(), range: range(), deviceId: deviceId() }),
		(source) => getQueryStats(source.range, source.deviceId || undefined),
	);
	const [schedules] = createResource(getSchedules);
	const [devices] = createResource(getDevices);
	const [lists] = createResource(getLists);
	const [logs] = createResource(
		() => ({
			tick: tick(),
			range: range(),
			deviceId: deviceId(),
			page: logPage(),
		}),
		(source) =>
			getQueryLogs({
				range: source.range,
				...(source.deviceId ? { deviceId: source.deviceId } : {}),
				limit: LOG_PAGE_SIZE,
				offset: source.page * LOG_PAGE_SIZE,
			}),
	);

	createEffect(() => {
		range();
		deviceId();
		setLogPage(0);
	});

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
	const description = () =>
		deviceId()
			? `One device in focus, last ${range() === "24h" ? "24 hours" : range() === "7d" ? "7 days" : "30 days"}.`
			: `Your DNS universe, observed on ${formatDate()}.`;

	return (
		<div class="h-screen flex flex-col overflow-hidden">
			<PageShell
				title="Dashboard"
				description={description()}
				contentClass="p-0"
				actions={
					<div class="flex flex-shrink-0 flex-row items-center gap-8">
						<div class="relative flex flex-shrink-0 items-center gap-1 text-sm font-medium tracking-wider text-zinc-700 hover:text-black">
							<span aria-hidden="true" class="whitespace-nowrap">
								{deviceId()
									? devices()
											?.find((device) => String(device.id) === deviceId())
											?.name.toUpperCase()
									: "ALL DEVICES"}
							</span>
							<ChevronDown class="size-4 flex-shrink-0 text-zinc-400 pointer-events-none" />
							<select
								aria-label="Device"
								value={deviceId()}
								onChange={(event) => setDeviceId(event.currentTarget.value)}
								class="absolute inset-0 size-full opacity-0 cursor-pointer"
							>
								<option value="">ALL DEVICES</option>
								<For each={devices()}>
									{(device) => (
										<option value={String(device.id)}>
											{device.name.toUpperCase()}
										</option>
									)}
								</For>
							</select>
						</div>
						<div class="relative flex flex-shrink-0 items-center gap-1 text-sm font-medium tracking-wider text-zinc-700 hover:text-black">
							<span aria-hidden="true" class="whitespace-nowrap">
								{range() === "24h"
									? "24 HOURS"
									: range() === "7d"
										? "7 DAYS"
										: "30 DAYS"}
							</span>
							<ChevronDown class="size-4 flex-shrink-0 text-zinc-400 pointer-events-none" />
							<select
								aria-label="Range"
								value={range()}
								onChange={(event) =>
									setRange(event.currentTarget.value as TimeRange)
								}
								class="absolute inset-0 size-full opacity-0 cursor-pointer"
							>
								<option value="24h">24 HOURS</option>
								<option value="7d">7 DAYS</option>
								<option value="30d">30 DAYS</option>
							</select>
						</div>
					</div>
				}
			>
				<div class="flex flex-1 flex-row items-stretch min-h-0 overflow-hidden">
					<div class="p-12 min-h-0 flex flex-col flex-1 min-w-0 overflow-y-auto">
						<div class="pb-8 border-b border-zinc-200">
							<StatsPanel stats={stats()} />
						</div>
						<div class="h-96 flex-shrink-0">
							<QueriesPanel stats={stats()} range={range()} />
						</div>
						<Show when={!deviceId()}>
							<ActivityPanel
								activity={stats()?.activity}
								onSelect={setDeviceId}
							/>
						</Show>
						<DomainsPanel
							domains={stats()?.domains}
							showVerdict={!!deviceId()}
						/>
						<QueryLogPanel
							logs={logs()}
							page={logPage()}
							pageSize={LOG_PAGE_SIZE}
							onPrevious={() => setLogPage((page) => Math.max(0, page - 1))}
							onNext={() => setLogPage((page) => page + 1)}
							onExport={() => setExportOpen(true)}
						/>
					</div>
					<div class="py-12 pr-12 flex flex-shrink-0">
						<DashboardSidebar
							devices={devices()}
							activity={stats()?.activity}
							schedules={schedules()}
							lists={lists()}
							selectedDeviceId={deviceId()}
						/>
					</div>
				</div>
				<ExportQueriesModal
					open={exportOpen()}
					devices={devices()}
					onClose={() => setExportOpen(false)}
				/>
			</PageShell>
		</div>
	);
}
