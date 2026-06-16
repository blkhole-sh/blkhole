import { For, Show } from "solid-js";

const DAYS = [
	{ key: "monday", label: "M" },
	{ key: "tuesday", label: "T" },
	{ key: "wednesday", label: "W" },
	{ key: "thursday", label: "T" },
	{ key: "friday", label: "F" },
	{ key: "saturday", label: "S" },
	{ key: "sunday", label: "S" },
] as const;

export type DayKey = (typeof DAYS)[number]["key"];
export type Days = Record<DayKey, boolean>;

interface Props {
	label: string;
	value: Days;
	onChange: (days: Days) => void;
	error?: string;
}

export default function DaysInput(props: Props) {
	const toggle = (key: DayKey) => {
		props.onChange({ ...props.value, [key]: !props.value[key] });
	};

	return (
		<div class="flex flex-col gap-2">
			<p class="font-medium text-zinc-700 text-sm tracking-wider">
				{props.label}
			</p>
			<div class="flex flex-row gap-4">
				<For each={DAYS}>
					{(day) => (
						<button
							type="button"
							onclick={() => toggle(day.key)}
							class="text-sm tracking-wider cursor-pointer"
							classList={{
								"text-black font-medium": props.value[day.key],
								"text-zinc-400": !props.value[day.key],
							}}
						>
							{day.label}
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
