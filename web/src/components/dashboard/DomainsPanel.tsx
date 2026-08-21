import { For, Show } from "solid-js";
import type { DomainStat } from "~/lib/model";

interface Props {
	domains: DomainStat[] | undefined;
	showVerdict: boolean;
}

export default function DomainsPanel(props: Props) {
	const peak = () =>
		Math.max(0, ...(props.domains ?? []).map((row) => row.count));

	return (
		<section class="pt-8 mt-8 border-t border-zinc-200">
			<h2 class="mb-6 font-medium text-zinc-700 text-sm tracking-wider">
				{props.showVerdict ? "TOP REQUESTED DOMAINS" : "TOP BLOCKED DOMAINS"}
			</h2>
			<Show
				when={(props.domains?.length ?? 0) > 0}
				fallback={
					<div class="h-55 flex items-center justify-center">
						<p class="text-center text-sm text-zinc-400">No queries recorded</p>
					</div>
				}
			>
				<div class="flex flex-col divide-y divide-zinc-100">
					<For each={props.domains}>
						{(row) => (
							<div class="py-3 flex flex-row items-center gap-6">
								<span class="w-50 flex-shrink-0 truncate text-sm tracking-wider text-zinc-500">
									{row.domain}
								</span>
								<div class="h-0.5 flex-1 bg-zinc-100">
									<div
										class="h-full bg-black"
										style={{
											width: `${peak() === 0 ? 0 : (row.count / peak()) * 100}%`,
										}}
									/>
								</div>
								<span class="w-16 flex-shrink-0 text-right text-sm tracking-wider text-zinc-500">
									{row.count.toLocaleString()}
								</span>
								<Show when={props.showVerdict}>
									<span class="w-24 flex-shrink-0 text-right text-sm tracking-wider text-zinc-500">
										{row.blocked > 0 ? "Blocked" : "Allowed"}
									</span>
								</Show>
							</div>
						)}
					</For>
				</div>
			</Show>
		</section>
	);
}
