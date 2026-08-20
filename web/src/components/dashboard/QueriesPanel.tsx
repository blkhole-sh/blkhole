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
	// Build a map keyed by timestamp so every point carries both series values
	// as concrete numbers. Series that have no point at a given second default
	// to 0 — never undefined, so the chart renderer never produces NaN.
	const data = () => {
		if (!props.stats) return [];

		const byTimestamp = new Map<
			number,
			{ timestamp: number; total: number; blocked: number }
		>();

		for (const p of props.stats.qps ?? []) {
			const ts = new Date(p.timestamp).getTime();
			const row = byTimestamp.get(ts) ?? {
				timestamp: ts,
				total: 0,
				blocked: 0,
			};
			row.total = p.count;
			byTimestamp.set(ts, row);
		}

		for (const p of props.stats.blockedQps ?? []) {
			const ts = new Date(p.timestamp).getTime();
			const row = byTimestamp.get(ts) ?? {
				timestamp: ts,
				total: 0,
				blocked: 0,
			};
			row.blocked = p.count;
			byTimestamp.set(ts, row);
		}

		const format = labelFormats[props.range ?? "24h"] ?? labelFormats["24h"];
		return [...byTimestamp.values()]
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

		const first = d[0];
		if (!first) return [];
		if (count === 1) return [first.xAxis];

		const ticks: string[] = [];
		for (let i = 0; i < count; i++) {
			const idx = Math.floor((i * (d.length - 1)) / (count - 1));
			const point = d[idx];
			if (point) ticks.push(point.xAxis);
		}

		return ticks;
	};

	return (
		<div class="pt-8 flex flex-col flex-1">
			<div class="flex flex-row items-center pb-4">
				<p class="font-medium text-zinc-500 text-sm tracking-wider flex-1">
					QUERIES PER SECOND
				</p>
				<Show when={hasData()}>
					<div class="flex flex-row items-center gap-6">
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
				</Show>
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
