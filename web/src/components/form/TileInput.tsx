import { For, JSX, Show } from "solid-js";

export interface TileOption {
	value: string;
	label: string;
	icon: () => JSX.Element;
}

interface Props {
	label?: string;
	options: TileOption[];
	value: string;
	onChange: (value: string) => void;
	error?: string;
}

export default function TileInput(props: Props) {
	return (
		<div class="flex flex-col gap-1">
			<Show when={props.label}>
				<p class="font-medium text-zinc-700 text-sm tracking-wider">
					{props.label}
				</p>
			</Show>
			<div
				class="grid gap-2 pt-2"
				style={{
					"grid-template-columns": `repeat(${props.options.length}, minmax(0, 1fr))`,
				}}
			>
				<For each={props.options}>
					{(option) => (
						<button
							type="button"
							onclick={() => props.onChange(option.value)}
							class="flex flex-col items-center gap-2 p-3 border cursor-pointer"
							classList={{
								"border-black text-black": props.value === option.value,
								"border-zinc-200 text-zinc-400 hover:border-zinc-400":
									props.value !== option.value,
							}}
						>
							{option.icon()}
							<span class="text-xs tracking-wider">{option.label}</span>
						</button>
					)}
				</For>
			</div>
			<Show when={props.error}>
				<p class="text-xs text-red-700">{props.error}</p>
			</Show>
		</div>
	);
}
