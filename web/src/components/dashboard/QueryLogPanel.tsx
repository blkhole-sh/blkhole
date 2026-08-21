import { For, Show } from "solid-js";
import type { QueryLogPage } from "~/lib/model";
import ActionButton from "../ui/ActionButton";

interface Props {
	logs: QueryLogPage | undefined;
	page: number;
	pageSize: number;
	onPrevious: () => void;
	onNext: () => void;
	onExport: () => void;
}

export default function QueryLogPanel(props: Props) {
	const from = () => (props.logs?.total ? props.page * props.pageSize + 1 : 0);
	const to = () =>
		Math.min((props.page + 1) * props.pageSize, props.logs?.total ?? 0);
	const time = (timestamp: number) =>
		new Date(timestamp * 1000).toLocaleTimeString([], {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
			hour12: false,
		});

	return (
		<section class="pt-8 mt-8 border-t border-zinc-200">
			<div class="mb-6 flex items-baseline justify-between gap-4">
				<h2 class="font-medium text-zinc-700 text-sm tracking-wider">
					LIVE QUERIES
				</h2>
				<Show when={(props.logs?.total ?? 0) > 0}>
					<ActionButton onclick={props.onExport}>EXPORT</ActionButton>
				</Show>
			</div>
			<Show
				when={(props.logs?.items.length ?? 0) > 0}
				fallback={
					<div class="h-55 flex items-center justify-center">
						<p class="text-center text-sm text-zinc-400">No queries recorded</p>
					</div>
				}
			>
				<div class="flex flex-col divide-y divide-zinc-100">
					<For each={props.logs?.items}>
						{(row) => (
							<div class="py-3 flex flex-row items-center gap-6 text-sm tracking-wider text-zinc-500">
								<span class="w-18 flex-shrink-0">{time(row.timestamp)}</span>
								<span class="w-35 flex-shrink-0 truncate">
									{row.deviceName}
								</span>
								<span class="flex-1 min-w-0 truncate">{row.domain}</span>
								<span class="w-24 flex-shrink-0 text-right">
									{row.blocked ? "Blocked" : "Allowed"}
								</span>
							</div>
						)}
					</For>
				</div>
			</Show>
			<div class="pt-4 flex items-baseline justify-between gap-8 text-sm tracking-wider">
				<p class="text-zinc-500">
					{from()}–{to()} of {(props.logs?.total ?? 0).toLocaleString()}
				</p>
				<div class="flex gap-6">
					<ActionButton
						onclick={props.onPrevious}
						disabled={props.page === 0}
						class={props.page === 0 ? "text-zinc-400" : undefined}
					>
						PREVIOUS
					</ActionButton>
					<ActionButton
						onclick={props.onNext}
						disabled={to() >= (props.logs?.total ?? 0)}
						class={
							to() >= (props.logs?.total ?? 0) ? "text-zinc-400" : undefined
						}
					>
						NEXT
					</ActionButton>
				</div>
			</div>
		</section>
	);
}
