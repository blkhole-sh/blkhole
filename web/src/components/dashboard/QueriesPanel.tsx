import { Axis, AxisLabel, AxisLine, Chart, Line } from "solid-charts";
import { Show } from "solid-js";
import type { QueryStats } from "~/lib/model";

interface Props {
	stats: QueryStats | undefined;
	range?: "24h" | "7d" | "30d";
	tickCount?: number; // number of x-axis ticks, default 4
}

// X-axis label formats per range: time of day for 24h, weekday plus time
// for 7d, calendar date for 30d.
const labelFormats: Record<string, Intl.DateTimeFormatOptions> = {
	"24h": { hour: "2-digit", minute: "2-digit" },
	"7d": { weekday: "short", hour: "2-digit", minute: "2-digit" },
	"30d": { month: "short", day: "numeric" },
};

export default function QueriesPanel(props: Props) {
	// Both series have identical window counts. For each window we emit two
	// rows — one at the total peak timestamp and one at the blocked peak
	// timestamp — so every row carries both values and the renderer never sees
	// undefined (which would produce NaN in the SVG path).
	const data = () => {
		if (!props.stats) return [];

		const qps = props.stats.qps ?? [];
		const blockedQps = props.stats.blockedQps ?? [];
		const n = Math.min(qps.length, blockedQps.length);

		const rows: Array<{
			timestamp: number;
			total: number;
			blocked: number;
		}> = [];

		for (let i = 0; i < n; i++) {
			const ts1 = new Date(qps[i].timestamp).getTime();
			const ts2 = new Date(blockedQps[i].timestamp).getTime();
			rows.push({
				timestamp: ts1,
				total: qps[i].count,
				blocked: blockedQps[i].count,
			});
			if (ts2 !== ts1) {
				rows.push({
					timestamp: ts2,
					total: qps[i].count,
					blocked: blockedQps[i].count,
				});
			}
		}

		const format = labelFormats[props.range ?? "24h"] ?? labelFormats["24h"];
		return rows
			.sort((a, b) => a.timestamp - b.timestamp)
			.map((row) => ({
				...row,
				xAxis: new Date(row.timestamp).toLocaleString([], format),
			}));
	};

	const hasData = () =>
		data().some((point) => (point.total ?? 0) > 0 || (point.blocked ?? 0) > 0);

	// Generate parameterized x-axis ticks without rounding
	const getXAxisTicks = (count: number = 4) => {
		const d = data();
		if (d.length === 0 || count <= 0) return [];

		if (count === 1) return [d[0]?.xAxis];

		const ticks: string[] = [];
		for (let i = 0; i < count; i++) {
			const idx = Math.floor((i * (d.length - 1)) / (count - 1));
			ticks.push(d[idx]?.xAxis);
		}

		return ticks;
	};

	return (
		<div class="pt-8 flex flex-col flex-1">
			<div class="flex flex-row items-center pb-4">
				<p class="font-medium text-zinc-500 text-sm tracking-wider flex-1">
					QUERIES PER SECOND
				</p>
				<div class="flex flex-row items-center gap-5">
					<div class="flex flex-row items-center gap-2">
						<svg width="20" height="10" aria-hidden="true">
							<line
								x1="0"
								y1="5"
								x2="20"
								y2="5"
								stroke="black"
								stroke-width="2"
							/>
						</svg>
						<span class="text-zinc-400 text-xs tracking-wider">TOTAL</span>
					</div>
					<div class="flex flex-row items-center gap-2">
						<svg width="20" height="10" aria-hidden="true">
							<line
								x1="0"
								y1="5"
								x2="20"
								y2="5"
								stroke="#a1a1aa"
								stroke-width="2"
								stroke-dasharray="4 3"
							/>
						</svg>
						<span class="text-zinc-400 text-xs tracking-wider">BLOCKED</span>
					</div>
				</div>
			</div>
			<div class="flex-1">
				<Show
					when={hasData()}
					fallback={
						<div class="flex items-center justify-center h-full">
							<p class="text-zinc-400 text-sm">No queries recorded</p>
						</div>
					}
				>
					<Chart data={data()}>
						<Axis axis="y" position="left">
							<AxisLabel class="text-zinc-400 text-xs" />
						</Axis>
						<Axis
							axis="x"
							position="bottom"
							dataKey="xAxis"
							tickValues={getXAxisTicks(props.tickCount ?? 4)}
						>
							<AxisLabel class="text-zinc-400 text-xs" />
							<AxisLine class="stroke-zinc-200" />
						</Axis>
						<Line dataKey="total" stroke-width={2} class="stroke-black" />
						<Line
							dataKey="blocked"
							stroke-width={2}
							class="stroke-zinc-400"
							stroke-dasharray="4 3"
						/>
					</Chart>
				</Show>
			</div>
		</div>
	);
}
