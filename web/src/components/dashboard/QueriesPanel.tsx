import { Axis, AxisLabel, AxisLine, Chart, Line } from "solid-charts";
import { Show } from "solid-js";
import type { QueryStats } from "~/lib/model";

interface Props {
	stats: QueryStats | undefined;
	tickCount?: number; // number of x-axis ticks, default 4
}

export default function QueriesPanel(props: Props) {
	// Map data without rounding
	const data = () => {
		if (!props.stats) return [];

		const blockedMap = new Map(
			props.stats.blocked.map((s) => [s.timestamp, s.count]),
		);

		return props.stats.total.map((stat) => ({
			timestamp: stat.timestamp,
			xAxis: new Date(stat.timestamp).toLocaleTimeString([], {
				hour: "2-digit",
				minute: "2-digit",
			}),
			total: stat.count,
			blocked: blockedMap.get(stat.timestamp) || 0,
		}));
	};

	const hasData = () =>
		data().some((point) => point.total > 0 || point.blocked > 0);

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
					QUERIES OVER TIME
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
