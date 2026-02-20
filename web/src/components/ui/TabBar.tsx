export interface Tab<T extends string = string> {
	key: T;
	label: string;
}

interface Props<T extends string> {
	tabs: Tab<T>[];
	active: T;
	onChange: (key: T) => void;
}

export default function TabBar<T extends string>(props: Props<T>) {
	return (
		<div class="flex flex-row border-b border-zinc-200">
			{props.tabs.map((tab) => (
				<button
					type="button"
					onclick={() => props.onChange(tab.key)}
					class="pb-2 mr-6 text-sm tracking-wider cursor-pointer"
					classList={{
						"border-b-2 border-black text-black -mb-px":
							props.active === tab.key,
						"text-zinc-400": props.active !== tab.key,
					}}
				>
					{tab.label}
				</button>
			))}
		</div>
	);
}
