import { For, Show } from "solid-js";
import type { DeviceActivity } from "~/lib/model";

interface Props {
	activity: DeviceActivity[] | undefined;
	onSelect: (deviceId: string) => void;
}

const shades = ["#fafafa", "#e4e4e7", "#a1a1aa", "#52525b", "#000"];

export default function ActivityPanel(props: Props) {
	const peak = () =>
		Math.max(0, ...(props.activity ?? []).flatMap((row) => row.hours));
	const hasActivity = () => peak() > 0;
	const color = (count: number) => {
		if (count === 0 || peak() === 0) return shades[0];
		return shades[Math.min(4, Math.ceil((count / peak()) * 4))];
	};

	return (
		<section class="pt-8 mt-8 border-t border-zinc-200">
			<h2 class="mb-6 font-medium text-zinc-700 text-sm tracking-wider">
				ACTIVITY BY HOUR
			</h2>
			<Show
				when={hasActivity()}
				fallback={
					<div class="h-55 flex items-center justify-center">
						<p class="text-center text-sm text-zinc-400">No queries recorded</p>
					</div>
				}
			>
				<div class="flex flex-col gap-1">
					<For each={props.activity}>
						{(row) => (
							<button
								type="button"
								class="flex flex-row items-center gap-6 text-left cursor-pointer"
								onclick={() => props.onSelect(String(row.deviceId))}
							>
								<span class="w-50 flex-shrink-0 truncate text-sm tracking-wider text-zinc-500">
									{row.deviceName}
								</span>
								<span class="flex flex-1 min-w-0 flex-row gap-0.5">
									<For each={row.hours}>
										{(count) => (
											<span
												class="h-5 flex-1"
												style={{ background: color(count) }}
												title={`${count.toLocaleString()} queries`}
											/>
										)}
									</For>
								</span>
							</button>
						)}
					</For>
					<div class="flex flex-row items-center gap-6">
						<span class="w-50 flex-shrink-0" />
						<div class="pt-2 flex flex-1 justify-between text-xs text-zinc-400">
							<span>00</span>
							<span>06</span>
							<span>12</span>
							<span>18</span>
							<span>24</span>
						</div>
					</div>
				</div>
			</Show>
		</section>
	);
}
